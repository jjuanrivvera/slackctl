# slackctl

A fast, scriptable command-line tool for the [Slack Web API](https://docs.slack.dev/apis/web-api/)
— `gh`-style ergonomics, structured output, and a live event stream.

```console
$ slackctl conversations list
ID           NAME        IS_PRIVATE  IS_ARCHIVED  NUM_MEMBERS
C0123ABCD    general     false       false        42
C0456EFGH    eng-alerts  false       false        17

$ slackctl msg post --channel C0456EFGH --text "deploy finished ✅"
$ slackctl listen --dms --json | jq -r .text
```

## What it does

- **Conversations & messaging** — list/inspect channels, DMs and threads, read history,
  post/edit/schedule/delete messages, mark as read, export a channel to JSONL, post from a
  template.
- **Search, people & groups** — full-text search (user or bot token), user lookup, set your
  status/presence/DND, usergroup management.
- **Files, canvases, reactions, pins, bookmarks & saved items** — upload/download files,
  create canvases, and the usual channel adornments.
- **A live event stream** (`slackctl listen`) over **Socket Mode** *or* **RTM**, auto-selected
  by the credential you have.
- **Structured output** — table / json / yaml / csv, `--columns`, a built-in `--jq` filter,
  and cursor pagination (`--all`).
- **Safe by design** — tokens in the OS keyring, `--dry-run` prints the exact `curl`, an MCP
  server and an `agent guard` so AI agents can drive Slack without touching the dangerous verbs.

## Install

```sh
go install github.com/jjuanrivvera/slackctl/cmd/slackctl@latest
# or: brew install jjuanrivvera/slackctl/slackctl-cli
```

## First steps

```sh
slackctl auth login          # store a bot token — or --kind session for xoxc/xoxd
slackctl auth status         # who am I?
slackctl conversations list
slackctl msg post --channel C0123456 --text "hello from slackctl"
```

## Where to next

- **[Getting started](getting-started.md)** — install, authenticate, first commands.
- **[Authentication & tokens](authentication.md)** — bot/user/app/session tokens, workspaces,
  the keyring, and driving slackctl with a browser session (no Slack app).
- **[Output & filtering](output.md)** — formats, `--columns`, `--jq`, pagination.
- **[The `listen` command](listen.md)** — live event streaming over Socket Mode or RTM.
- **[AI agents](agents.md)** — the MCP server and the agent guard.
- **[Command reference](commands/slackctl.md)** — every command and flag.
