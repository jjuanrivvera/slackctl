package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestClient spins an httptest server and a Client pointed at it. The fast retry policy
// keeps failure-path tests off the real backoff clock.
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	auth, err := NewTokenAuth("xoxb-test-token")
	require.NoError(t, err)
	return New(auth,
		WithBaseURL(srv.URL),
		WithHTTPClient(srv.Client()),
		WithRPS(0), // no pacing in tests
		WithRetryPolicy(retryPolicy{maxAttempts: 3, base: time.Millisecond, max: 5 * time.Millisecond}),
	)
}

func TestCall_ReadIsGETWithBearer(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/conversations.list", r.URL.Path)
		assert.Equal(t, "Bearer xoxb-test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "50", r.URL.Query().Get("limit"))
		_, _ = w.Write([]byte(`{"ok":true,"channels":[{"id":"C1"}]}`))
	})
	raw, err := c.Call(t.Context(), "conversations.list", map[string]any{"limit": 50}, true)
	require.NoError(t, err)
	assert.Contains(t, string(raw), `"C1"`)
}

func TestCall_WriteIsFormPOST(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.Header.Get("Content-Type"), "application/x-www-form-urlencoded")
		require.NoError(t, r.ParseForm())
		assert.Equal(t, "C1", r.PostForm.Get("channel"))
		assert.Equal(t, "hi", r.PostForm.Get("text"))
		// Non-scalar args arrive as JSON text inside the form field (DECISIONS.md).
		assert.JSONEq(t, `[{"type":"section"}]`, r.PostForm.Get("blocks"))
		_, _ = w.Write([]byte(`{"ok":true,"ts":"123.456"}`))
	})
	blocks := []any{map[string]any{"type": "section"}}
	raw, err := c.Call(t.Context(), "chat.postMessage", map[string]any{"channel": "C1", "text": "hi", "blocks": blocks}, false)
	require.NoError(t, err)
	assert.Contains(t, string(raw), "123.456")
}

func TestCall_OkFalseIsTypedError(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ok":false,"error":"channel_not_found"}`)) // HTTP 200!
	})
	_, err := c.Call(t.Context(), "conversations.info", map[string]any{"channel": "C404"}, true)
	var ae *APIError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, "channel_not_found", ae.Code)
	assert.Equal(t, http.StatusOK, ae.StatusCode)
	assert.Contains(t, err.Error(), "conversations list") // the hint names the next move
}

func TestCall_MissingScopeCarriesNeeded(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ok":false,"error":"missing_scope","needed":"chat:write","provided":"users:read"}`))
	})
	_, err := c.Call(t.Context(), "chat.postMessage", nil, false)
	var ae *APIError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, "chat:write", ae.Needed)
	assert.Contains(t, err.Error(), "chat:write")
}

func TestCall_Retries429HonoringRetryAfter(t *testing.T) {
	var calls int
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"ok":false,"error":"ratelimited"}`))
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	// A 429 retries even for writes: the request was rejected, not processed.
	_, err := c.Call(t.Context(), "chat.postMessage", map[string]any{"channel": "C1", "text": "x"}, false)
	require.NoError(t, err)
	assert.Equal(t, 2, calls)
}

func TestCall_EnvelopeRatelimitedWithHTTP200Retries(t *testing.T) {
	var calls int
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls == 1 {
			_, _ = w.Write([]byte(`{"ok":false,"error":"ratelimited"}`)) // 200 + envelope error
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	_, err := c.Call(t.Context(), "users.list", nil, true)
	require.NoError(t, err)
	assert.Equal(t, 2, calls)
}

func TestCall_NoRetryForWriteOn5xx(t *testing.T) {
	var calls int
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"ok":false,"error":"internal_error"}`))
	})
	_, err := c.Call(t.Context(), "chat.postMessage", map[string]any{"text": "x"}, false)
	require.Error(t, err)
	assert.Equal(t, 1, calls, "a timed-out write may have landed — never replay it")
}

func TestCall_ReadRetriesOn5xx(t *testing.T) {
	var calls int
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	_, err := c.Call(t.Context(), "users.list", nil, true)
	require.NoError(t, err)
	assert.Equal(t, 3, calls)
}

func TestCall_NonJSONResponse(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`<html>bad gateway</html>`))
	})
	_, err := c.Call(t.Context(), "chat.postMessage", nil, false)
	var ae *APIError
	require.ErrorAs(t, err, &ae)
	assert.Equal(t, http.StatusBadGateway, ae.StatusCode)
	assert.Contains(t, ae.Body, "non-JSON")
}

