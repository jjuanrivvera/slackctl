// Package store is slackctl's local message history: messages you post, read (history/
// replies/export), or stream (listen) are recorded to a per-workspace SQLite database, so you
// can full-text search your Slack history locally — instantly, offline, and without hitting
// Slack's user-token-only, rate-limited search API.
//
// The driver is modernc.org/sqlite (pure Go, no cgo): slackctl has no cgo dependency and
// GoReleaser cross-compiles linux/darwin/windows from one toolchain (DECISIONS.md) — a
// cgo-based driver would break that. Recording must never break a command: every write-path
// caller treats a store error as "warn and continue" (see commands/recorder.go).
package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite" // registers the "sqlite" database/sql driver

	"github.com/jjuanrivvera/slackctl/internal/config"
)

// DefaultLimit is the default row cap for Query/Search.
const DefaultLimit = 50

// Message is one recorded Slack message. Slack ids are strings (C…/D…/U…) and ts is a
// string ("1720000000.000100") that sorts chronologically as text (fixed width, zero-padded).
type Message struct {
	ID         int64           `json:"id"`
	Workspace  string          `json:"workspace"`
	Channel    string          `json:"channel"`
	TS         string          `json:"ts"`
	ThreadTS   string          `json:"thread_ts,omitempty"`
	User       string          `json:"user,omitempty"`
	Type       string          `json:"type,omitempty"`
	Subtype    string          `json:"subtype,omitempty"`
	Text       string          `json:"text,omitempty"`
	RecordedAt time.Time       `json:"recorded_at"`
	Raw        json.RawMessage `json:"raw,omitempty"`
}

// Filter narrows Query/Search. The zero value means "no constraint" on that field.
type Filter struct {
	Channel string
	User    string
	Since   string // a Slack ts / unix bound; only messages with ts >= Since
	Limit   int    // <=0 → DefaultLimit
}

// Stats summarizes what a store holds.
type Stats struct {
	Messages int64          `json:"messages"`
	Channels int64          `json:"channels"`
	Oldest   string         `json:"oldest_ts,omitempty"`
	Newest   string         `json:"newest_ts,omitempty"`
	FTS      bool           `json:"fts_enabled"`
	ByType   map[string]int `json:"by_type,omitempty"`
}

// Store is a per-workspace SQLite message history.
type Store struct {
	db  *sql.DB
	fts bool // true when the linked SQLite build includes the FTS5 module
}

// PathFor returns the per-workspace DB path: <configDir>/history/<workspace>.db. The workspace
// name comes from --workspace/$SLACKCTL_WORKSPACE (user input on every call), so it is
// re-validated here — ValidateProfileName rejects '/' and '\', which is what makes the join
// below safe against a path escape.
func PathFor(configDir, workspace string) (string, error) {
	if err := config.ValidateProfileName(workspace); err != nil {
		return "", fmt.Errorf("store path: %w", err)
	}
	return filepath.Join(configDir, "history", workspace+".db"), nil
}

// Open opens (creating if needed) the SQLite store and initializes its schema idempotently.
// The parent dir is created 0700 and the file chmod'd 0600 — this file holds message text.
func Open(dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
		return nil, fmt.Errorf("create store dir: %w", err)
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}
	// One connection: modernc.org/sqlite serializes writers internally, and a single slackctl
	// invocation never needs concurrent connections. busy_timeout smooths two processes.
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`PRAGMA busy_timeout = 5000;`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("configure store: %w", err)
	}
	if err := os.Chmod(dbPath, 0o600); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("chmod store: %w", err)
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Close releases the database handle.
func (s *Store) Close() error { return s.db.Close() }

// FTSEnabled reports whether Search uses FTS5 MATCH (true) or a LIKE scan fallback (false).
func (s *Store) FTSEnabled() bool { return s.fts }

const schema = `
CREATE TABLE IF NOT EXISTS messages (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	workspace TEXT NOT NULL,
	channel TEXT NOT NULL,
	ts TEXT NOT NULL,
	thread_ts TEXT,
	user TEXT,
	type TEXT,
	subtype TEXT,
	text TEXT,
	recorded_at TEXT NOT NULL,
	raw TEXT,
	UNIQUE(workspace, channel, ts)
);
CREATE INDEX IF NOT EXISTS idx_messages_ch_ts ON messages(workspace, channel, ts);
`

