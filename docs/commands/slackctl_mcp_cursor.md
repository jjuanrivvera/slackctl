## slackctl mcp cursor

Manage Cursor MCP servers

### Synopsis

Manage MCP server configuration for Cursor

### Options

```
  -h, --help   help for cursor
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
* [slackctl mcp cursor disable](slackctl_mcp_cursor_disable.md)	 - Remove server from Cursor config
* [slackctl mcp cursor enable](slackctl_mcp_cursor_enable.md)	 - Add server to Cursor config
* [slackctl mcp cursor list](slackctl_mcp_cursor_list.md)	 - Show Cursor MCP servers

