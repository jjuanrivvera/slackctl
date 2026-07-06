package commands

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBookmarks(t *testing.T) {
	cases := []struct {
		name   string
		method string
		body   string
		want   string
		args   []string
	}{
		{"list", "bookmarks.list", `{"ok":true,"bookmarks":[{"id":"Bk1","title":"Runbook","link":"https://x","type":"link"}]}`, "Runbook",
			[]string{"bookmarks", "list", "--channel", "C1"}},
		{"add", "bookmarks.add", `{"ok":true,"bookmark":{"id":"Bk2","title":"Wiki","link":"https://w"}}`, "Bk2",
			[]string{"bookmarks", "add", "--channel", "C1", "--title", "Wiki", "--link", "https://w"}},
		{"edit", "bookmarks.edit", `{"ok":true,"bookmark":{"id":"Bk2","title":"Wiki2"}}`, "Wiki2",
			[]string{"bookmarks", "edit", "--channel", "C1", "--bookmark", "Bk2", "--title", "Wiki2", "-o", "json"}},
		{"remove", "bookmarks.remove", `{"ok":true}`, "true",
			[]string{"bookmarks", "remove", "--channel", "C1", "--bookmark", "Bk2", "-o", "json"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := newServer(t, routes{tc.method: tc.body})
			out, errb, err := run(t, srv, tc.args...)
			require.NoError(t, err, errb)
			mustContain(t, out, tc.want)
		})
	}
}

func TestBookmarksAdd_SendsChannelIdParam(t *testing.T) {
	var form map[string][]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		form = r.PostForm
		_, _ = w.Write([]byte(`{"ok":true,"bookmark":{"id":"Bk1"}}`))
	}))
	t.Cleanup(srv.Close)
	_, _, err := run(t, srv, "bookmarks", "add", "--channel", "C9", "--title", "T", "--link", "https://x")
	require.NoError(t, err)
	assert.Equal(t, "C9", form["channel_id"][0], "--channel maps to channel_id")
	assert.Equal(t, "link", form["type"][0], "default type is link")
}

func TestAssistantSearchContext_BotToken(t *testing.T) {
	var authz string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"ok":true,"results":{"messages":[{"ts":"1.0","channel_id":"C1","author_user_id":"U1","message":"deploy failed"}]}}`))
	}))
	t.Cleanup(srv.Close)
	out, _, err := run(t, srv, "assistant", "search-context", "--query", "deploy")
	require.NoError(t, err)
	assert.Equal(t, "Bearer xoxb-test-token", authz, "assistant.search.context works with a bot token")
	mustContain(t, out, "deploy failed")
}

func TestCanvases(t *testing.T) {
	cases := []struct {
		name   string
		method string
		body   string
		want   string
		args   []string
	}{
		{"create", "canvases.create", `{"ok":true,"canvas_id":"F9"}`, "F9",
			[]string{"canvases", "create", "--title", "Runbook", "--content", `{"type":"markdown","markdown":"# hi"}`}},
		{"edit", "canvases.edit", `{"ok":true}`, "true",
			[]string{"canvases", "edit", "--canvas", "F9", "--changes", `[{"operation":"insert_at_end"}]`, "-o", "json"}},
		{"delete", "canvases.delete", `{"ok":true}`, "true",
			[]string{"canvases", "delete", "--canvas", "F9", "-o", "json"}},
		{"access-set", "canvases.access.set", `{"ok":true}`, "true",
			[]string{"canvases", "access-set", "--canvas", "F9", "--access-level", "write", "--channels", "C1", "-o", "json"}},
		{"sections-lookup", "canvases.sections.lookup", `{"ok":true,"sections":[{"id":"s1"}]}`, "s1",
			[]string{"canvases", "sections-lookup", "--canvas", "F9", "--criteria", `{"contains_text":"TODO"}`, "-o", "json"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := newServer(t, routes{tc.method: tc.body})
			out, errb, err := run(t, srv, tc.args...)
			require.NoError(t, err, errb)
			mustContain(t, out, tc.want)
		})
	}
}

func TestCanvasCreate_ContentIsJSONObject(t *testing.T) {
	var content string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		content = r.PostForm.Get("document_content")
		_, _ = w.Write([]byte(`{"ok":true,"canvas_id":"F1"}`))
	}))
	t.Cleanup(srv.Close)
	_, _, err := run(t, srv, "canvases", "create", "--content", `{"type":"markdown","markdown":"# hi"}`)
	require.NoError(t, err)
	assert.JSONEq(t, `{"type":"markdown","markdown":"# hi"}`, content, "content is sent as a JSON object, not a string")
}

func TestSetTopic_CanClearWithEmptyString(t *testing.T) {
	var form url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		form = r.PostForm
		_, _ = w.Write([]byte(`{"ok":true,"channel":{"id":"C1","topic":{"value":""}}}`))
	}))
	t.Cleanup(srv.Close)
	// An explicitly-empty --topic must send topic="" so the topic is CLEARED, not omitted.
	_, _, err := run(t, srv, "conversations", "set-topic", "--channel", "C1", "--topic", "", "-o", "json")
	require.NoError(t, err)
	_, sent := form["topic"]
	assert.True(t, sent, "an explicit empty --topic must be sent (to clear), not dropped")
	assert.Equal(t, "", form.Get("topic"))
}

func TestAssistantSearch_HasLimitFlag(t *testing.T) {
	var limit string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit = r.URL.Query().Get("limit")
		_, _ = w.Write([]byte(`{"ok":true,"results":{"messages":[]}}`))
	}))
	t.Cleanup(srv.Close)
	_, _, err := run(t, srv, "assistant", "search-context", "--query", "x", "--limit", "5")
	require.NoError(t, err, "the --limit flag referenced in the help must exist")
	assert.Equal(t, "5", limit)
}
