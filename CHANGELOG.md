# Changelog

All notable changes to slackctl are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); versions follow SemVer.

## [0.3.0] - 2026-07-06

### Added
- **Local message history + search (`slackctl log`).** slackctl records the messages it sees
  — posts, fetched history/replies, and streamed `listen` events — into a per-workspace
  SQLite database (pure-Go, FTS5). Search it offline and instantly, without Slack's
  user-token-only, rate-limited search API:
  - `log` / `log search <query>` (FTS5 operators, with a graceful substring fallback),
    filterable by `--channel`/`--user`/`--since`.
  - `log stats`, `log path`, `log prune --older-than <dur>`.
  - Recording is on by default; disable with `--no-store`. The DB holds message text — it is
    per-workspace, `0600`, local-only, and `log` is excluded from the MCP tool surface.

## [0.2.1] - 2026-07-06

### Fixed
- An explicitly-empty string flag is now sent, so `conversations set-topic --topic ""`
  (and set-purpose / users set-status) can CLEAR a value instead of silently omitting it.
- `assistant search-context` gained the `--limit` flag its help already referenced.

## [0.2.0] - 2026-07-06

### Added
- `files` family: list/info/delete plus `upload` (the external-upload flow, since
  files.upload was sunset) and `download` (authed fetch of a private URL).
- `dnd` family (info/set-snooze/end-snooze/end-dnd/team-info) and `users set-status` /
  `users set-presence` for managing your own availability.
- `bookmarks` family (list/add/edit/remove).
- `assistant search-context` — full-text search that works with a **bot** token (unlike
  `search`, which is user-token-only).
- `canvases` family (create/edit/delete/access-set/access-delete/sections-lookup).
- `conversations export` — dump a channel's full history (and, with `--threads`, its
  replies) to JSONL.
- `msg template` — render a Go text/template with `--set key=value` variables and post it
  (`--blocks` for Block Kit).
- `listen --since` — replay recent `--channels` history before the live stream.

### Changed
- `listen` bounds each frame read with a timeout so a half-open connection reconnects
  instead of hanging.
- Clearer `token_expired` hint for rotating browser-session credentials.
- Manifest coverage rises to 86/308 enumerated methods (27%).

## [0.1.0] - 2026-07-06

### Added
- Full Slack Web API command surface: `conversations` (18 verbs incl. composite
  `unreads`), `msg` (post/update/delete/ephemeral/me/permalink/schedule), `search`
  (messages/files/all, user token), `users` (incl. client-side `search`), `usergroups`,
  `reactions`, `saved` (stars), `pins`, `emoji`, `team`.
- `slackctl listen`: hand-written Socket Mode stream — NDJSON events, `--dms`,
  `--channels`, `--events`, `--raw`, ack-before-handler, fresh-URL reconnect with
  full-jitter backoff.
- Workspace profiles with three token kinds (bot/user/app) in the OS keyring;
  `--as-user`; SLACK_BOT_TOKEN/SLACK_USER_TOKEN/SLACK_APP_TOKEN env support.
- Output formats table/json/yaml/csv/id with `--jq`, `--columns`, CSV
  formula-injection sanitization, deterministic column order.
- Resilient client: `ok`-envelope errors with actionable hints, cursor pagination
  walker (`--all`/`--limit`), Retry-After-aware retries, idempotent-only replay,
  fixed-RPS pacing with halve-on-429.
- Meta commands: auth login/logout/status, config, init wizard, doctor, completion,
  alias, raw `api` escape hatch, version.
- MCP server (`slackctl mcp`) with read-only/write/destructive tool annotations and
  `agent guard` (claude-code/codex/opencode) with an obfuscation-resistant PreToolUse
  hook.
- Docs: generated command reference, README, SECURITY, DECISIONS, Claude Code
  plugin + skill packaging.
