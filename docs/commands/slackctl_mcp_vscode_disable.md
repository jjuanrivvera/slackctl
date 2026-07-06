## slackctl mcp vscode disable

Remove server from VSCode config

### Synopsis

Remove this application from VSCode MCP servers

```
slackctl mcp vscode disable [flags]
```

### Options

```
      --config-path string   Path to VSCode config file
  -h, --help                 help for disable
      --server-name string   Name of the MCP server to remove (default: derived from executable name)
      --workspace            Remove from workspace settings (.vscode/mcp.json) instead of user settings
```

### Options inherited from parent commands

```
      --as-user           use the stored user token (xoxp-) instead of the bot token
      --base-url string   Web API base URL (default https://slack.com/api)
      --columns strings   explicit, ordered table/csv columns
      --dry-run           print the equivalent curl and make no request
      --jq string         gojq expression applied to the result before rendering
      --no-color          disable colored output
  -o, --output string     output format: table|json|yaml|csv|id (default "table")
      --quiet             suppress notes on stderr
      --rps float         client-side requests-per-second cap (0 = default)
      --show-token        do not redact the token in --dry-run output
  -v, --verbose           log raw API responses to stderr
```

### SEE ALSO

* [slackctl mcp vscode](slackctl_mcp_vscode.md)	 - Manage VSCode MCP servers

