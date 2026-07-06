package commands

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

// runShared runs slackctl against srv with a SHARED config dir (so the recorder's DB persists
// across calls) and env session creds. It returns after each call closes the store.
func runShared(t *testing.T, dir string, srv *httptest.Server, args ...string) (string, string, error) {
	return runIn(t, dir, srv, "xoxb-test-token", args...)
}

// TestLog_RecordsPostsAndSearches is the end-to-end store flow: a post is recorded, then
// `log` and `log search` find it — all local, no API call for the search.
func TestLog_RecordsPostsAndSearches(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true,"channel":"C1","ts":"1720000000.001","message":{"type":"message","user":"U1","text":"deploy finished ok","ts":"1720000000.001"}}`))
	}))
	t.Cleanup(srv.Close)

	keyring.MockInit()
	dir := t.TempDir()

	// Post a message → the recorder persists it to the local store.
	_, errb, err := runShared(t, dir, srv, "msg", "post", "--channel", "C1", "--text", "deploy finished ok")
	require.NoError(t, err, errb)

	// `log` lists it (no server needed — reads the local DB).
	out, _, err := runShared(t, dir, nil, "log", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "deploy finished ok")

	// `log search` finds it by full-text.
	out, _, err = runShared(t, dir, nil, "log", "search", "deploy")
	require.NoError(t, err)
	assert.Contains(t, out, "deploy finished ok")

	// A non-matching search returns nothing.
	out, _, err = runShared(t, dir, nil, "log", "search", "nonexistent-term")
	require.NoError(t, err)
	assert.NotContains(t, out, "deploy finished")
}

// TestLog_RecordsFetchedHistory: fetching conversations.history persists each message, so it
// becomes searchable locally afterward.
func TestLog_RecordsFetchedHistory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true,"messages":[
			{"type":"message","user":"U1","ts":"2.0","text":"incident started"},
			{"type":"message","user":"U2","ts":"1.0","text":"all good"}
		],"response_metadata":{"next_cursor":""}}`))
	}))
	t.Cleanup(srv.Close)
	keyring.MockInit()
	dir := t.TempDir()

	_, _, err := runShared(t, dir, srv, "conversations", "history", "--channel", "C9")
	require.NoError(t, err)

	out, _, err := runShared(t, dir, nil, "log", "search", "incident", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "incident started")
	assert.Contains(t, out, "C9", "the channel from the request params is recorded")
}

func TestLog_NoStoreDisablesRecording(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true,"channel":"C1","ts":"1.0","message":{"type":"message","text":"secret","ts":"1.0"}}`))
	}))
	t.Cleanup(srv.Close)
	keyring.MockInit()
	dir := t.TempDir()

	// --no-store: the post must NOT be recorded.
	_, _, err := runShared(t, dir, srv, "msg", "post", "--channel", "C1", "--text", "secret", "--no-store")
	require.NoError(t, err)
	out, _, err := runShared(t, dir, nil, "log", "-o", "json")
	require.NoError(t, err)
	assert.NotContains(t, out, "secret", "--no-store must skip recording")
}

func TestLog_StatsAndPathAndPrune(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true,"channel":"C1","ts":"1.0","message":{"type":"message","text":"hi","ts":"1.0"}}`))
	}))
	t.Cleanup(srv.Close)
	keyring.MockInit()
	dir := t.TempDir()
	_, _, err := runShared(t, dir, srv, "msg", "post", "--channel", "C1", "--text", "hi")
	require.NoError(t, err)

	stats, _, err := runShared(t, dir, nil, "log", "stats", "-o", "json")
	require.NoError(t, err)
	assert.Contains(t, stats, `"messages": 1`)
	assert.Contains(t, stats, `"fts_enabled": true`)

	path, _, err := runShared(t, dir, nil, "log", "path")
	require.NoError(t, err)
	assert.True(t, strings.Contains(path, "history") && strings.HasSuffix(strings.TrimSpace(path), ".db"))

	// prune with a huge window removes nothing; with 0 removes the fresh row.
	out, _, err := runShared(t, dir, nil, "log", "prune", "--older-than", "0s")
	require.NoError(t, err)
	assert.Contains(t, out, "pruned 1 message")
}

func TestLog_SinceRejectsGarbage(t *testing.T) {
	keyring.MockInit()
	dir := t.TempDir()
	_, _, err := runShared(t, dir, nil, "log", "--since", "not-a-time")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --since")
}
