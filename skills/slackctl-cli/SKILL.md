---
name: slackctl-cli
description: Operate Slack from the terminal with the `slackctl` CLI — list/read/manage conversations, post/edit/schedule/template messages, export a channel to JSONL, search (user or bot token), users & usergroups, set your status/presence/DND, upload/download files, canvases, reactions, pins, bookmarks, saved items, and stream events live (Socket Mode or RTM) with `slackctl listen`. Use whenever the user wants to read or send Slack messages, inspect channels/DMs/threads, check unreads, upload/download files, react/pin/bookmark/save, look up users, set a Slack status, manage @mention groups, or watch/replay Slack events from a script. Prefer it over raw curl to the Slack Web API.
version: 0.2.0
homepage: https://github.com/jjuanrivvera/slackctl
license: MIT
allowed-tools: Bash(slackctl:*)
metadata: {"openclaw":{"category":"messaging","emoji":"💬","requires":{"bins":["slackctl"],"env":["SLACK_BOT_TOKEN"]},"install":[{"kind":"go","package":"github.com/jjuanrivvera/slackctl/cmd/slackctl@latest","bins":["slackctl"]}]}}
---

# slackctl — Slack Web API CLI

## Prerequisites

- `slackctl` on PATH (`go install github.com/jjuanrivvera/slackctl/cmd/slackctl@latest`).
- A configured workspace: `slackctl auth status` must succeed. If it doesn't, the human
  runs `slackctl init` (tokens go to the OS keyring; `SLACK_BOT_TOKEN` also works).
- `search` and `saved` need **user-grade** creds — a user token (`auth login --kind user`)
  or a browser-session pair.
- No Slack app? A **browser web-client session** pair works for everything except Socket
  Mode: `auth login --kind session` (xoxc token + xoxd cookie), or `SLACK_XOXC_TOKEN` +
  `SLACK_XOXD_TOKEN`.
- `listen` picks its transport automatically: **Socket Mode** with an app-level token
  (`auth login --kind app`), else **RTM** with a user/session token (legacy but needs no
  app). Force with `--transport socket|rtm`.

## Why the CLI (not raw curl)

slackctl already handles auth headers, Slack's `ok:false` error envelope (with actionable
hints), cursor pagination (`--all`/`--limit`), rate-limit backoff with Retry-After, and
consistent output. `--dry-run` prints the exact curl when you need to inspect a request.

## Golden rules

1. **Discover before acting**: resolve channel ids with `conversations list` / `users
   search` — never guess a `C…`/`U…`/`D…` id.
2. **Destructive ops need explicit human intent**: `msg delete`, `conversations
   archive|kick|leave`, `usergroups disable`, `files delete`, `bookmarks remove`, `canvases delete`. Never run them speculatively.
3. Use `-o json` + `--jq` for parsing; keep `table` only for human display.
4. Paginated lists default to 100 items — pass `--all` when completeness matters.
5. Posting: keep `--text` under 4,000 chars; reply in threads with `--thread-ts`.

## Workflow: auth → discover → act → verify

```sh
slackctl auth status                                   # who am I / which workspace
slackctl conversations list --types public_channel,private_channel -o json
slackctl msg post --channel C0123 --text "hello"       # act
slackctl conversations history --channel C0123 --limit 5   # verify
```

## Cheatsheet

```sh
# Conversations
slackctl conversations list --all -o json
slackctl conversations history --channel C0123 --limit 50
slackctl conversations replies --channel C0123 --ts 1720000000.000100
slackctl conversations unreads --as-user
slackctl conversations mark --channel C0123 --ts 1720000000.000100

# Messages
slackctl msg post --channel C0123 --text "deploy done ✅"
slackctl msg post --channel C0123 --thread-ts 1720000000.000100 --text "in thread"
slackctl msg update --channel C0123 --ts 1720000000.000200 --text "fixed"
slackctl msg schedule --channel C0123 --post-at 1735689600 --text "future"
slackctl msg template --channel C0123 --file alert.tmpl --set service=api --set status=down

# Export / archive a channel to JSONL
slackctl conversations export --channel C0123 --threads > history.jsonl

# Search / users / status / groups
slackctl search messages --query "incident in:#eng-alerts" --sort timestamp   # user token
slackctl assistant search-context --query "incident"    # bot token OK
slackctl users lookup-email --email ada@example.com
slackctl users set-status --text "In a meeting" --emoji :calendar:            # user token
slackctl dnd set-snooze --minutes 60                                          # user token
slackctl usergroups members --usergroup S0123 -o id

# Files / canvases
slackctl files upload --file report.pdf --channels C0123 --comment "Q3"
slackctl files download --file F0123 --out ./report.pdf
slackctl canvases create --title Runbook --content '{"type":"markdown","markdown":"# Runbook"}'

# Reactions / pins / bookmarks / saved
slackctl reactions add --channel C0123 --ts 1720000000.000100 --name thumbsup
slackctl pins add --channel C0123 --ts 1720000000.000100
slackctl bookmarks add --channel C0123 --title Runbook --link https://wiki/runbook
slackctl saved list                  # user token; legacy stars API

# Live events (auto transport: Socket Mode with app token, else RTM with user/session)
slackctl listen --dms --json | jq -r '.text'
slackctl listen --channels C0123 --since 1h --json     # replay recent history, then go live

# Local history search (offline; recorded automatically as you use slackctl)
slackctl log search "deploy failed"
slackctl log --channel C0123 --since 24h -o json
slackctl log stats

# Escape hatch for unwrapped methods
slackctl api conversations.info -q channel=C0123 --idempotent
```

## Troubleshooting

- `not_authed` / `invalid_auth` → `slackctl auth login` (check the workspace with
  `--workspace`).
- `missing_scope` → the error names the needed scope; add it under OAuth & Permissions
  and reinstall the app.
- `not_allowed_token_type` → the method is user-token-only: `auth login --kind user`,
  then `--as-user` (search/saved use the user token automatically).
- `not_in_channel` → `slackctl conversations join --channel <id>` (public) or ask a human
  to /invite the bot (private).
- `listen` exits immediately → missing app token or Socket Mode disabled in the app
  config; also confirm event subscriptions (message.im, message.channels, …).
- Anything else: re-run with `-v` (raw responses) or `--dry-run` (exact request), and
  `slackctl doctor` for the full picture.