func TestCallInto_Decodes(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true,"team":"acme","user":"bot"}`))
	})
	var out struct {
		Team string `json:"team"`
		User string `json:"user"`
	}
	require.NoError(t, c.CallInto(t.Context(), "auth.test", nil, true, &out))
	assert.Equal(t, "acme", out.Team)
}

func TestAuthTest_Identity(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/auth.test", r.URL.Path)
		_, _ = w.Write([]byte(`{"ok":true,"url":"https://acme.slack.com/","team":"Acme","user":"slackctl","team_id":"T1","user_id":"U1","bot_id":"B1"}`))
	})
	id, err := c.AuthTest(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "Acme", id.Team)
	assert.Equal(t, "B1", id.BotID)
}

func TestCallAllPages_WalksCursor(t *testing.T) {
	var cursors []string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		cur := r.URL.Query().Get("cursor")
		cursors = append(cursors, cur)
		switch cur {
		case "":
			_, _ = w.Write([]byte(`{"ok":true,"channels":[{"id":"C1"},{"id":"C2"}],"response_metadata":{"next_cursor":"page2"}}`))
		case "page2":
			_, _ = w.Write([]byte(`{"ok":true,"channels":[{"id":"C3"}],"response_metadata":{"next_cursor":""}}`))
		default:
			t.Fatalf("unexpected cursor %q", cur)
		}
	})
	raw, err := c.CallAllPages(t.Context(), "conversations.list", map[string]any{"limit": 2}, "channels", 0)
	require.NoError(t, err)
	var items []map[string]string
	require.NoError(t, json.Unmarshal(raw, &items))
	assert.Len(t, items, 3)
	assert.Equal(t, []string{"", "page2"}, cursors)
}

func TestCallAllPages_MaxCapsItems(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true,"members":["U1","U2","U3"],"response_metadata":{"next_cursor":"more"}}`))
	})
	raw, err := c.CallAllPages(t.Context(), "conversations.members", nil, "members", 2)
	require.NoError(t, err)
	var items []string
	require.NoError(t, json.Unmarshal(raw, &items))
	assert.Equal(t, []string{"U1", "U2"}, items, "stop at max even though the API had more")
}

func TestCallAllPages_MissingKeyErrors(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	_, err := c.CallAllPages(t.Context(), "conversations.list", nil, "channels", 0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"channels"`)
}

func TestDryRun_PrintsRedactedCurlAndSkipsNetwork(t *testing.T) {
	c := newTestClient(t, func(http.ResponseWriter, *http.Request) {
		t.Fatal("dry-run must not hit the network")
	})
	var buf bytes.Buffer
	c.DryRun = true
	c.dryRunW = &buf

	raw, err := c.Call(t.Context(), "chat.postMessage", map[string]any{"channel": "C1", "text": "it's here"}, false)
	require.NoError(t, err)
	assert.Nil(t, raw)

	curl := buf.String()
	assert.Contains(t, curl, "curl -sS -X POST")
	assert.Contains(t, curl, "/chat.postMessage")
	assert.Contains(t, curl, "Authorization: Bearer xoxb-****")
	assert.NotContains(t, curl, "xoxb-test-token")
	assert.Contains(t, curl, `--data-urlencode 'channel=C1'`)
	assert.Contains(t, curl, `it'\''s here`, "single quotes must be shell-escaped")
}

func TestDryRun_GETPutsParamsInURL(t *testing.T) {
	c := newTestClient(t, func(http.ResponseWriter, *http.Request) {
		t.Fatal("dry-run must not hit the network")
	})
	var buf bytes.Buffer
	c.DryRun = true
	c.dryRunW = &buf
	_, err := c.Call(t.Context(), "conversations.history", map[string]any{"channel": "C1", "limit": 5}, true)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "conversations.history?channel=C1&limit=5")
	assert.NotContains(t, buf.String(), "-X POST")
}

func TestDryRun_ShowTokenExposesSecret(t *testing.T) {
	c := newTestClient(t, nil)
	var buf bytes.Buffer
	c.DryRun = true
	c.ShowToken = true
	c.dryRunW = &buf
	_, err := c.Call(t.Context(), "auth.test", nil, true)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "xoxb-test-token")
}

func TestCall_ContextCancelStopsRetryLoop(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	var calls int
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		calls++
		cancel() // cancel mid-flight; the retry sleep must observe it
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"ok":false,"error":"ratelimited"}`))
	})
	_, err := c.Call(ctx, "users.list", nil, true)
	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 1, calls)
}

