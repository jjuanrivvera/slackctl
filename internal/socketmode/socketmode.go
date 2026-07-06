// Package socketmode is a hand-written Slack Socket Mode client (no Slack SDK): it fetches
// a fresh wss URL from apps.connections.open (app-level xapp token), reads envelopes,
// acknowledges each within Slack's 3-second window BEFORE handing it to the caller, and
// reconnects on disconnect frames and socket errors. See DECISIONS.md ("listen").
package socketmode

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"time"

	"github.com/coder/websocket"

	"github.com/jjuanrivvera/slackctl/internal/slackevent"
)

// Envelope is one Socket Mode frame. hello/disconnect are connection-lifecycle frames;
// events_api/slash_commands/interactive carry work items that must be acked.
type Envelope struct {
	EnvelopeID             string          `json:"envelope_id,omitempty"`
	Type                   string          `json:"type"`
	Reason                 string          `json:"reason,omitempty"` // disconnect frames
	Payload                json.RawMessage `json:"payload,omitempty"`
	AcceptsResponsePayload bool            `json:"accepts_response_payload,omitempty"`
	RetryAttempt           int             `json:"retry_attempt,omitempty"`
	RetryReason            string          `json:"retry_reason,omitempty"`
	NumConnections         int             `json:"num_connections,omitempty"` // hello frames
}

// conn is the slice of *websocket.Conn the client uses — injectable for tests.
type conn interface {
	Read(ctx context.Context) (websocket.MessageType, []byte, error)
	Write(ctx context.Context, typ websocket.MessageType, p []byte) error
	Close(code websocket.StatusCode, reason string) error
}

// Client streams Socket Mode envelopes.
type Client struct {
	// OpenURL fetches a fresh wss:// URL (apps.connections.open). Slack's URLs are
	// single-use tickets, so every (re)connection asks for a new one.
	OpenURL func(ctx context.Context) (string, error)
	// Dial opens the websocket. Defaults to coder/websocket; injectable for tests.
	Dial func(ctx context.Context, url string) (conn, error)
	// Log receives connection-lifecycle notes (connected, reconnecting). Never event data.
	Log io.Writer
	// DebugReconnects appends debug_reconnects=true so Slack rotates the connection every
	// ~360s — for exercising the reconnect path.
	DebugReconnects bool

	rng func() float64
}

// New builds a Client over an OpenURL func.
func New(openURL func(ctx context.Context) (string, error), logW io.Writer) *Client {
	return &Client{
		OpenURL: openURL,
		Dial: func(ctx context.Context, url string) (conn, error) {
			c, _, err := websocket.Dial(ctx, url, nil) //nolint:bodyclose // library owns the handshake response
			if err != nil {
				return nil, err
			}
			// Slack events (message payloads with blocks) can exceed the 32KiB default.
			c.SetReadLimit(1 << 22)
			return c, nil
		},
		Log: logW,
		rng: rand.Float64, //nolint:gosec // G404: reconnect jitter is not a security boundary
	}
}

// Run connects and streams work envelopes to handler until ctx is cancelled. Every envelope
// with an envelope_id is acked BEFORE handler runs — Slack redelivers anything not acked
// within ~3s, and a slow downstream pipe must not cause duplicate delivery (DECISIONS.md).
// Reconnects (disconnect frame, socket error, expired URL) fetch a fresh URL with
// full-jitter backoff; the backoff resets after each successful hello.
func (c *Client) Run(ctx context.Context, handler func(Envelope)) error {
	var attempt int
	for {
		if err := ctx.Err(); err != nil {
			return nil //nolint:nilerr // cancellation between connections is a clean shutdown, not an error
		}
		helloSeen, err := c.runOnce(ctx, handler)
		if ctx.Err() != nil {
			return nil //nolint:nilerr // Ctrl-C mid-read surfaces as a read error; treat as clean shutdown
		}
		if helloSeen {
			attempt = 0 // the connection was healthy; don't punish routine rotation
		}
		if err != nil {
			c.logf("socket error: %v", err)
		}
		attempt++
		wait := c.backoff(attempt)
		c.logf("reconnecting in %s…", wait.Round(time.Millisecond))
		if serr := sleepCtx(ctx, wait); serr != nil {
			return nil //nolint:nilerr // the only sleep error is ctx cancellation — clean shutdown
		}
	}
}

