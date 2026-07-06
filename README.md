# slackctl

📖 **Docs:** <https://jjuanrivvera.github.io/slackctl/>

A fast, scriptable command-line tool for the **Slack Web API** — `gh`-style ergonomics,
table/json/yaml/csv output, named profiles for multiple workspaces, OS-keyring token
storage, a Socket Mode event stream (`slackctl listen`), an MCP server so AI agents can
drive Slack safely, and an agent guard that hard-blocks the irreversible stuff.

```console
$ slackctl conversations list
ID           NAME        IS_PRIVATE  IS_ARCHIVED  NUM_MEMBERS
C0123ABCD    general     false       false        42
C0456EFGH    eng-alerts  false       false        17

$ slackctl msg post --channel C0456EFGH --text "deploy finished ✅"
$ slackctl listen --dms --json | jq -r .text
```

## Install

```sh
# From source (Go 1.24+)
go install github.com/jjuanrivvera/slackctl/cmd/slackctl@latest

# Or build locally
git clone https://github.com/jjuanrivvera/slackctl && cd slackctl && make build
```

Homebrew/Scoop/deb/rpm/apk packaging ships with tagged releases (GoReleaser config is in
the repo; releases are cut from `main`).

## Setup

slackctl talks to Slack as a **Slack app**. Create one at <https://api.slack.com/apps>
(from scratch), install it to your workspace, and grab tokens:

| Token | Where | Unlocks |
|---|---|---|
| Bot `xoxb-…` | OAuth & Permissions → Bot User OAuth Token | almost everything (default) |
| User `xoxp-…` | OAuth & Permissions → User OAuth Token | `search`, `saved` (user-only methods), `--as-user`, `listen` via RTM |
| App-level `xapp-…` | Basic Information → App-Level Tokens (scope `connections:write`) | `slackctl listen` via Socket Mode |
| Session `xoxc-…` + `xoxd-…` | your Slack web-client session (browser token + `d` cookie) | **everything except Socket Mode** — no Slack app needed; `listen` runs over RTM |

### No Slack app? Use your browser session

If you have a Slack **web-client session** — an `xoxc` token plus the paired `xoxd` cookie
(the credentials your browser uses) — slackctl can authenticate with it directly, no app to
create:

```sh
slackctl auth login --kind session      # paste the xoxc token, then the xoxd cookie
# or, env-only:
export SLACK_XOXC_TOKEN=xoxc-…
export SLACK_XOXD_TOKEN=xoxd-…
slackctl auth status                     # verifies as your user
slackctl conversations list              # reads/writes as you
slackctl listen --dms --json             # streams over RTM (no app needed)
```

Session creds carry your own identity, so they back bot- and user-kind commands (search and
saved items included). The one thing they can't do is Socket Mode; `listen` uses RTM instead
(see the caveat below).

Bot-token scopes to request, matched to what you'll use: `channels:read`, `groups:read`,
`im:read`, `mpim:read` (listing), `channels:history` + `groups:history` + `im:history` +
`mpim:history` (history/replies and message events), `chat:write` (posting),
`channels:manage`, `groups:write`, `im:write`, `mpim:write` (create/invite/topic/mark),
`channels:join`, `reactions:read`, `reactions:write`, `users:read`, `users:read.email`,
`usergroups:read`, `usergroups:write`, `pins:read`, `pins:write`, `emoji:read`,
`team:read`. For the user token: `search:read`, `stars:read`, `stars:write`.

Then:

```sh
slackctl init                      # wizard: bot token + optional user/app tokens
# or piecemeal:
slackctl auth login                # bot token → OS keyring
slackctl auth login --kind user    # user token (search, saved items)
slackctl auth login --kind app     # app-level token (listen)
slackctl auth status               # who am I?
slackctl doctor                    # full diagnostics
```

Tokens live in the OS keyring (macOS Keychain / Linux Secret Service / Windows Credential
Manager), never in config files. Env overrides: `SLACK_BOT_TOKEN`, `SLACK_USER_TOKEN`,
`SLACK_APP_TOKEN`, or `SLACKCTL_TOKEN` (explicit, any kind).

### Multiple workspaces

A profile is a workspace. `--workspace acme` (env `SLACKCTL_WORKSPACE`) selects one;
`slackctl config use acme` switches the default. `--profile` works as a hidden alias.

## Commands

```text
conversations  list · info · history · replies · members · create · rename · archive ·
               unarchive · invite · join · leave · kick · set-topic · set-purpose ·
               mark · open · close · unreads · export   (aliases: conv, channels)
msg            post · update · delete · ephemeral · me · permalink · schedule ·
               scheduled · delete-scheduled · template  (aliases: chat, message)
search         messages · files · all                   (user token)
assistant      search-context                           (bot-token search)
files          list · info · delete · upload · download (aliases: file)
canvases       create · edit · delete · access-set · access-delete · sections-lookup
users          list · info · lookup-email · conversations · presence · profile · search ·
               set-presence · set-status                (aliases: user)
usergroups     list · create · update · enable · disable · members · members-update
dnd            info · set-snooze · end-snooze · end-dnd · team-info
reactions      add · remove · get · list
pins           list · add · remove       bookmarks   list · add · edit · remove
saved          list · add · remove                      (stars API; user token)
emoji          list          team    info · profile
listen         live event stream: Socket Mode (app token) OR RTM (session/user)
log            search your local message history (list · search · stats · prune · path)
auth · config · init · doctor · completion · alias · api · version · mcp · agent
```

