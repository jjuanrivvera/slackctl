# AI agents

slackctl is built to be driven by AI agents safely. Two features make that work: an **MCP
server** that exposes the commands as tools, and an **agent guard** that generates host safety
rules blocking the irreversible ones.

## MCP server

```sh
slackctl mcp start
```

Runs slackctl as a [Model Context Protocol](https://modelcontextprotocol.io) server. Each
command becomes a tool, annotated read-only / write / destructive so a host can gate writes.
The server reuses the active workspace and keyring, and **excludes** secret and instance flags
(`--show-token`, `--workspace`, `--as-user`, `--base-url`) plus the setup commands (`auth`,
`config`, `init`, `alias`) — an agent can't read your token or switch workspaces.

Install it into a host:

```sh
slackctl mcp claude      # Claude Desktop / Code
slackctl mcp cursor
slackctl mcp vscode
```

## Agent guard

```sh
slackctl agent guard --host claude-code
slackctl agent guard --host codex
slackctl agent guard --host opencode
```

Classifies every command from the live tree into read / write / irreversible, then emits host
safety config that:

- **hard-blocks irreversible operations** — `msg delete`, `conversations archive`/`kick`/
  `leave`, `usergroups disable` — across canonical **and alias** command paths;
- **gates ordinary writes** behind approval;
- **leaves reads free.**

For Claude Code it also emits a `PreToolUse` hook that resists quoting and path-prefix
obfuscation, and gates the raw `slackctl api` escape hatch to read-shaped methods only.

!!! note "What's guaranteed"
    MCP-only operation is the hard guarantee. The Bash hook is best-effort — it defeats
    quoting/path tricks but not variable indirection or shell aliases. For an airtight setup,
    run the agent against the MCP server (or a read-only sandbox) rather than a shell.

Regenerate the guard after upgrading slackctl so new commands are covered:

```sh
slackctl agent guard --host claude-code --out .claude/settings.json
```
