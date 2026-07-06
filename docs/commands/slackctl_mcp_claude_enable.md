## slackctl mcp claude enable

Add server to Claude config

### Synopsis

Add this application as an MCP server in Claude Desktop

```
slackctl mcp claude enable [flags]
```

### Options

```
      --config-path string   Path to Claude config file
  -e, --env stringToString   Environment variables (e.g., --env KEY1=value1 --env KEY2=value2) (default [])
  -h, --help                 help for enable
      --log-level string     Log level (debug, info, warn, error)
      --server-name string   Name for the MCP server (default: derived from executable name)
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

* [slackctl mcp claude](slackctl_mcp_claude.md)	 - Manage Claude Desktop MCP servers

