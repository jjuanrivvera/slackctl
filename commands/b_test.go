package commands

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

func TestConversationsExport_JSONL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// One page, newest-first, no next cursor.
		_, _ = w.Write([]byte(`{"ok":true,"messages":[
			{"type":"message","ts":"2.0","user":"U2","text":"second"},
			{"type":"message","ts":"1.0","user":"U1","text":"first"}
		],"response_metadata":{"next_cursor":""}}`))
	}))
	t.Cleanup(srv.Close)
	out, _, err := run(t, srv, "conversations", "export", "--channel", "C1")
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	require.Len(t, lines, 2, "one JSON object per line")
	assert.Contains(t, lines[0], "second")
	assert.Contains(t, lines[1], "first")
}

func TestConversationsExport_ToFileWithThreads(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
		switch method {
		case "conversations.history":
			_, _ = w.Write([]byte(`{"ok":true,"messages":[
				{"type":"message","ts":"10.0","user":"U1","text":"root","thread_ts":"10.0","reply_count":1}
			],"response_metadata":{"next_cursor":""}}`))
		case "conversations.replies":
			_, _ = w.Write([]byte(`{"ok":true,"messages":[
				{"type":"message","ts":"10.0","user":"U1","text":"root"},
				{"type":"message","ts":"11.0","user":"U2","text":"a reply"}
			],"response_metadata":{"next_cursor":""}}`))
		default:
			t.Fatalf("unexpected %s", method)
		}
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	dest := filepath.Join(dir, "history.jsonl")
	_, errb, err := run(t, srv, "conversations", "export", "--channel", "C1", "--threads", "--out", dest)
	require.NoError(t, err, errb)
	data, err := os.ReadFile(dest)
	require.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	require.Len(t, lines, 2, "root + one reply (parent deduped from the replies call)")
	assert.Contains(t, string(data), "root")
	assert.Contains(t, string(data), "a reply")
	assert.Contains(t, errb, "exported 2 messages")
}

func TestMsgTemplate_RendersAndPosts(t *testing.T) {
	var text string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		text = r.PostForm.Get("text")
		_, _ = w.Write([]byte(`{"ok":true,"channel":"C1","ts":"1.0"}`))
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	tmpl := filepath.Join(dir, "alert.tmpl")
	require.NoError(t, os.WriteFile(tmpl, []byte("🚨 {{.service}} is {{.status}}"), 0o600))

	_, _, err := run(t, srv, "msg", "template", "--channel", "C1", "--file", tmpl,
		"--set", "service=api", "--set", "status=down")
	require.NoError(t, err)
	assert.Equal(t, "🚨 api is down", text)
}

func TestMsgTemplate_MissingVarErrors(t *testing.T) {
	dir := t.TempDir()
	tmpl := filepath.Join(dir, "x.tmpl")
	require.NoError(t, os.WriteFile(tmpl, []byte("{{.missing}}"), 0o600))
	_, _, err := run(t, nil, "msg", "template", "--channel", "C1", "--file", tmpl)
	require.Error(t, err, "an unresolved template variable must fail, not post an empty message")
}

func TestMsgTemplate_BlocksMode(t *testing.T) {
	var blocks string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		blocks = r.PostForm.Get("blocks")
		_, _ = w.Write([]byte(`{"ok":true,"channel":"C1","ts":"1.0"}`))
	}))
	t.Cleanup(srv.Close)
	dir := t.TempDir()
	tmpl := filepath.Join(dir, "card.tmpl")
	require.NoError(t, os.WriteFile(tmpl, []byte(`[{"type":"section","text":{"type":"mrkdwn","text":"*{{.title}}*"}}]`), 0o600))
	_, _, err := run(t, srv, "msg", "template", "--channel", "C1", "--file", tmpl, "--set", "title=Deploy", "--blocks")
	require.NoError(t, err)
	assert.JSONEq(t, `[{"type":"section","text":{"type":"mrkdwn","text":"*Deploy*"}}]`, blocks)
}

// TestListen_SinceBackfill: --since replays channel history before the live stream. We mock
// rtm.connect + a ws that only says hello, so after backfill the stream idles until the
// context deadline; the backfilled message must appear in the output.
func TestListen_SinceBackfill(t *testing.T) {
	wsSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		_ = c.Write(r.Context(), websocket.MessageText, []byte(`{"type":"hello"}`))
		<-r.Context().Done()
	}))
	t.Cleanup(wsSrv.Close)
	wsURL := "ws" + strings.TrimPrefix(wsSrv.URL, "http")

	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
		switch method {
		case "conversations.history":
			_, _ = w.Write([]byte(`{"ok":true,"messages":[{"type":"message","ts":"9.0","user":"U1","text":"backfilled note"}],"response_metadata":{"next_cursor":""}}`))
		case "rtm.connect":
			_, _ = w.Write([]byte(`{"ok":true,"url":"` + wsURL + `"}`))
		default:
			t.Fatalf("unexpected %s", method)
		}
	}))
	t.Cleanup(apiSrv.Close)

	keyring.MockInit()
	for _, v := range []string{"SLACK_BOT_TOKEN", "SLACK_USER_TOKEN", "SLACK_APP_TOKEN", "SLACKCTL_TOKEN"} {
		t.Setenv(v, "")
	}
	t.Setenv("SLACK_XOXC_TOKEN", "xoxc-s")
	t.Setenv("SLACK_XOXD_TOKEN", "xoxd-c")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("NO_COLOR", "1")

	root := NewRootCmd()
	var outbuf, errbuf strings.Builder
	root.SetOut(&outbuf)
	root.SetErr(&errbuf)
	root.SetArgs([]string{"listen", "--channels", "C1", "--since", "1h", "--json", "--base-url", apiSrv.URL})

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	require.NoError(t, root.ExecuteContext(ctx), errbuf.String())
	assert.Contains(t, outbuf.String(), "backfilled note", "--since must replay history before the live stream")
}

func TestResolveSince(t *testing.T) {
	// A raw ts passes through; a duration becomes an absolute unix bound in the past.
	assert.Equal(t, "1720000000.000000", resolveSince("1720000000.000000"))
	got := resolveSince("1h")
	assert.NotEqual(t, "1h", got)
	assert.Regexp(t, `^\d+$`, got)
}