Every command supports `-o table|json|yaml|csv|id`, `--jq <expr>`, `--columns`,
`--dry-run` (prints the equivalent curl, token redacted), `--all`/`--limit` on paginated
lists, and `--as-user` to run with the stored user token.

```sh
slackctl conversations history --channel C0123 --limit 50 -o json
slackctl conversations export --channel C0123 --threads > history.jsonl   # archive a channel
slackctl files upload --file report.pdf --channels C0123 --comment "Q3"
slackctl users set-status --text "In a meeting" --emoji :calendar:
slackctl dnd set-snooze --minutes 60
slackctl msg template --channel C0123 --file alert.tmpl --set service=api --set status=down
slackctl assistant search-context --query "deploy failed"       # search with a bot token
slackctl api conversations.info -q channel=C0123 --idempotent   # raw escape hatch
```

## `slackctl listen` — live event stream

Streams events (messages, reactions, …) as they happen, one line each — built for pipes.
It has **two transports** and auto-selects by the credential you have, so it works whether
or not you created a Slack app:

| Transport | Needs | Notes |
|---|---|---|
| **RTM** (`--transport rtm`) | user token (`xoxp-`) or the session pair (`xoxc-`+`xoxd-`) | No Slack app required — streams with a browser web-client session. Legacy/unofficial for `xoxc`; a workspace may block it. |
| **Socket Mode** (`--transport socket`) | app-level token (`xapp-`, `connections:write`) + Socket Mode enabled + event subscriptions | Official and robust. Only delivers subscribed events for conversations the bot is in. |

`--transport auto` (default) picks Socket Mode when an app token is present, else RTM.

```sh
slackctl listen --dms --json                          # auto transport, DM events as NDJSON
slackctl listen --transport rtm --json                # force RTM (session/user token)
slackctl listen --dms --channels C0123,C0456 --json   # DMs OR those channels
slackctl listen --channels C0123 --since 1h --json    # replay the last hour, then go live
slackctl listen --events message,reaction_added       # filter by event type
slackctl listen --raw | jq .                           # full wire frames
```

Both transports feed one filter/render path, reconnect with a fresh URL on drop (RTM keeps
the socket alive with a periodic ping; Socket Mode acks each envelope before filtering), and
shut down cleanly on Ctrl-C.

> **RTM caveat:** Slack marks RTM as legacy and doesn't officially support it for `xoxc`
> session tokens — it works today but could be disabled, and some Enterprise Grid workspaces
> block it. For a durable listener, create a Slack app (no need to publish it), grab an
> app-level token, and use `--transport socket`.

## Local history & search

slackctl records the messages it sees — posts, fetched history/replies, and streamed `listen`
events — into a per-workspace SQLite database, so you can full-text search your Slack history
**offline and without Slack's user-token-only, rate-limited search API**.

```sh
slackctl listen --channels C0123 &          # mirror a channel in real time
slackctl log search "deploy failed"          # search it offline, instantly
slackctl log --channel C0123 --since 24h     # filter by channel + time
slackctl log stats                           # size + FTS mode        log path
slackctl log prune --older-than 2160h        # drop anything older than 90 days
```

Recording is on by default; disable it with `--no-store` (per call). The database holds
message text — it lives at `slackctl log path` with `0600` perms, is never uploaded, and `log`
is excluded from the MCP tool surface. See the [history guide](https://jjuanrivvera.github.io/slackctl/history/).

## AI agents

```sh
slackctl mcp start                            # run as an MCP server (annotated tools)
slackctl agent guard --host claude-code       # generate safety rails
slackctl agent guard --host codex             # read-only sandbox config
slackctl agent guard --host opencode          # permission map
```

The MCP server exposes each command as a tool with read-only/write/destructive
annotations and hides secret/instance flags (`--show-token`, `--workspace`, `--as-user`,
`--base-url`). The agent guard hard-blocks irreversible operations (message deletes,
archive/kick/leave, usergroup disable) across canonical **and alias** command paths, gates
the raw `api` escape hatch to read-shaped methods, and emits a PreToolUse hook that
resists quoting/path obfuscation. MCP-only operation is the hard guarantee; the Bash hook
is best-effort (documented limits: variable indirection, shell aliases).

## Honest limits & alternatives

- **Slack's official [CLI](https://tools.slack.dev/slack-cli/)** targets building/deploying
  Slack *apps* (manifests, functions, triggers) — it is not a Web-API data tool. If you're
  developing a Slack platform app, use it; slackctl is for driving the API from scripts.
- **`search` and `saved` need a user token** — Slack rejects bot tokens on those methods
  (`not_allowed_token_type`), and the stars API (the only public saved-items surface) is
  deprecated: items you star may not show in the Later tab.
- **Unreads are cursor-relative.** `conversations unreads` reports unreads for whoever owns
  the token; use `--as-user` for your own read state.
- **New commercially-distributed non-Marketplace apps** (created after May 2025) get
  `conversations.history`/`replies` limited to 1 req/min · 15 items. Internal apps — the
  normal slackctl case — keep standard Tier 3 limits.
- **No file upload yet**: Slack sunset `files.upload` (Nov 2025); the external-upload flow
  is on the roadmap.

## Development

`make verify` is the gate: fmt, vet, golangci-lint, tests (race on CI), coverage ≥80%,
`spec-check` (CLI surface ⊆ `api-manifest.json`), `spec-completeness` (manifest vs the
enumerated 308-method API), and the Definition-of-Done checks. See `AGENTS.md` and
`DECISIONS.md` for architecture and pinned assumptions.

## License

MIT
