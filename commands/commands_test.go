package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"

	"github.com/jjuanrivvera/slackctl/internal/auth"
)

func TestConversationsList_TableWithDefaultColumns(t *testing.T) {
	srv := newServer(t, routes{
		"conversations.list": `{"ok":true,"channels":[
			{"id":"C1","name":"general","is_private":false,"is_archived":false,"num_members":12},
			{"id":"C2","name":"secret","is_private":true,"is_archived":false,"num_members":3}]}`,
	})
	out, _, err := run(t, srv, "conversations", "list")
	require.NoError(t, err)
	mustContain(t, out, "general")
	mustContain(t, out, "NUM_MEMBERS")
	assert.NotContains(t, out, `"ok"`, "the envelope must not leak into tables")
}

func TestConversationsList_PaginatesWithAll(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Query().Get("cursor") == "" {
			_, _ = w.Write([]byte(`{"ok":true,"channels":[{"id":"C1","name":"a"}],"response_metadata":{"next_cursor":"n2"}}`))
			return
		}
		_, _ = w.Write([]byte(`{"ok":true,"channels":[{"id":"C2","name":"b"}],"response_metadata":{"next_cursor":""}}`))
	}))
	t.Cleanup(srv.Close)
	out, _, err := run(t, srv, "conversations", "list", "--all", "-o", "json")
	require.NoError(t, err)
	assert.Equal(t, 2, calls)
	mustContain(t, out, "C1")
	mustContain(t, out, "C2")
}

func TestConversationsInfo_ExtractsChannel(t *testing.T) {
	srv := newServer(t, routes{
		"conversations.info": `{"ok":true,"channel":{"id":"C1","name":"general","topic":{"value":"hi"}}}`,
	})
	out, _, err := run(t, srv, "conversations", "info", "--channel", "C1", "-o", "json")
	require.NoError(t, err)
	var ch map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &ch))
	assert.Equal(t, "general", ch["name"])
	_, hasOK := ch["ok"]
	assert.False(t, hasOK)
}

func TestMsgPost_SendsFormAndRendersTS(t *testing.T) {
	var form map[string][]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		form = r.PostForm
		_, _ = w.Write([]byte(`{"ok":true,"channel":"C1","ts":"1720000000.000100"}`))
	}))
	t.Cleanup(srv.Close)
	out, _, err := run(t, srv, "msg", "post", "--channel", "C1", "--text", "hola", "--thread-ts", "1720000000.000001")
	require.NoError(t, err)
	assert.Equal(t, "hola", form["text"][0])
	assert.Equal(t, "1720000000.000001", form["thread_ts"][0])
	mustContain(t, out, "1720000000.000100")
}

func TestMsgPost_BlocksMustBeValidJSON(t *testing.T) {
	_, _, err := run(t, nil, "msg", "post", "--channel", "C1", "--blocks", "{not json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--blocks must be valid JSON")
}

func TestErrorHintSurfaces(t *testing.T) {
	srv := newServer(t, routes{
		"conversations.join": `{"ok":false,"error":"channel_not_found"}`,
	})
	_, _, err := run(t, srv, "conversations", "join", "--channel", "CBAD")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "channel_not_found")
	assert.Contains(t, err.Error(), "hint:")
}

func TestSearchMessages_UsesUserTokenAndExtractsMatches(t *testing.T) {
	var authz string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"ok":true,"messages":{"matches":[{"ts":"1.0","username":"ada","text":"deploy done","permalink":"https://x"}],"paging":{"count":20,"total":1,"page":1,"pages":1}}}`))
	}))
	t.Cleanup(srv.Close)
	out, _, err := run(t, srv, "search", "messages", "--query", "deploy")
	require.NoError(t, err)
	assert.Equal(t, "Bearer xoxp-test-token", authz, "search must use the user token")
	mustContain(t, out, "deploy done")
}

func TestSearch_FailsFastWithoutUserToken(t *testing.T) {
	_, _, err := runNoToken(t, nil, "", "search", "messages", "--query", "x")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--kind user")
}

func TestSavedList_RequiresUserToken(t *testing.T) {
	var authz string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"ok":true,"items":[{"type":"message","channel":"C1"}]}`))
	}))
	t.Cleanup(srv.Close)
	out, _, err := run(t, srv, "saved", "list")
	require.NoError(t, err)
	assert.Equal(t, "Bearer xoxp-test-token", authz)
	mustContain(t, out, "message")
}

