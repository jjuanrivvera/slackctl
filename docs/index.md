# slackctl

A fast, scriptable command-line tool for the [Slack Web API](https://docs.slack.dev/apis/web-api/).

- Conversations, messages, search, users, usergroups, reactions, pins, saved items.
- A live event stream (`slackctl listen`) over **Socket Mode** *or* **RTM** — auto-selected
  by the credential you have.
- table / json / yaml / csv output, `--columns`, and a built-in `--jq` filter.
- OS-keyring token storage, named profiles for multiple workspaces.
- An MCP server and an `agent guard` so AI agents can drive Slack safely.

## Get started

```sh
go install github.com/jjuanrivvera/slackctl/cmd/slackctl@latest
slackctl auth login          # store a bot token (or --kind session for xoxc/xoxd)
slackctl auth status
slackctl conversations list
slackctl msg post --channel C0123456 --text "hello from slackctl"
```

## No Slack app? Use your browser session

If you already drive Slack with `slack-mcp-server` (a browser `xoxc` token + `xoxd` cookie),
point slackctl at the same credentials — no app to create:

```sh
slackctl auth login --kind session      # paste the xoxc token, then the xoxd cookie
slackctl listen --dms --json            # streams over RTM (no app needed)
```

Session credentials carry your own identity, so they back bot- and user-kind commands
(search and saved items included). The one thing they can't do is Socket Mode; `listen`
uses RTM instead.

See the full [command reference](commands/slackctl.md), or the
[README](https://github.com/jjuanrivvera/slackctl#readme) for installation options and the
RTM caveat.
