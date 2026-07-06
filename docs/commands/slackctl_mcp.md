## slackctl mcp

MCP server management

### Synopsis

Manage MCP servers for AI assistants and code editors

### Options

```
  -h, --help   help for mcp
```

### Options inherited from parent commands

```
      --as-user            use the stored user token (xoxp-) instead of the bot token
      --base-url string    Web API base URL (default https://slack.com/api)
      --columns strings    explicit, ordered table/csv columns
      --dry-run            print the equivalent curl and make no request
      --jq string          gojq expression applied to the result before rendering
      --no-color           disable colored output
  -o, --output string      output format: table|json|yaml|csv|id (default "table")
      --quiet              suppress notes on stderr
      --rps float          client-side requests-per-second cap (0 = default)
      --show-token         do not redact the token in --dry-run output
  -v, --verbose            log raw API responses to stderr
      --workspace string   workspace to use: a named profile/credential (env SLACKCTL_WORKSPACE)
```

### SEE ALSO

* [slackctl](slackctl.md)	 - Command-line tool for the Slack Web API
* [slackctl mcp claude](slackctl_mcp_claude.md)	 - Manage Claude Desktop MCP servers
* [slackctl mcp cursor](slackctl_mcp_cursor.md)	 - Manage Cursor MCP servers
* [slackctl mcp start](slackctl_mcp_start.md)	 - Start the MCP server
* [slackctl mcp stream](slackctl_mcp_stream.md)	 - Stream the MCP server over HTTP
* [slackctl mcp tools](slackctl_mcp_tools.md)	 - Export tools as JSON
* [slackctl mcp vscode](slackctl_mcp_vscode.md)	 - Manage VSCode MCP servers

