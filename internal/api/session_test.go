package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSessionAuth_Validation(t *testing.T) {
	_, err := NewSessionAuth("xoxc-123", "xoxd-abc")
	require.NoError(t, err)

	_, err = NewSessionAuth("xoxb-123", "xoxd-abc")
	require.Error(t, err, "a non-xoxc token must be rejected")

	_, err = NewSessionAuth("xoxc-123", "")
	require.Error(t, err, "the xoxd cookie is required")
}

func TestSessionAuth_SendsTokenAndCookie(t *testing.T) {
	var gotAuth, gotCookie string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotCookie = r.Header.Get("Cookie")
		_, _ = w.Write([]byte(`{"ok":true,"team":"Acme","user":"ada","team_id":"T1","user_id":"U1"}`))
	}))
	t.Cleanup(srv.Close)

	authr, err := NewSessionAuth("xoxc-secret", "xoxd-cookieval")
	require.NoError(t, err)
	c := New(authr, WithBaseURL(srv.URL), WithHTTPClient(srv.Client()), WithRPS(0))

	id, err := c.AuthTest(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "Acme", id.Team)
	assert.Equal(t, "Bearer xoxc-secret", gotAuth)
	assert.Equal(t, "d=xoxd-cookieval", gotCookie, "the d cookie carries the session")
}

func TestSessionAuth_MethodAndRedaction(t *testing.T) {
	authr, err := NewSessionAuth("xoxc-supersecret", "xoxd-supersecret")
	require.NoError(t, err)
	assert.Equal(t, "session-token", authr.Method())
	assert.NotContains(t, authr.Redacted(), "supersecret")
	// The cookie is redacted in dry-run header output unless show-token is set.
	assert.NotContains(t, authr.ExtraHeaders(true)["Cookie"], "supersecret")
	assert.Contains(t, authr.ExtraHeaders(false)["Cookie"], "supersecret")
}

func TestSessionAuth_DryRunCurlIncludesRedactedCookie(t *testing.T) {
	authr, err := NewSessionAuth("xoxc-secret", "xoxd-cookieval")
	require.NoError(t, err)
	var buf writerString
	c := New(authr, WithDryRun(true), WithDryRunWriter(&buf))
	_, err = c.Call(t.Context(), "auth.test", nil, true)
	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "Authorization: Bearer xoxc-****")
	assert.Contains(t, out, "Cookie: d=xoxd-****")
	assert.NotContains(t, out, "cookieval", "the cookie value must be redacted by default")

	// With show-token, the real cookie is emitted so the curl actually works.
	var buf2 writerString
	c2 := New(authr, WithDryRun(true), WithShowToken(true), WithDryRunWriter(&buf2))
	_, err = c2.Call(t.Context(), "auth.test", nil, true)
	require.NoError(t, err)
	assert.Contains(t, buf2.String(), "Cookie: d=xoxd-cookieval")
}

// writerString is a tiny io.Writer capturing output, avoiding a bytes.Buffer import churn.
type writerString struct{ b []byte }

func (w *writerString) Write(p []byte) (int, error) { w.b = append(w.b, p...); return len(p), nil }
func (w *writerString) String() string              { return string(w.b) }
