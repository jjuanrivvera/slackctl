## slackctl team info

Show the workspace's name, domain, and icon

```
slackctl team info [flags]
```

### Examples

```
  slackctl team info
```

### Options

```
      --domain string   look up a team by its slack.com subdomain
  -h, --help            help for info
      --team string     team id (org tokens; default: the token's team)
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

* [slackctl team](slackctl_team.md)	 - Workspace information

