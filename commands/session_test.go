package commands

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

// TestSessionEnv_BacksBotCommands is the browser-session path: with only
// SLACK_XOXC_TOKEN + SLACK_XOXD_TOKEN set (no bot token anywhere), a plain bot-kind
// command must authenticate via the session pair.
func TestSessionEnv_BacksBotCommands(t *testing.T) {
	var gotAuth, gotCookie string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotCookie = r.Header.Get("Cookie")
		_, _ = w.Write([]byte(`{"ok":true,"channels":[{"id":"C1","name":"general"}]}`))
	}))
	t.Cleanup(srv.Close)

	keyring.MockInit()
	for _, v := range []string{"SLACK_BOT_TOKEN", "SLACK_USER_TOKEN", "SLACK_APP_TOKEN", "SLACKCTL_TOKEN"} {
		t.Setenv(v, "")
	}
	t.Setenv("SLACK_XOXC_TOKEN", "xoxc-session")
	t.Setenv("SLACK_XOXD_TOKEN", "xoxd-cookie")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("NO_COLOR", "1")

	root := NewRootCmd()
	root.SetArgs([]string{"conversations", "list", "--base-url", srv.URL})
	require.NoError(t, root.ExecuteContext(t.Context()))
	assert.Equal(t, "Bearer xoxc-session", gotAuth)
	assert.Equal(t, "d=xoxd-cookie", gotCookie)
}

// TestSessionEnv_BacksUserOnlyCommands proves session creds satisfy user-token-only
// methods (search), since a session is the user's own identity.
func TestSessionEnv_BacksUserOnlyCommands(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"ok":true,"messages":{"matches":[{"ts":"1.0","text":"hit"}]}}`))
	}))
	t.Cleanup(srv.Close)

	keyring.MockInit()
	for _, v := range []string{"SLACK_BOT_TOKEN", "SLACK_USER_TOKEN", "SLACK_APP_TOKEN", "SLACKCTL_TOKEN"} {
		t.Setenv(v, "")
	}
	t.Setenv("SLACK_XOXC_TOKEN", "xoxc-session")
	t.Setenv("SLACK_XOXD_TOKEN", "xoxd-cookie")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("NO_COLOR", "1")

	root := NewRootCmd()
	root.SetArgs([]string{"search", "messages", "--query", "x", "--base-url", srv.URL})
	require.NoError(t, root.ExecuteContext(t.Context()))
	assert.Equal(t, "Bearer xoxc-session", gotAuth, "session backs the user-only search method")
}

// TestListen_ForceSocketWithoutAppTokenFails: forcing Socket Mode with only session creds
// must fail fast — apps.connections.open needs a real xapp app-level token, and session
// creds are NOT an app-token fallback.
func TestListen_ForceSocketWithoutAppTokenFails(t *testing.T) {
	keyring.MockInit()
	for _, v := range []string{"SLACK_BOT_TOKEN", "SLACK_USER_TOKEN", "SLACK_APP_TOKEN", "SLACKCTL_TOKEN"} {
		t.Setenv(v, "")
	}
	t.Setenv("SLACK_XOXC_TOKEN", "xoxc-session")
	t.Setenv("SLACK_XOXD_TOKEN", "xoxd-cookie")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("NO_COLOR", "1")

	root := NewRootCmd()
	root.SetArgs([]string{"listen", "--transport", "socket"})
	err := root.ExecuteContext(t.Context())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "app token")
}

// TestListen_AutoPicksRTMForSession: with only session creds and no app token, `listen`
// (auto transport) must select RTM. We prove the selection without opening a socket by
// forcing --transport auto and a --dry-run, which errors specifically because the chosen
// transport tried to build its live client (not because of a missing app token).
func TestListen_AutoPicksRTMForSession(t *testing.T) {
	keyring.MockInit()
	for _, v := range []string{"SLACK_BOT_TOKEN", "SLACK_USER_TOKEN", "SLACK_APP_TOKEN", "SLACKCTL_TOKEN"} {
		t.Setenv(v, "")
	}
	t.Setenv("SLACK_XOXC_TOKEN", "xoxc-session")
	t.Setenv("SLACK_XOXD_TOKEN", "xoxd-cookie")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("NO_COLOR", "1")

	root := NewRootCmd()
	root.SetArgs([]string{"listen", "--dry-run"})
	err := root.ExecuteContext(t.Context())
	require.Error(t, err)
	// RTM was selected and its client built (session creds resolved) — the error is the
	// dry-run guard, NOT "no app token" (which would mean it wrongly chose Socket Mode).
	assert.Contains(t, err.Error(), "dry-run")
	assert.NotContains(t, err.Error(), "app token")
}

// TestSessionLogin_StoresPairAndVerifies drives `auth login --kind session` end to end.
func TestSessionLogin_StoresPairAndVerifies(t *testing.T) {
	srv := newServer(t, routes{
		"auth.test": `{"ok":true,"team":"Acme","user":"ada","team_id":"T1","user_id":"U9"}`,
	})
	keyring.MockInit()
	dir := t.TempDir()

	out, errb, err := runIn(t, dir, srv, "", "auth", "login", "--kind", "session",
		"--token", "xoxc-abc", "--cookie", "xoxd-def")
	require.NoError(t, err, errb)
	mustContain(t, out, "stored session credentials")
	mustContain(t, errb, "verified as ada in Acme")

	// The pair round-trips through the store.
	creds, err := keyring.Get("slackctl", "default#session")
	require.NoError(t, err)
	assert.Contains(t, creds, "xoxc-abc")
	assert.Contains(t, creds, "xoxd-def")

	// A subsequent bot-kind command uses the stored session (no bot token present).
	var gotAuth string
	wsSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"ok":true,"channels":[]}`))
	}))
	t.Cleanup(wsSrv.Close)
	_, _, err = runIn(t, dir, wsSrv, "", "conversations", "list")
	require.NoError(t, err)
	assert.Equal(t, "Bearer xoxc-abc", gotAuth)
}

func TestSessionLogin_RejectsBadCookie(t *testing.T) {
	_, _, err := runNoToken(t, nil, "", "auth", "login", "--kind", "session",
		"--token", "xoxc-abc", "--cookie", "", "--no-verify")
	// With an empty cookie and no prompt input, NewSessionAuth rejects the pair.
	require.Error(t, err)
}