// runOnce handles a single connection's lifetime: dial, hello, envelope loop. It returns
// whether a hello arrived (the health signal for backoff reset) and the terminating error.
func (c *Client) runOnce(ctx context.Context, handler func(Envelope)) (bool, error) {
	url, err := c.OpenURL(ctx)
	if err != nil {
		return false, fmt.Errorf("apps.connections.open: %w", err)
	}
	if c.DebugReconnects {
		url += "&debug_reconnects=true"
	}
	ws, err := c.Dial(ctx, url)
	if err != nil {
		return false, fmt.Errorf("websocket dial: %w", err)
	}
	defer func() { _ = ws.Close(websocket.StatusNormalClosure, "bye") }()

	var helloSeen bool
	for {
		_, data, err := ws.Read(ctx)
		if err != nil {
			return helloSeen, err
		}
		var env Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			c.logf("skipping unparseable frame: %v", err)
			continue
		}
		switch env.Type {
		case "hello":
			helloSeen = true
			c.logf("connected (%d connection(s))", max(env.NumConnections, 1))
			continue
		case "disconnect":
			// refresh_requested is routine rotation; link_disabled means Socket Mode was
			// turned off — reconnect either way (the latter will fail loudly at OpenURL).
			return helloSeen, fmt.Errorf("server disconnect: %s", env.Reason)
		}
		if env.EnvelopeID != "" {
			ack, _ := json.Marshal(map[string]string{"envelope_id": env.EnvelopeID})
			if err := ws.Write(ctx, websocket.MessageText, ack); err != nil {
				return helloSeen, fmt.Errorf("ack %s: %w", env.EnvelopeID, err)
			}
		}
		handler(env)
	}
}

func (c *Client) logf(format string, args ...any) {
	if c.Log != nil {
		_, _ = fmt.Fprintf(c.Log, "listen: "+format+"\n", args...)
	}
}

// backoff is full jitter — random(0, min(30s, 1s·2^attempt)) — matching the API client's
// deliberate retry design.
func (c *Client) backoff(attempt int) time.Duration {
	ceiling := math.Min(float64(time.Second)*math.Pow(2, float64(attempt-1)), float64(30*time.Second))
	r := c.rng
	if r == nil {
		r = rand.Float64 //nolint:gosec // G404: jitter is not a security boundary
	}
	// Never hammer instantly even when the jitter rolls ~0.
	return time.Duration(r()*ceiling)/2 + time.Duration(ceiling/4)
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// EventsAPIPayload is the standard Events API callback wrapper inside an events_api
// envelope; Event is the actual event object (type message/reaction_added/…).
type EventsAPIPayload struct {
	TeamID  string          `json:"team_id"`
	Type    string          `json:"type"` // "event_callback"
	EventID string          `json:"event_id"`
	Event   json.RawMessage `json:"event"`
}

// ParseEvent extracts the inner event and its filter metadata from an events_api envelope.
// The metadata type is shared with the RTM transport (internal/slackevent).
func ParseEvent(env Envelope) (json.RawMessage, slackevent.Meta, error) {
	if env.Type != "events_api" {
		return nil, slackevent.Meta{}, errors.New("not an events_api envelope")
	}
	var p EventsAPIPayload
	if err := json.Unmarshal(env.Payload, &p); err != nil {
		return nil, slackevent.Meta{}, err
	}
	meta, err := slackevent.ParseMeta(p.Event)
	if err != nil {
		return nil, slackevent.Meta{}, err
	}
	return p.Event, meta, nil
}
