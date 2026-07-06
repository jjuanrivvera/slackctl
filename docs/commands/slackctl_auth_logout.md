## slackctl auth logout

Remove stored tokens for the active workspace

```
slackctl auth logout [flags]
```

### Examples

```
  slackctl auth logout               # remove all tokens for the workspace
  slackctl auth logout --kind user   # remove only the user token
```

### Options

```
  -h, --help          help for logout
      --kind string   remove only this token kind: bot|user|app
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

* [slackctl auth](slackctl_auth.md)	 - Manage Slack tokens and verify authentication

