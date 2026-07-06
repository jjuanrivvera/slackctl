package rtm

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

// wsServer runs a real websocket endpoint driven per-connection by a script.
type wsServer struct {
	srv   *httptest.Server
	mu    sync.Mutex
	conns int
	pings chan json.RawMessage
}

func newWSServer(t *testing.T, script func(ctx context.Context, c *websocket.Conn, connIdx int)) *wsServer {
	t.Helper()
	s := &wsServer{pings: make(chan json.RawMessage, 16)}
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
		go func() { // capture anything the client writes (pings)
			for {
				_, data, err := c.Read(ctx)
				if err != nil {
					return
				}
				s.pings <- json.RawMessage(data)
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

func newTestClient(s *wsServer) *Client {
	c := New(func(context.Context) (string, error) { return s.wsURL(), nil }, io.Discard)
	c.rng = func() float64 { return 0 }
	c.PingInterval = -1 // disable the keepalive pinger in tests
	return c
}

func TestRun_StreamsRawEventFrames(t *testing.T) {
	s := newWSServer(t, func(ctx context.Context, c *websocket.Conn, _ int) {
		send(ctx, c, map[string]any{"type": "hello"})
		_ = c.Write(ctx, websocket.MessageText, []byte(`{"type":"message","channel":"D1","user":"U1","text":"hi","ts":"1.0"}`))
		<-ctx.Done()
	})
	ctx, cancel := context.WithCancel(t.Context())
	var got []json.RawMessage
	done := make(chan error, 1)
	go func() {
		done <- newTestClient(s).Run(ctx, func(frame json.RawMessage) {
			got = append(got, frame)
			cancel()
		})
	}()
	require.NoError(t, <-done)
	require.Len(t, got, 1, "the raw event frame is delivered as-is")
	assert.Contains(t, string(got[0]), `"hi"`)
}

func TestRun_HelloAndPongAreNotEvents(t *testing.T) {
	s := newWSServer(t, func(ctx context.Context, c *websocket.Conn, _ int) {
		send(ctx, c, map[string]any{"type": "hello"})
		send(ctx, c, map[string]any{"type": "pong", "reply_to": 1})
		_ = c.Write(ctx, websocket.MessageText, []byte(`{"type":"reaction_added","user":"U1","item":{"channel":"C1"}}`))
		<-ctx.Done()
	})
	ctx, cancel := context.WithCancel(t.Context())
	var got int
	done := make(chan error, 1)
	go func() {
		done <- newTestClient(s).Run(ctx, func(json.RawMessage) { got++; cancel() })
	}()
	require.NoError(t, <-done)
	assert.Equal(t, 1, got, "hello and pong frames are lifecycle, not events")
}

func TestRun_ReconnectsOnGoodbye(t *testing.T) {
	s := newWSServer(t, func(ctx context.Context, c *websocket.Conn, idx int) {
		send(ctx, c, map[string]any{"type": "hello"})
		if idx == 1 {
			send(ctx, c, map[string]any{"type": "goodbye"})
			return
		}
		_ = c.Write(ctx, websocket.MessageText, []byte(`{"type":"message","channel":"C1","ts":"2.0"}`))
		<-ctx.Done()
	})
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- newTestClient(s).Run(ctx, func(json.RawMessage) { cancel() }) }()
	require.NoError(t, <-done)
	assert.GreaterOrEqual(t, s.connCount(), 2, "a goodbye frame must trigger a reconnect")
}

func TestRun_FreshURLPerConnection(t *testing.T) {
	s := newWSServer(t, func(ctx context.Context, c *websocket.Conn, idx int) {
		send(ctx, c, map[string]any{"type": "hello"})
		if idx == 1 {
			send(ctx, c, map[string]any{"type": "goodbye"})
			return
		}
		_ = c.Write(ctx, websocket.MessageText, []byte(`{"type":"message","ts":"3.0"}`))
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
	c.PingInterval = -1
	done := make(chan error, 1)
	go func() { done <- c.Run(ctx, func(json.RawMessage) { cancel() }) }()
	require.NoError(t, <-done)
	mu.Lock()
	defer mu.Unlock()
	assert.GreaterOrEqual(t, urlCalls, 2, "every RTM (re)connection mints a fresh single-use URL")
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
		return "", fmt.Errorf("not_allowed_token_type")
	}, io.Discard)
	c.rng = func() float64 { return 0 }
	c.PingInterval = -1
	require.NoError(t, c.Run(ctx, func(json.RawMessage) {}))
	mu.Lock()
	defer mu.Unlock()
	assert.GreaterOrEqual(t, calls, 3)
}

func TestRun_CancelIsCleanShutdown(t *testing.T) {
	s := newWSServer(t, func(ctx context.Context, c *websocket.Conn, _ int) {
		send(ctx, c, map[string]any{"type": "hello"})
		<-ctx.Done()
	})
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- newTestClient(s).Run(ctx, func(json.RawMessage) {}) }()
	time.Sleep(100 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not exit on cancel")
	}
}

func TestPinger_SendsPings(t *testing.T) {
	s := newWSServer(t, func(ctx context.Context, c *websocket.Conn, _ int) {
		send(ctx, c, map[string]any{"type": "hello"})
		<-ctx.Done()
	})
	c := New(func(context.Context) (string, error) { return s.wsURL(), nil }, io.Discard)
	c.rng = func() float64 { return 0 }
	c.PingInterval = 20 * time.Millisecond // fast pings for the test
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- c.Run(ctx, func(json.RawMessage) {}) }()

	select {
	case p := <-s.pings:
		assert.Contains(t, string(p), `"type":"ping"`)
	case <-time.After(5 * time.Second):
		t.Fatal("no keepalive ping received")
	}
	cancel()
	<-done
}

