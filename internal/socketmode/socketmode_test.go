package socketmode

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// wsServer runs a real websocket endpoint whose per-connection script is driven by the
// test. Each accepted connection calls script(conn, connIndex); acks written by the client
// land in the acks channel.
type wsServer struct {
	t     *testing.T
	srv   *httptest.Server
	acks  chan string
	mu    sync.Mutex
	conns int
}

func newWSServer(t *testing.T, script func(ctx context.Context, c *websocket.Conn, connIdx int)) *wsServer {
	t.Helper()
	s := &wsServer{t: t, acks: make(chan string, 64)}
	s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		s.mu.Lock()
		s.conns++
		idx := s.conns
		s.mu.Unlock()

		ctx := r.Context()
		// Drain client acks concurrently so the script can just push frames.
		go func() {
			for {
				_, data, err := c.Read(ctx)
				if err != nil {
					return
				}
				var ack struct {
					EnvelopeID string `json:"envelope_id"`
				}
				if json.Unmarshal(data, &ack) == nil && ack.EnvelopeID != "" {
					s.acks <- ack.EnvelopeID
				}
			}
		}()
		script(ctx, c, idx)
	}))
	t.Cleanup(s.srv.Close)
	return s
}

func (s *wsServer) wsURL() string { return "ws" + strings.TrimPrefix(s.srv.URL, "http") }

func (s *wsServer) connCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conns
}

func send(ctx context.Context, c *websocket.Conn, v any) {
	b, _ := json.Marshal(v)
	_ = c.Write(ctx, websocket.MessageText, b)
}

func hello() map[string]any { return map[string]any{"type": "hello", "num_connections": 1} }

func eventEnvelope(id, eventJSON string) map[string]any {
	return map[string]any{
		"envelope_id": id,
		"type":        "events_api",
		"payload": map[string]any{
			"type":  "event_callback",
			"event": json.RawMessage(eventJSON),
		},
	}
}

// newTestClient wires a Client to the ws server with instant backoff.
func newTestClient(s *wsServer) *Client {
	c := New(func(context.Context) (string, error) { return s.wsURL(), nil }, io.Discard)
	c.rng = func() float64 { return 0 }
	return c
}

func TestRun_AcksBeforeHandler(t *testing.T) {
	s := newWSServer(t, func(ctx context.Context, c *websocket.Conn, _ int) {
		send(ctx, c, hello())
		send(ctx, c, eventEnvelope("env-1", `{"type":"message","channel":"D1","user":"U1","text":"hi","ts":"1.0"}`))
		<-ctx.Done()
	})
	ctx, cancel := context.WithCancel(t.Context())
	var got []Envelope
	done := make(chan error, 1)
	go func() {
		done <- newTestClient(s).Run(ctx, func(e Envelope) {
			got = append(got, e)
			cancel() // one event is enough
		})
	}()

	select {
	case id := <-s.acks:
		assert.Equal(t, "env-1", id)
	case <-time.After(5 * time.Second):
		t.Fatal("no ack received")
	}
	require.NoError(t, <-done)
	require.Len(t, got, 1)
	assert.Equal(t, "events_api", got[0].Type)
}

func TestRun_ReconnectsOnDisconnectFrame(t *testing.T) {
	s := newWSServer(t, func(ctx context.Context, c *websocket.Conn, idx int) {
		send(ctx, c, hello())
		if idx == 1 {
			// Routine rotation: the client must come back on a fresh connection.
			send(ctx, c, map[string]any{"type": "disconnect", "reason": "refresh_requested"})
			return
		}
		send(ctx, c, eventEnvelope("env-2", `{"type":"message","channel":"C1","ts":"2.0"}`))
		<-ctx.Done()
	})
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() {
		done <- newTestClient(s).Run(ctx, func(e Envelope) { cancel() })
	}()

	select {
	case id := <-s.acks:
		assert.Equal(t, "env-2", id)
	case <-time.After(5 * time.Second):
		t.Fatal("never got the post-reconnect event")
	}
	require.NoError(t, <-done)
	assert.GreaterOrEqual(t, s.connCount(), 2, "must have dialed a second connection")
}

func TestRun_FreshURLPerConnection(t *testing.T) {
	s := newWSServer(t, func(ctx context.Context, c *websocket.Conn, idx int) {
		send(ctx, c, hello())
		if idx == 1 {
			send(ctx, c, map[string]any{"type": "disconnect", "reason": "refresh_requested"})
			return
		}
		send(ctx, c, eventEnvelope("env-3", `{"type":"message","ts":"3.0"}`))
		<-ctx.Done()
	})
	var urlCalls int
	var mu sync.Mutex
	ctx, cancel := context.WithCancel(t.Context())
	c := New(func(context.Context) (string, error) {
		mu.Lock()
		urlCalls++
		mu.Unlock()
		return s.wsURL(), nil
	}, io.Discard)
	c.rng = func() float64 { return 0 }
	done := make(chan error, 1)
	go func() { done <- c.Run(ctx, func(Envelope) { cancel() }) }()
	require.NoError(t, <-done)
	mu.Lock()
	defer mu.Unlock()
	assert.GreaterOrEqual(t, urlCalls, 2, "every (re)connection must mint a fresh single-use URL")
}

