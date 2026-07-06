package commands

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"

	"github.com/jjuanrivvera/slackctl/internal/slackevent"
)

func TestConfigPath(t *testing.T) {
	out, _, err := run(t, nil, "config", "path")
	require.NoError(t, err)
	mustContain(t, out, "config.yaml")
}

func TestConfigView(t *testing.T) {
	out, _, err := run(t, nil, "config", "view", "-o", "json")
	require.NoError(t, err)
	mustContain(t, out, "token_storage")
}

func TestConfigUseAndListProfiles(t *testing.T) {
	keyring.MockInit()
	srv := newServer(t, routes{"auth.test": `{"ok":true,"team":"Acme","user":"bot","team_id":"T1","user_id":"U1"}`})
	dir := t.TempDir()

	_, _, err := runIn(t, dir, srv, "", "auth", "login", "--token", "xoxb-prod-token", "--workspace", "prod")
	require.NoError(t, err)

	out, _, err := runIn(t, dir, srv, "", "config", "list-profiles", "-o", "json")
	require.NoError(t, err)
	mustContain(t, out, "prod")

	_, _, err = runIn(t, dir, srv, "", "config", "use", "prod")
	require.NoError(t, err)

	_, _, err = runIn(t, dir, srv, "", "config", "set", "base_url", "https://slack.com/api", "--workspace", "prod")
	require.NoError(t, err)
}

func TestConfigUse_UnknownProfile(t *testing.T) {
	_, _, err := run(t, nil, "config", "use", "ghost")
	require.Error(t, err)
	mustContain(t, err.Error(), "no such profile")
}

func TestAliasSet_And_RejectsBuiltin(t *testing.T) {
	out, _, err := run(t, nil, "alias", "set", "unr", "conversations unreads")
	require.NoError(t, err)
	mustContain(t, out, "unr")

	_, _, err = run(t, nil, "alias", "set", "msg", "conversations list")
	require.Error(t, err)
	mustContain(t, err.Error(), "built-in")
}

func TestExpandAliases_BuiltinWins(t *testing.T) {
	got := ExpandAliases([]string{"conversations", "list"})
	assert.Equal(t, []string{"conversations", "list"}, got)
}

func TestCompletionGeneratesScripts(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish"} {
		out, _, err := run(t, nil, "completion", shell)
		require.NoError(t, err, shell)
		assert.NotEmpty(t, out)
	}
}

func TestInitWizard_EndToEnd(t *testing.T) {
	srv := newServer(t, routes{"auth.test": `{"ok":true,"team":"Acme","user":"bot","team_id":"T1","user_id":"U1","bot_id":"B1"}`})
	// Prompts: bot token, optional user token (skip), optional app token (skip).
	stdin := "xoxb-wizard-token\n\n\n"
	out, errb, err := runNoToken(t, srv, stdin, "init")
	require.NoError(t, err, errb)
	mustContain(t, out, `Workspace "default" ready`)
	mustContain(t, out, "conversations list")

	tok, kerr := keyring.Get("slackctl", "default")
	require.NoError(t, kerr)
	assert.Equal(t, "xoxb-wizard-token", tok)
}

func TestInitWizard_StoresOptionalTokens(t *testing.T) {
	srv := newServer(t, routes{"auth.test": `{"ok":true,"team":"Acme","user":"bot","team_id":"T1","user_id":"U1"}`})
	stdin := "xoxb-wizard-token\nxoxp-user-token\nxapp-1-app-token\n"
	_, _, err := runNoToken(t, srv, stdin, "init")
	require.NoError(t, err)
	userTok, err := keyring.Get("slackctl", "default#user")
	require.NoError(t, err)
	assert.Equal(t, "xoxp-user-token", userTok)
	appTok, err := keyring.Get("slackctl", "default#app")
	require.NoError(t, err)
	assert.Equal(t, "xapp-1-app-token", appTok)
}

