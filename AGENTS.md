# AGENTS.md — working in the slackctl repo

`slackctl` is a command-line tool for the **Slack Web API** (plus a Socket Mode event
listener), built to the cliwright standard (Go + Cobra + GoReleaser). This file orients an
AI agent (or human) contributing.

## The one rule that matters
**`make verify` is the gate.** A change is done only when `make verify` exits `0`. It runs
`make check` (fmt, vet, golangci-lint, tests) + `spec-check` (the built surface matches
`api-manifest.json`) + `spec-completeness` (the manifest wraps the enumerated API or a
recorded waiver) + `cover-check` (≥80% coverage) + `dod-check.sh`. Run the full
`make verify` for any change that touches the command surface or a documented behavior —
not just `make check`.

## Architecture (where things live)
- `internal/api/` — the generic client core (bearer auth, retry, rate limit, dry-run curl,
  flexible JSON types, the typed `Call`/`CallInto`/`CallAllPages`, the `{ok:false}` envelope
  → typed APIError with hints). Written once; never copy-paste per method.
- `commands/` — thin, declarative command groups. Adding a Web API method is a few lines in
  a group file via `registerGroup` — **zero edits to shared code**. The generic builder
  stamps MCP read-only/write/destructive annotations from each command's `Kind`.
- `internal/socketmode/` — the hand-written Socket Mode client behind `slackctl listen`
  (apps.connections.open → wss dial → ack-within-3s → NDJSON events).
- `internal/{config,auth,output,version}` — workspace profiles + manual precedence (no
  Viper), keyring token storage (bot/user/app kinds), the table/json/yaml/csv renderer,
  build metadata.
- `cmd/slackctl/main.go` — entry point: `signal.NotifyContext` (Ctrl-C cancels in-flight
  work) + alias expansion before cobra parses.

## House rules
- Comments explain **WHY**, not WHAT.
- Thread `cmd.Context()` everywhere; never `context.Background()` (it breaks Ctrl-C). Tests
  use `t.Context()`.
- Secrets live in the OS keyring — never in config-in-repo, code, or commit messages.
- Pin every ambiguous API assumption in `DECISIONS.md`; read it back, never silently
  re-decide.
- The resource set is derived from the enumerated Web API surface (`api-manifest.json`), not
  hand-picked; the deliberate scope (core messaging/read surface + listen) is a recorded coverage-waiver.
- Success/failure is Slack's `ok` field, never the HTTP status; error hints key off Slack's
  string error codes (see `internal/api/errors.go`).
