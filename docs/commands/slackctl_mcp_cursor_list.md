## slackctl mcp cursor list

Show Cursor MCP servers

### Synopsis

Show all MCP servers configured in Cursor

```
slackctl mcp cursor list [flags]
```

### Options

```
      --config-path string   Path to Cursor config file
  -h, --help                 help for list
      --workspace            List from workspace settings (.cursor/mcp.json) instead of user settings
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

* [slackctl mcp cursor](slackctl_mcp_cursor.md)	 - Manage Cursor MCP servers

