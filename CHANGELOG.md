# Changelog

All notable changes to slackctl are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); versions follow SemVer.

## [Unreleased]

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
