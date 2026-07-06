## slackctl mcp vscode

Manage VSCode MCP servers

### Synopsis

Manage MCP server configuration for Visual Studio Code

### Options

```
  -h, --help   help for vscode
```

### Options inherited from parent commands

```
      --as-user                 use the stored user token (xoxp-) instead of the bot token
      --base-url string         Web API base URL (default https://slack.com/api)
      --columns strings         explicit, ordered table/csv columns
      --dry-run                 print the equivalent curl and make no request
      --jq string               gojq expression applied to the result before rendering
      --no-color                disable colored output
      --no-store slackctl log   do not record messages to the local history store (see slackctl log)
  -o, --output string           output format: table|json|yaml|csv|id (default "table")
      --quiet                   suppress notes on stderr
      --rps float               client-side requests-per-second cap (0 = default)
      --show-token              do not redact the token in --dry-run output
  -v, --verbose                 log raw API responses to stderr
      --workspace string        workspace to use: a named profile/credential (env SLACKCTL_WORKSPACE)
```

### SEE ALSO

* [slackctl mcp](slackctl_mcp.md)	 - MCP server management
* [slackctl mcp vscode disable](slackctl_mcp_vscode_disable.md)	 - Remove server from VSCode config
* [slackctl mcp vscode enable](slackctl_mcp_vscode_enable.md)	 - Add server to VSCode config
* [slackctl mcp vscode list](slackctl_mcp_vscode_list.md)	 - Show VSCode MCP servers