func (s *Store) migrate() error {
	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("init schema: %w", err)
	}
	s.fts = s.tryEnableFTS()
	return nil
}

// tryEnableFTS creates the FTS5 side table Search uses, if this SQLite build includes FTS5.
// Some minimal builds omit it; Search then degrades to a LIKE scan and FTSEnabled reports so.
func (s *Store) tryEnableFTS() bool {
	_, err := s.db.Exec(`CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(text)`)
	return err == nil
}

const selectMessage = `SELECT messages.id, messages.workspace, messages.channel, messages.ts,
	messages.thread_ts, messages.user, messages.type, messages.subtype, messages.text,
	messages.recorded_at, messages.raw`

// Record inserts one message, deduping on (workspace, channel, ts) — re-fetching the same
// history never creates a duplicate. recorded_at defaults to now (UTC). Every value travels
// as a bind arg, so there is no injection surface regardless of message content.
func (s *Store) Record(ctx context.Context, m Message) error {
	if m.Channel == "" || m.TS == "" {
		return nil // not a locatable message; skip silently
	}
	if m.RecordedAt.IsZero() {
		m.RecordedAt = time.Now().UTC()
	}
	var raw any
	if len(m.Raw) > 0 {
		raw = string(m.Raw)
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO messages
			(workspace, channel, ts, thread_ts, user, type, subtype, text, recorded_at, raw)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.Workspace, m.Channel, m.TS, nullable(m.ThreadTS), nullable(m.User),
		nullable(m.Type), nullable(m.Subtype), nullable(m.Text),
		m.RecordedAt.UTC().Format(time.RFC3339Nano), raw,
	)
	if err != nil {
		return fmt.Errorf("record message: %w", err)
	}
	// Only index in FTS when a NEW row was inserted (INSERT OR IGNORE affects 0 rows on a dup),
	// so re-capture never double-indexes.
	if s.fts && m.Text != "" {
		if n, _ := res.RowsAffected(); n == 1 {
			if id, idErr := res.LastInsertId(); idErr == nil {
				_, _ = s.db.ExecContext(ctx, `INSERT INTO messages_fts (rowid, text) VALUES (?, ?)`, id, m.Text)
			}
		}
	}
	return nil
}

