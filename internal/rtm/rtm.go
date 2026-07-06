// Package rtm is a hand-written client for Slack's legacy Real Time Messaging API — the
// WebSocket stream that works with a user/session token (xoxc+xoxd), the credential a
// slack-mcp-style setup already has. Unlike Socket Mode (which needs an app-level token and
// acks enveloped work items), RTM delivers raw event objects and requires no ack; the client
// just reads frames and keeps the socket alive with periodic pings. RTM is legacy and not
// officially supported for xoxc tokens — see DECISIONS.md ("listen transports").
package rtm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"time"

	"github.com/coder/websocket"
)

// conn is the slice of *websocket.Conn the client uses — injectable for tests.
type conn interface {
	Read(ctx context.Context) (websocket.MessageType, []byte, error)
	Write(ctx context.Context, typ websocket.MessageType, p []byte) error
	Close(code websocket.StatusCode, reason string) error
}

// Client streams raw RTM event frames.
type Client struct {
	// OpenURL fetches a fresh wss:// URL (rtm.connect). RTM URLs are single-use and must be
	// connected within ~30s, so every (re)connection asks for a new one.
	OpenURL func(ctx context.Context) (string, error)
	// Dial opens the websocket. Defaults to coder/websocket; injectable for tests.
	Dial func(ctx context.Context, url string) (conn, error)
	// Log receives connection-lifecycle notes (connected, reconnecting). Never event data.
	Log io.Writer
	// PingInterval is how often to send an application-level RTM ping to keep the
	// connection alive. Zero uses the 30s default; negative disables pings (tests).
	PingInterval time.Duration

	rng func() float64
}

// New builds a Client over an OpenURL func (which calls rtm.connect).
func New(openURL func(ctx context.Context) (string, error), logW io.Writer) *Client {
	return &Client{
		OpenURL: openURL,
		Dial: func(ctx context.Context, url string) (conn, error) {
			c, _, err := websocket.Dial(ctx, url, nil) //nolint:bodyclose // library owns the handshake response
			if err != nil {
				return nil, err
			}
			c.SetReadLimit(1 << 22) // RTM message events can be large
			return c, nil
		},
		Log: logW,
		rng: rand.Float64, //nolint:gosec // G404: reconnect jitter is not a security boundary
	}
}

// Run connects and streams every event frame to handler until ctx is cancelled. RTM needs
// no per-event ack. Reconnects (socket error, expired URL) fetch a fresh URL with full-jitter
// backoff; the backoff resets after each successful hello.
func (c *Client) Run(ctx context.Context, handler func(json.RawMessage)) error {
	var attempt int
	for {
		if err := ctx.Err(); err != nil {
			return nil //nolint:nilerr // cancellation between connections is a clean shutdown
		}
		helloSeen, err := c.runOnce(ctx, handler)
		if ctx.Err() != nil {
			return nil //nolint:nilerr // Ctrl-C mid-read surfaces as a read error; clean shutdown
		}
		if helloSeen {
			attempt = 0
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

// rtmFrame is the minimal shape needed to route a frame: its type. Everything else is passed
// through to the handler verbatim.
type rtmFrame struct {
	Type    string `json:"type"`
	ReplyTo int    `json:"reply_to,omitempty"`
}

func (c *Client) runOnce(ctx context.Context, handler func(json.RawMessage)) (bool, error) {
	url, err := c.OpenURL(ctx)
	if err != nil {
		return false, fmt.Errorf("rtm.connect: %w", err)
	}
	ws, err := c.Dial(ctx, url)
	if err != nil {
		return false, fmt.Errorf("websocket dial: %w", err)
	}
	defer func() { _ = ws.Close(websocket.StatusNormalClosure, "bye") }()

	// Keep-alive: RTM expects the client to ping periodically. A cancelable child context
	// stops the pinger when this connection ends.
	pingCtx, stopPing := context.WithCancel(ctx)
	defer stopPing()
	go c.pinger(pingCtx, ws)

	var helloSeen bool
	for {
		_, data, err := ws.Read(ctx)
		if err != nil {
			return helloSeen, err
		}
		var f rtmFrame
		if err := json.Unmarshal(data, &f); err != nil {
			c.logf("skipping unparseable frame: %v", err)
			continue
		}
		switch f.Type {
		case "hello":
			helloSeen = true
			c.logf("connected (RTM)")
			continue
		case "pong":
			continue // keep-alive reply; not an event
		case "goodbye":
			// Slack asks us to reconnect (server maintenance/rotation).
			return helloSeen, fmt.Errorf("server goodbye")
		case "error":
			c.logf("RTM error frame: %s", truncate(string(data), 200))
			continue
		case "":
			// Reply acks to our own sends carry reply_to but no type; ignore.
			continue
		}
		handler(json.RawMessage(data))
	}
}

// pinger sends an application-level RTM ping at PingInterval so the server does not drop an
// idle connection. Failures are ignored: the read loop will surface a dead socket.
func (c *Client) pinger(ctx context.Context, ws conn) {
	interval := c.PingInterval
	if interval == 0 {
		interval = 30 * time.Second
	}
	if interval < 0 {
		return // disabled (tests)
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	id := 1
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			ping, _ := json.Marshal(map[string]any{"id": id, "type": "ping"})
			if err := ws.Write(ctx, websocket.MessageText, ping); err != nil {
				return
			}
			id++
		}
	}
}

func (c *Client) logf(format string, args ...any) {
	if c.Log != nil {
		_, _ = fmt.Fprintf(c.Log, "listen: "+format+"\n", args...)
	}
}

func (c *Client) backoff(attempt int) time.Duration {
	ceiling := math.Min(float64(time.Second)*math.Pow(2, float64(attempt-1)), float64(30*time.Second))
	r := c.rng
	if r == nil {
		r = rand.Float64 //nolint:gosec // G404: jitter is not a security boundary
	}
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

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