func TestRun_CancelledContextIsCleanShutdown(t *testing.T) {
	s := newWSServer(t, func(ctx context.Context, c *websocket.Conn, _ int) {
		send(ctx, c, hello())
		<-ctx.Done()
	})
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- newTestClient(s).Run(ctx, func(Envelope) {}) }()
	time.Sleep(100 * time.Millisecond) // let it connect
	cancel()
	select {
	case err := <-done:
		assert.NoError(t, err, "Ctrl-C must not report an error")
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not exit on cancel")
	}
}

func TestRun_OpenURLFailureRetries(t *testing.T) {
	var calls int
	var mu sync.Mutex
	ctx, cancel := context.WithCancel(t.Context())
	c := New(func(context.Context) (string, error) {
		mu.Lock()
		calls++
		n := calls
		mu.Unlock()
		if n >= 3 {
			cancel()
		}
		return "", fmt.Errorf("boom")
	}, io.Discard)
	c.rng = func() float64 { return 0 }
	require.NoError(t, c.Run(ctx, func(Envelope) {})) // exits cleanly once cancelled
	mu.Lock()
	defer mu.Unlock()
	assert.GreaterOrEqual(t, calls, 3, "OpenURL failures must keep retrying with backoff")
}

func TestRun_SkipsUnparseableFrames(t *testing.T) {
	s := newWSServer(t, func(ctx context.Context, c *websocket.Conn, _ int) {
		send(ctx, c, hello())
		_ = c.Write(ctx, websocket.MessageText, []byte("not json"))
		send(ctx, c, eventEnvelope("env-4", `{"type":"message","ts":"4.0"}`))
		<-ctx.Done()
	})
	ctx, cancel := context.WithCancel(t.Context())
	var got int
	done := make(chan error, 1)
	go func() {
		done <- newTestClient(s).Run(ctx, func(Envelope) { got++; cancel() })
	}()
	require.NoError(t, <-done)
	assert.Equal(t, 1, got, "garbage frames are skipped, the stream continues")
}

func TestParseEvent(t *testing.T) {
	env := Envelope{
		Type: "events_api",
		Payload: json.RawMessage(`{"type":"event_callback","event":
			{"type":"message","channel":"D111","channel_type":"im","user":"U1","ts":"5.0","text":"hey"}}`),
	}
	event, meta, err := ParseEvent(env)
	require.NoError(t, err)
	assert.Equal(t, "message", meta.Type)
	assert.Equal(t, "D111", meta.ChannelOf())
	assert.True(t, meta.IsDM())
	assert.Contains(t, string(event), `"hey"`)

	_, _, err = ParseEvent(Envelope{Type: "slash_commands"})
	assert.Error(t, err)
}

func TestEventMeta_ReactionItemChannel(t *testing.T) {
	var meta EventMeta
	require.NoError(t, json.Unmarshal([]byte(`{"type":"reaction_added","user":"U1","item":{"channel":"C42"}}`), &meta))
	assert.Equal(t, "C42", meta.ChannelOf())
	assert.False(t, meta.IsDM())
}

func TestEventMeta_IsDMByPrefix(t *testing.T) {
	m := EventMeta{Channel: "D0AB12"}
	assert.True(t, m.IsDM())
	m = EventMeta{Channel: "C0AB12"}
	assert.False(t, m.IsDM())
}

func TestBackoff_CappedAndNonZero(t *testing.T) {
	c := New(nil, io.Discard)
	for attempt := 1; attempt < 12; attempt++ {
		d := c.backoff(attempt)
		assert.Greater(t, d, time.Duration(0))
		assert.LessOrEqual(t, d, 30*time.Second)
	}
}

func TestDebugReconnectsAppendsParam(t *testing.T) {
	var dialed string
	c := New(func(context.Context) (string, error) { return "wss://x/link?ticket=t", nil }, io.Discard)
	c.DebugReconnects = true
	c.Dial = func(_ context.Context, url string) (conn, error) {
		dialed = url
		return nil, fmt.Errorf("stop here")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_ = c.Run(ctx, func(Envelope) {})
	// Run exits before dialing on a pre-cancelled ctx; call runOnce directly instead.
	_, _ = c.runOnce(t.Context(), func(Envelope) {})
	assert.Equal(t, "wss://x/link?ticket=t&debug_reconnects=true", dialed)
}