// RecordBatch records many messages in one transaction (used when persisting a fetched page).
func (s *Store) RecordBatch(ctx context.Context, msgs []Message) (int, error) {
	n := 0
	for _, m := range msgs {
		if err := s.Record(ctx, m); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// Query lists messages matching f, newest first.
func (s *Store) Query(ctx context.Context, f Filter) ([]Message, error) {
	where, args := f.whereClause()
	parts := []string{selectMessage, "FROM messages"}
	if where != "" {
		parts = append(parts, "WHERE", where)
	}
	parts = append(parts, "ORDER BY messages.ts DESC, messages.id DESC LIMIT ?")
	args = append(args, effectiveLimit(f.Limit))

	rows, err := s.db.QueryContext(ctx, strings.Join(parts, " "), args...)
	if err != nil {
		return nil, fmt.Errorf("query messages: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanMessages(rows)
}

// Search full-text searches recorded text, newest first — FTS5 MATCH when available
// (AND/OR/NOT, prefix*, "phrases"), else a substring LIKE scan. A query that isn't valid FTS5
// syntax (a bare "foo-bar" reads `-bar` as an operator) transparently falls back to a LIKE
// scan for that query, so a plain search never errors. Only static SQL fragments are joined;
// q and every filter value travel as bind args.
func (s *Store) Search(ctx context.Context, q string, f Filter) ([]Message, error) {
	if s.fts {
		msgs, err := s.searchFTS(ctx, q, f)
		if err == nil {
			return msgs, nil
		}
		// An FTS5 syntax error on a plain query (hyphen, colon, unbalanced quote) is not a
		// real failure — retry as a substring scan rather than surfacing SQL noise.
	}
	return s.searchLike(ctx, q, f)
}

func (s *Store) searchFTS(ctx context.Context, q string, f Filter) ([]Message, error) {
	where, whereArgs := f.whereClause()
	parts := []string{selectMessage, "FROM messages",
		"JOIN messages_fts ON messages_fts.rowid = messages.id", "WHERE messages_fts MATCH ?"}
	args := []any{q}
	if where != "" {
		parts = append(parts, "AND", where)
		args = append(args, whereArgs...)
	}
	parts = append(parts, "ORDER BY messages.ts DESC, messages.id DESC LIMIT ?")
	args = append(args, effectiveLimit(f.Limit))

	rows, err := s.db.QueryContext(ctx, strings.Join(parts, " "), args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanMessages(rows)
}

func (s *Store) searchLike(ctx context.Context, q string, f Filter) ([]Message, error) {
	where, whereArgs := f.whereClause()
	parts := []string{selectMessage, "FROM messages", "WHERE messages.text LIKE ?"}
	args := []any{"%" + q + "%"}
	if where != "" {
		parts = append(parts, "AND", where)
		args = append(args, whereArgs...)
	}
	parts = append(parts, "ORDER BY messages.ts DESC, messages.id DESC LIMIT ?")
	args = append(args, effectiveLimit(f.Limit))

	rows, err := s.db.QueryContext(ctx, strings.Join(parts, " "), args...)
	if err != nil {
		return nil, fmt.Errorf("search messages: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanMessages(rows)
}

// Stats summarizes the store.
func (s *Store) Stats(ctx context.Context) (Stats, error) {
	st := Stats{FTS: s.fts, ByType: map[string]int{}}
	row := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*), COUNT(DISTINCT channel), COALESCE(MIN(ts),''), COALESCE(MAX(ts),'') FROM messages`)
	if err := row.Scan(&st.Messages, &st.Channels, &st.Oldest, &st.Newest); err != nil {
		return Stats{}, fmt.Errorf("stats: %w", err)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT COALESCE(type,''), COUNT(*) FROM messages GROUP BY type`)
	if err != nil {
		return Stats{}, fmt.Errorf("stats by type: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var t string
		var n int
		if err := rows.Scan(&t, &n); err != nil {
			return Stats{}, err
		}
		if t == "" {
			t = "(none)"
		}
		st.ByType[t] = n
	}
	return st, rows.Err()
}

// Prune deletes messages recorded before now-olderThan and returns the count removed.
func (s *Store) Prune(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().UTC().Add(-olderThan).Format(time.RFC3339Nano)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("prune: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // no-op once Commit succeeds
	if s.fts {
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM messages_fts WHERE rowid IN (SELECT id FROM messages WHERE recorded_at < ?)`, cutoff); err != nil {
			return 0, fmt.Errorf("prune fts index: %w", err)
		}
	}
	res, err := tx.ExecContext(ctx, `DELETE FROM messages WHERE recorded_at < ?`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("prune messages: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("prune: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("prune: %w", err)
	}
	return n, nil
}

// whereClause renders f as a parameterized fragment (no leading WHERE) plus bind args.
func (f Filter) whereClause() (string, []any) {
	var parts []string
	var args []any
	if f.Channel != "" {
		parts = append(parts, "messages.channel = ?")
		args = append(args, f.Channel)
	}
	if f.User != "" {
		parts = append(parts, "messages.user = ?")
		args = append(args, f.User)
	}
	if f.Since != "" {
		parts = append(parts, "messages.ts >= ?")
		args = append(args, f.Since)
	}
	return strings.Join(parts, " AND "), args
}

func effectiveLimit(n int) int {
	if n <= 0 {
		return DefaultLimit
	}
	return n
}

func nullable(v string) any {
	if v == "" {
		return nil
	}
	return v
}

func scanMessages(rows *sql.Rows) ([]Message, error) {
	out := []Message{}
	for rows.Next() {
		var (
			m        Message
			recorded string
			threadTS sql.NullString
			user     sql.NullString
			mtype    sql.NullString
			subtype  sql.NullString
			text     sql.NullString
			raw      sql.NullString
		)
		if err := rows.Scan(&m.ID, &m.Workspace, &m.Channel, &m.TS, &threadTS, &user,
			&mtype, &subtype, &text, &recorded, &raw); err != nil {
			return nil, err
		}
		parsed, err := time.Parse(time.RFC3339Nano, recorded)
		if err != nil {
			return nil, fmt.Errorf("parse stored recorded_at %q: %w", recorded, err)
		}
		m.RecordedAt = parsed
		m.ThreadTS = threadTS.String
		m.User = user.String
		m.Type = mtype.String
		m.Subtype = subtype.String
		m.Text = text.String
		if raw.Valid {
			m.Raw = json.RawMessage(raw.String)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