func TestRun_SkipsUnparseableFrames(t *testing.T) {
	s := newWSServer(t, func(ctx context.Context, c *websocket.Conn, _ int) {
		send(ctx, c, map[string]any{"type": "hello"})
		_ = c.Write(ctx, websocket.MessageText, []byte("not json"))
		_ = c.Write(ctx, websocket.MessageText, []byte(`{"type":"message","ts":"4.0"}`))
		<-ctx.Done()
	})
	ctx, cancel := context.WithCancel(t.Context())
	var got int
	done := make(chan error, 1)
	go func() { done <- newTestClient(s).Run(ctx, func(json.RawMessage) { got++; cancel() }) }()
	require.NoError(t, <-done)
	assert.Equal(t, 1, got)
}

func TestDial_SendsCredentialHeaders(t *testing.T) {
	var gotCookie string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCookie = r.Header.Get("Cookie") // captured from the WebSocket UPGRADE request
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		send(r.Context(), c, map[string]any{"type": "hello"})
		<-r.Context().Done()
	}))
	t.Cleanup(srv.Close)
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")

	c := New(func(context.Context) (string, error) { return wsURL, nil }, io.Discard)
	c.rng = func() float64 { return 0 }
	c.PingInterval = -1
	c.Header = http.Header{"Cookie": {"d=xoxd-session-cookie"}}

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	// Cancel shortly after connect; we only need the handshake to happen.
	go func() {
		_, err := c.runOnce(ctx, func(json.RawMessage) {})
		done <- err
	}()
	// Give the handshake a moment, then stop.
	go func() { time.Sleep(200 * time.Millisecond); cancel() }()
	<-done
	assert.Equal(t, "d=xoxd-session-cookie", gotCookie, "the d cookie must ride the WebSocket handshake (RTM gateway re-validates it)")
}

func TestBackoff_CappedAndNonZero(t *testing.T) {
	c := New(nil, io.Discard)
	for attempt := 1; attempt < 12; attempt++ {
		d := c.backoff(attempt)
		assert.Greater(t, d, time.Duration(0))
		assert.LessOrEqual(t, d, 30*time.Second)
	}
}
