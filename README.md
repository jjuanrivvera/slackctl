# slackctl

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
| User `xoxp-…` | OAuth & Permissions → User OAuth Token | `search`, `saved` (user-only methods), `--as-user` |
| App-level `xapp-…` | Basic Information → App-Level Tokens (scope `connections:write`) | `slackctl listen` (Socket Mode) |

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
               mark · open · close · unreads          (aliases: conv, channels)
msg            post · update · delete · ephemeral · me · permalink · schedule ·
               scheduled · delete-scheduled           (aliases: chat, message)
search         messages · files · all                 (user token)
users          list · info · lookup-email · conversations · presence · profile · search
usergroups     list · create · update · enable · disable · members · members-update
reactions      add · remove · get · list
saved          list · add · remove                    (stars API; user token)
pins           list · add · remove
emoji          list          team    info · profile
listen         Socket Mode event stream (see below)
auth · config · init · doctor · completion · alias · api · version · mcp · agent
```

Every command supports `-o table|json|yaml|csv|id`, `--jq <expr>`, `--columns`,
`--dry-run` (prints the equivalent curl, token redacted), `--all`/`--limit` on paginated
lists, and `--as-user` to run with the stored user token.

```sh
slackctl conversations history --channel C0123 --limit 50 -o json
slackctl users search ada
slackctl search messages --query "deploy failed in:#eng-alerts" --sort timestamp
slackctl conversations unreads --as-user --types im
slackctl api conversations.info -q channel=C0123 --idempotent   # raw escape hatch
```

## `slackctl listen` — Socket Mode stream

Streams events (messages, reactions, …) as they happen, one line each — built for pipes.
Requires the app-level token, Socket Mode enabled, and event subscriptions configured
(`message.im`, `message.channels`, `reaction_added`, …). Slack only delivers message
events for conversations the bot is a member of.

```sh
slackctl listen --dms --json                          # DM events as NDJSON
slackctl listen --dms --channels C0123,C0456 --json   # DMs OR those channels
slackctl listen --events message,reaction_added       # filter by event type
slackctl listen --raw | jq .payload.event.type        # full envelopes
```

Envelopes are acknowledged to Slack immediately (before filtering), reconnects fetch a
fresh URL with jittered backoff, and Ctrl-C shuts down cleanly.

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