func TestAsUserSwitchesToken(t *testing.T) {
	var authz string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"ok":true,"channels":[]}`))
	}))
	t.Cleanup(srv.Close)
	_, _, err := run(t, srv, "conversations", "list", "--as-user")
	require.NoError(t, err)
	assert.Equal(t, "Bearer xoxp-test-token", authz)
}

func TestUsersSearch_FiltersClientSide(t *testing.T) {
	srv := newServer(t, routes{
		"users.list": `{"ok":true,"members":[
			{"id":"U1","name":"ada","real_name":"Ada Lovelace","profile":{"email":"ada@example.com"}},
			{"id":"U2","name":"bob","real_name":"Bob Builder","profile":{"email":"bob@example.com"}}]}`,
	})
	out, _, err := run(t, srv, "users", "search", "lovelace", "-o", "json")
	require.NoError(t, err)
	mustContain(t, out, "U1")
	assert.NotContains(t, out, "U2")
}

func TestUsersSearch_MatchesEmail(t *testing.T) {
	srv := newServer(t, routes{
		"users.list": `{"ok":true,"members":[{"id":"U9","name":"x","profile":{"email":"unique@corp.io"}}]}`,
	})
	out, _, err := run(t, srv, "users", "search", "unique@corp.io", "-o", "id")
	require.NoError(t, err)
	mustContain(t, out, "U9")
}

func TestUnreads_CompositeCountsAndFilters(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "users.conversations"):
			_, _ = w.Write([]byte(`{"ok":true,"channels":[{"id":"D1","user":"U7","is_im":true},{"id":"C1","name":"general"}]}`))
		case strings.HasSuffix(r.URL.Path, "conversations.info"):
			if r.URL.Query().Get("channel") == "D1" {
				_, _ = w.Write([]byte(`{"ok":true,"channel":{"id":"D1","unread_count":3,"unread_count_display":2,"last_read":"1.0"}}`))
				return
			}
			_, _ = w.Write([]byte(`{"ok":true,"channel":{"id":"C1","unread_count":0,"last_read":"2.0"}}`))
		default:
			t.Fatalf("unexpected call %s", r.URL.Path)
		}
	}))
	t.Cleanup(srv.Close)
	out, _, err := run(t, srv, "conversations", "unreads", "-o", "json")
	require.NoError(t, err)
	mustContain(t, out, `"D1"`)
	assert.NotContains(t, out, `"C1"`, "zero-unread rows are dropped without --include-zero")

	out, _, err = run(t, srv, "conversations", "unreads", "--include-zero", "-o", "json")
	require.NoError(t, err)
	mustContain(t, out, `"C1"`)
}

func TestAuthLoginStatusLogoutFlow(t *testing.T) {
	srv := newServer(t, routes{
		"auth.test": `{"ok":true,"url":"https://acme.slack.com/","team":"Acme","user":"bot","team_id":"T1","user_id":"U1","bot_id":"B1"}`,
	})
	keyring.MockInit()
	dir := t.TempDir()

	out, errb, err := runIn(t, dir, srv, "", "auth", "login", "--token", "xoxb-secret-1")
	require.NoError(t, err, errb)
	mustContain(t, out, `stored bot token`)
	mustContain(t, errb, "verified as bot in Acme")

	out, _, err = runIn(t, dir, srv, "", "auth", "status", "-o", "json")
	require.NoError(t, err)
	mustContain(t, out, `"team": "Acme"`)

	out, _, err = runIn(t, dir, srv, "", "auth", "logout")
	require.NoError(t, err)
	mustContain(t, out, "logged out")

	_, _, err = runIn(t, dir, srv, "", "auth", "status")
	require.Error(t, err, "status must fail once the token is gone")
}

func TestAuthLogin_UserKindStoredSeparately(t *testing.T) {
	srv := newServer(t, routes{
		"auth.test": `{"ok":true,"team":"Acme","user":"ada","team_id":"T1","user_id":"U2"}`,
	})
	keyring.MockInit()
	dir := t.TempDir()
	_, _, err := runIn(t, dir, srv, "", "auth", "login", "--kind", "user", "--token", "xoxp-secret-2")
	require.NoError(t, err)

	tok, err := keyring.Get(auth.Service, auth.Key("default", auth.KindUser))
	require.NoError(t, err)
	assert.Equal(t, "xoxp-secret-2", tok)
	_, err = keyring.Get(auth.Service, "default")
	assert.Error(t, err, "the bot slot must stay empty")
}

func TestAuthLogin_AppKindSkipsVerify(t *testing.T) {
	// No auth.test route: an app-token login must not call it.
	srv := newServer(t, routes{})
	keyring.MockInit()
	dir := t.TempDir()
	out, _, err := runIn(t, dir, srv, "", "auth", "login", "--kind", "app", "--token", "xapp-1-A1-secret")
	require.NoError(t, err)
	mustContain(t, out, "stored app token")
}

func TestAuthLogin_RejectsBadKindAndBadToken(t *testing.T) {
	_, _, err := runNoToken(t, nil, "", "auth", "login", "--kind", "nope", "--token", "xoxb-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --kind")

	_, _, err = runNoToken(t, nil, "", "auth", "login", "--token", "not-a-slack-token", "--no-verify")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not look like a Slack token")
}

func TestDoctor_ReportsChecksAndFailsOnBadToken(t *testing.T) {
	srv := newServer(t, routes{
		"auth.test": `{"ok":false,"error":"invalid_auth"}`,
	})
	out, _, err := run(t, srv, "doctor")
	require.Error(t, err)
	mustContain(t, out, "✗ API reachable + token valid")
	mustContain(t, out, "✓ credentials resolvable")
}

func TestDoctor_HealthyPath(t *testing.T) {
	srv := newServer(t, routes{
		"auth.test": `{"ok":true,"team":"Acme","user":"bot","team_id":"T1","user_id":"U1"}`,
	})
	out, _, err := run(t, srv, "doctor", "--json")
	require.NoError(t, err)
	mustContain(t, out, `"ok": true`)
	mustContain(t, out, "bot @ Acme")
}

func TestAPIEscapeHatch(t *testing.T) {
	var method, body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		require.NoError(t, r.ParseForm())
		body = r.PostForm.Get("channel")
		_, _ = w.Write([]byte(`{"ok":true,"echo":true}`))
	}))
	t.Cleanup(srv.Close)
	out, _, err := run(t, srv, "api", "conversations.setTopic", "-q", "channel=C1", "-q", "topic=hey", "-o", "json")
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, method, "raw calls default to write semantics")
	assert.Equal(t, "C1", body)
	mustContain(t, out, `"echo"`)
}

func TestAPIEscapeHatch_IdempotentIsGET(t *testing.T) {
	var method string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)
	_, _, err := run(t, srv, "api", "auth.test", "--idempotent")
	require.NoError(t, err)
	assert.Equal(t, http.MethodGet, method)
}

func TestConfigViewRedactsNothingSecret(t *testing.T) {
	srv := newServer(t, routes{
		"auth.test": `{"ok":true,"team":"Acme","user":"bot","team_id":"T1","user_id":"U1","bot_id":"B1"}`,
	})
	keyring.MockInit()
	dir := t.TempDir()
	_, _, err := runIn(t, dir, srv, "", "auth", "login", "--token", "xoxb-super-secret")
	require.NoError(t, err)
	out, _, err := runIn(t, dir, srv, "", "config", "view")
	require.NoError(t, err)
	assert.NotContains(t, out, "xoxb-super-secret", "config view must never show a token")
	mustContain(t, out, "Acme")
}

func TestListen_RejectsDryRun(t *testing.T) {
	_, _, err := run(t, nil, "listen", "--dry-run")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "dry-run")
}

func TestListen_NeedsAppToken(t *testing.T) {
	_, _, err := runNoToken(t, nil, "", "listen")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--kind app")
}

func TestInvalidOutputFormatRejected(t *testing.T) {
	_, _, err := run(t, nil, "conversations", "list", "-o", "nope")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --output")
}

func TestVersionCommand(t *testing.T) {
	out, _, err := run(t, nil, "version")
	require.NoError(t, err)
	mustContain(t, out, "slackctl")
}

func TestWorkspaceFlagSelectsProfile(t *testing.T) {
	srv := newServer(t, routes{
		"auth.test": `{"ok":true,"team":"Beta","user":"bot","team_id":"T2","user_id":"U1"}`,
	})
	keyring.MockInit()
	dir := t.TempDir()
	_, _, err := runIn(t, dir, srv, "", "auth", "login", "--workspace", "beta", "--token", "xoxb-beta-token")
	require.NoError(t, err)
	out, _, err := runIn(t, dir, srv, "", "auth", "status", "--workspace", "beta", "-o", "json")
	require.NoError(t, err)
	mustContain(t, out, `"workspace": "beta"`)

	// The hidden --profile alias must reach the same profile.
	out, _, err = runIn(t, dir, srv, "", "auth", "status", "--profile", "beta", "-o", "json")
	require.NoError(t, err)
	mustContain(t, out, `"workspace": "beta"`)
}
