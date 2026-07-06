package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func openTemp(t *testing.T) *Store {
	t.Helper()
	st, err := Open(filepath.Join(t.TempDir(), "hist.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func TestRecordAndQuery(t *testing.T) {
	st := openTemp(t)
	ctx := t.Context()
	require.NoError(t, st.Record(ctx, Message{Workspace: "w", Channel: "C1", TS: "1.0", User: "U1", Type: "message", Text: "first"}))
	require.NoError(t, st.Record(ctx, Message{Workspace: "w", Channel: "C1", TS: "2.0", User: "U2", Type: "message", Text: "second"}))

	msgs, err := st.Query(ctx, Filter{})
	require.NoError(t, err)
	require.Len(t, msgs, 2)
	assert.Equal(t, "2.0", msgs[0].TS, "newest first")
	assert.Equal(t, "second", msgs[0].Text)
}

func TestRecord_DedupOnChannelTS(t *testing.T) {
	st := openTemp(t)
	ctx := t.Context()
	m := Message{Workspace: "w", Channel: "C1", TS: "1.0", Text: "hi"}
	require.NoError(t, st.Record(ctx, m))
	require.NoError(t, st.Record(ctx, m)) // same (channel, ts) — must dedup
	msgs, err := st.Query(ctx, Filter{})
	require.NoError(t, err)
	assert.Len(t, msgs, 1, "re-recording the same (channel, ts) must not duplicate")
}

func TestRecord_SkipsUnlocatable(t *testing.T) {
	st := openTemp(t)
	require.NoError(t, st.Record(t.Context(), Message{Workspace: "w", Text: "no channel/ts"}))
	msgs, _ := st.Query(t.Context(), Filter{})
	assert.Empty(t, msgs)
}

func TestQuery_Filters(t *testing.T) {
	st := openTemp(t)
	ctx := t.Context()
	_ = st.Record(ctx, Message{Workspace: "w", Channel: "C1", TS: "1.0", User: "U1", Text: "a"})
	_ = st.Record(ctx, Message{Workspace: "w", Channel: "C2", TS: "2.0", User: "U2", Text: "b"})
	_ = st.Record(ctx, Message{Workspace: "w", Channel: "C1", TS: "3.0", User: "U2", Text: "c"})

	byChan, _ := st.Query(ctx, Filter{Channel: "C1"})
	assert.Len(t, byChan, 2)
	byUser, _ := st.Query(ctx, Filter{User: "U2"})
	assert.Len(t, byUser, 2)
	since, _ := st.Query(ctx, Filter{Since: "2.0"})
	assert.Len(t, since, 2, "ts >= 2.0 keeps 2.0 and 3.0")
	lim, _ := st.Query(ctx, Filter{Limit: 1})
	assert.Len(t, lim, 1)
}

func TestSearch(t *testing.T) {
	st := openTemp(t)
	ctx := t.Context()
	_ = st.Record(ctx, Message{Workspace: "w", Channel: "C1", TS: "1.0", Text: "deploy failed in prod"})
	_ = st.Record(ctx, Message{Workspace: "w", Channel: "C1", TS: "2.0", Text: "lunch time"})

	hits, err := st.Search(ctx, "deploy", Filter{})
	require.NoError(t, err)
	require.Len(t, hits, 1)
	assert.Contains(t, hits[0].Text, "deploy failed")

	none, err := st.Search(ctx, "nonexistent", Filter{})
	require.NoError(t, err)
	assert.Empty(t, none)
}

func TestSearch_WithChannelFilter(t *testing.T) {
	st := openTemp(t)
	ctx := t.Context()
	_ = st.Record(ctx, Message{Workspace: "w", Channel: "C1", TS: "1.0", Text: "incident here"})
	_ = st.Record(ctx, Message{Workspace: "w", Channel: "C2", TS: "2.0", Text: "incident there"})
	hits, err := st.Search(ctx, "incident", Filter{Channel: "C1"})
	require.NoError(t, err)
	require.Len(t, hits, 1)
	assert.Equal(t, "C1", hits[0].Channel)
}

func TestStats(t *testing.T) {
	st := openTemp(t)
	ctx := t.Context()
	_ = st.Record(ctx, Message{Workspace: "w", Channel: "C1", TS: "1.0", Type: "message", Text: "a"})
	_ = st.Record(ctx, Message{Workspace: "w", Channel: "C2", TS: "2.0", Type: "message", Text: "b"})
	s, err := st.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), s.Messages)
	assert.Equal(t, int64(2), s.Channels)
	assert.Equal(t, "1.0", s.Oldest)
	assert.Equal(t, "2.0", s.Newest)
	assert.Equal(t, 2, s.ByType["message"])
}

func TestPrune(t *testing.T) {
	st := openTemp(t)
	ctx := t.Context()
	old := Message{Workspace: "w", Channel: "C1", TS: "1.0", Text: "old", RecordedAt: time.Now().Add(-48 * time.Hour)}
	fresh := Message{Workspace: "w", Channel: "C1", TS: "2.0", Text: "fresh"}
	require.NoError(t, st.Record(ctx, old))
	require.NoError(t, st.Record(ctx, fresh))

	n, err := st.Prune(ctx, 24*time.Hour)
	require.NoError(t, err)
	assert.Equal(t, int64(1), n, "the 48h-old row is pruned, the fresh one kept")
	msgs, _ := st.Query(ctx, Filter{})
	require.Len(t, msgs, 1)
	assert.Equal(t, "fresh", msgs[0].Text)
}

func TestRecordBatch_And_RawRoundTrip(t *testing.T) {
	st := openTemp(t)
	ctx := t.Context()
	raw := json.RawMessage(`{"type":"message","text":"x","blocks":[1,2]}`)
	n, err := st.RecordBatch(ctx, []Message{
		{Workspace: "w", Channel: "C1", TS: "1.0", Text: "x", Raw: raw},
		{Workspace: "w", Channel: "C1", TS: "2.0", Text: "y"},
	})
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	msgs, _ := st.Query(ctx, Filter{Channel: "C1", Limit: 10})
	require.Len(t, msgs, 2)
	// The oldest (1.0) carries the raw payload.
	assert.JSONEq(t, string(raw), string(msgs[1].Raw))
}

func TestFTSEnabled(t *testing.T) {
	st := openTemp(t)
	// modernc.org/sqlite ships FTS5, so search uses MATCH (not the LIKE fallback).
	assert.True(t, st.FTSEnabled(), "the pure-Go SQLite build should include FTS5")
}

func TestPathFor_RejectsTraversal(t *testing.T) {
	_, err := PathFor("/cfg", "../escape")
	require.Error(t, err)
	p, err := PathFor("/cfg", "acme")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join("/cfg", "history", "acme.db"), p)
}

func TestOpen_FilePerms(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix file mode check")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "hist.db")
	st, err := Open(path)
	require.NoError(t, err)
	defer func() { _ = st.Close() }()
	fi, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), fi.Mode().Perm(), "the DB holds message text; it must be 0600")
}

func TestSearch_FallsBackOnFTSSyntaxError(t *testing.T) {
	st := openTemp(t)
	ctx := t.Context()
	_ = st.Record(ctx, Message{Workspace: "w", Channel: "C1", TS: "1.0", Text: "on-call rotation"})
	// "on-call" is invalid FTS5 (reads -call as an operator); must fall back to LIKE, not error.
	hits, err := st.Search(ctx, "on-call", Filter{})
	require.NoError(t, err)
	require.Len(t, hits, 1)
	assert.Contains(t, hits[0].Text, "on-call")
}