func TestNewTokenAuth_Validation(t *testing.T) {
	for _, tok := range []string{"xoxb-1-2-3", "xoxp-user", "xapp-1-A1-2-abc", "xoxe.xoxp-1-refresh"} {
		_, err := NewTokenAuth(tok)
		assert.NoError(t, err, tok)
	}
	for _, tok := range []string{"", "  ", "ghp_notslack", "Bearer xoxb-1"} {
		_, err := NewTokenAuth(tok)
		assert.Error(t, err, tok)
	}
}

func TestTokenAuth_MethodAndRedaction(t *testing.T) {
	cases := map[string]string{
		"xoxb-abc":        "bot-token",
		"xoxp-abc":        "user-token",
		"xoxe.xoxp-1-abc": "user-token",
		"xapp-1-abc":      "app-token",
		"xoxs-legacy":     "token",
	}
	for tok, want := range cases {
		a, err := NewTokenAuth(tok)
		require.NoError(t, err)
		assert.Equal(t, want, a.Method(), tok)
		assert.NotContains(t, a.Redacted(), "abc", "redaction must strip the secret part")
	}
	assert.Equal(t, "xoxb-****", RedactToken("xoxb-123456789"))
	assert.Equal(t, "****", RedactToken("nodash"))
}

func TestAPIError_HintsAreActionable(t *testing.T) {
	cases := map[string]string{
		"invalid_auth":                          "auth login",
		"token_expired":                         "--kind session",
		"missing_scope":                         "OAuth & Permissions",
		"not_allowed_token_type":                "--kind user",
		"channel_not_found":                     "conversations list",
		"not_in_channel":                        "conversations join",
		"is_archived":                           "unarchive",
		"ratelimited":                           "--rps",
		"invalid_cursor":                        "--cursor",
		"internal_error":                        "transient",
		"method_not_supported_for_channel_type": "channel kind",
	}
	for code, want := range cases {
		e := &APIError{Code: code, Method: "m", StatusCode: 200}
		assert.Contains(t, e.Error(), want, code)
	}
	// Status fallback when the body carried no usable code (proxy 502, raw 429).
	assert.Contains(t, (&APIError{StatusCode: 502, Method: "m"}).Error(), "transient")
	assert.Contains(t, (&APIError{StatusCode: 429, Method: "m"}).Error(), "rate limited")
	assert.Contains(t, (&APIError{StatusCode: 404, Method: "m"}).Error(), "base-url")
	assert.Contains(t, (&APIError{StatusCode: 403, Method: "m"}).Error(), "unauthorized")
}

func TestAPIError_RetryAfterInHint(t *testing.T) {
	e := &APIError{Code: "ratelimited", RetryAfter: 7}
	assert.Contains(t, e.Error(), "wait 7s")
}

func TestIsCode(t *testing.T) {
	err := error(&APIError{Code: "channel_not_found"})
	assert.True(t, IsCode(err, "channel_not_found"))
	assert.False(t, IsCode(err, "invalid_auth"))
	assert.False(t, IsCode(assert.AnError, "channel_not_found"))
}

func TestScalarString(t *testing.T) {
	assert.Equal(t, "x", scalarString("x"))
	assert.Equal(t, "", scalarString(nil))
	assert.Equal(t, "5", scalarString(5))
	assert.Equal(t, "true", scalarString(true))
	assert.Equal(t, `{"a":1}`, scalarString(json.RawMessage(`{"a":1}`)))
	assert.Equal(t, `["a","b"]`, scalarString([]string{"a", "b"}))
}

func TestVerboseLogsResponses(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true,"warning":"superfluous_charset"}`))
	})
	var buf bytes.Buffer
	c.Verbose = true
	c.verboseW = &buf
	_, err := c.Call(t.Context(), "api.test", nil, true)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "api.test 200")
	assert.Contains(t, buf.String(), "warning: superfluous_charset")
}

func TestTruncate(t *testing.T) {
	assert.Equal(t, "abc", truncate("abc", 5))
	assert.Equal(t, strings.Repeat("a", 5)+"…", truncate(strings.Repeat("a", 9), 5))
}