// TestListen_EndToEnd wires the full path: `listen` resolves the app token, calls
// apps.connections.open on the mock Web API, dials the local websocket, acks, filters
// --dms, and prints NDJSON. The context deadline is the stop signal (Ctrl-C stand-in).
func TestListen_EndToEnd(t *testing.T) {
	// Local Socket Mode endpoint: hello, a DM message, a channel message, then hold open.
	wsSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		ctx := r.Context()
		go func() { // drain acks
			for {
				if _, _, err := c.Read(ctx); err != nil {
					return
				}
			}
		}()
		writeJSON := func(v string) { _ = c.Write(ctx, websocket.MessageText, []byte(v)) }
		writeJSON(`{"type":"hello","num_connections":1}`)
		writeJSON(`{"envelope_id":"e1","type":"events_api","payload":{"type":"event_callback","event":{"type":"message","channel":"D1","channel_type":"im","user":"U1","ts":"1.0","text":"dm hello"}}}`)
		writeJSON(`{"envelope_id":"e2","type":"events_api","payload":{"type":"event_callback","event":{"type":"message","channel":"C1","channel_type":"channel","user":"U2","ts":"2.0","text":"channel noise"}}}`)
		<-ctx.Done()
	}))
	t.Cleanup(wsSrv.Close)
	wsURL := "ws" + strings.TrimPrefix(wsSrv.URL, "http")

	apiSrv := newServer(t, routes{
		"apps.connections.open": `{"ok":true,"url":"` + wsURL + `"}`,
	})

	keyring.MockInit()
	t.Setenv("SLACK_APP_TOKEN", "xapp-1-test")
	t.Setenv("SLACK_BOT_TOKEN", "xoxb-test")
	t.Setenv("SLACK_XOXC_TOKEN", "")
	t.Setenv("SLACK_XOXD_TOKEN", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("NO_COLOR", "1")

	root := NewRootCmd()
	var out, errb strings.Builder
	root.SetOut(&out)
	root.SetErr(&errb)
	// App token present → auto selects Socket Mode.
	root.SetArgs([]string{"listen", "--dms", "--json", "--base-url", apiSrv.URL})

	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	require.NoError(t, root.ExecuteContext(ctx), errb.String())

	got := out.String()
	mustContain(t, got, "dm hello")
	assert.NotContains(t, got, "channel noise", "--dms must filter out channel events")
	var event map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.SplitN(strings.TrimSpace(got), "\n", 2)[0]), &event),
		"--json output must be one JSON object per line")
	mustContain(t, errb.String(), "connected")
}

// TestListen_RTM_EndToEnd wires the RTM path: with only session creds (no app token),
// `listen` auto-selects RTM, calls rtm.connect on the mock Web API, dials the local
// WebSocket, and streams raw event frames — filtered by --channels.
func TestListen_RTM_EndToEnd(t *testing.T) {
	wsSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		ctx := r.Context()
		go func() {
			for {
				if _, _, err := c.Read(ctx); err != nil {
					return
				}
			}
		}()
		writeJSON := func(v string) { _ = c.Write(ctx, websocket.MessageText, []byte(v)) }
		writeJSON(`{"type":"hello"}`)
		// RTM frames ARE the event objects — no envelope.
		writeJSON(`{"type":"message","channel":"C1","user":"U1","ts":"1.0","text":"rtm hello"}`)
		writeJSON(`{"type":"message","channel":"C9","user":"U2","ts":"2.0","text":"other channel"}`)
		<-ctx.Done()
	}))
	t.Cleanup(wsSrv.Close)
	wsURL := "ws" + strings.TrimPrefix(wsSrv.URL, "http")

	apiSrv := newServer(t, routes{
		"rtm.connect": `{"ok":true,"url":"` + wsURL + `"}`,
	})

	keyring.MockInit()
	for _, v := range []string{"SLACK_BOT_TOKEN", "SLACK_USER_TOKEN", "SLACK_APP_TOKEN", "SLACKCTL_TOKEN"} {
		t.Setenv(v, "")
	}
	t.Setenv("SLACK_XOXC_TOKEN", "xoxc-session")
	t.Setenv("SLACK_XOXD_TOKEN", "xoxd-cookie")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("NO_COLOR", "1")

	root := NewRootCmd()
	var out, errb strings.Builder
	root.SetOut(&out)
	root.SetErr(&errb)
	root.SetArgs([]string{"listen", "--channels", "C1", "--json", "--base-url", apiSrv.URL})

	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	require.NoError(t, root.ExecuteContext(ctx), errb.String())

	got := out.String()
	mustContain(t, got, "rtm hello")
	assert.NotContains(t, got, "other channel", "--channels C1 must filter out C9 events")
	mustContain(t, errb.String(), "connected (RTM)")
}

func TestHumanEventLine(t *testing.T) {
	event := json.RawMessage(`{"type":"message","text":"hola","subtype":"","reaction":""}`)
	line := humanEventLine(event, metaFor("message", "C1", "U1", "9.0"))
	assert.Contains(t, line, "message")
	assert.Contains(t, line, "C1")
	assert.Contains(t, line, "U1")
	assert.Contains(t, line, "hola")

	reaction := json.RawMessage(`{"type":"reaction_added","reaction":"tada"}`)
	line = humanEventLine(reaction, metaFor("reaction_added", "C2", "U2", "9.1"))
	assert.Contains(t, line, ":tada:")
}

func metaFor(typ, ch, user, ts string) (m slackevent.Meta) {
	m.Type = typ
	m.Channel = ch
	m.User = user
	m.TS = ts
	return m
}
