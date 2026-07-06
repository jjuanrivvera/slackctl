## slackctl mcp claude

Manage Claude Desktop MCP servers

### Synopsis

Manage MCP server configuration for Claude Desktop

### Options

```
  -h, --help   help for claude
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

* [slackctl mcp](slackctl_mcp.md)	 - MCP server management
* [slackctl mcp claude disable](slackctl_mcp_claude_disable.md)	 - Remove server from Claude config
* [slackctl mcp claude enable](slackctl_mcp_claude_enable.md)	 - Add server to Claude config
* [slackctl mcp claude list](slackctl_mcp_claude_list.md)	 - Show Claude MCP servers

