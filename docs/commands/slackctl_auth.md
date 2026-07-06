## slackctl auth

Manage Slack tokens and verify authentication

### Synopsis

Capture, verify, and remove the tokens for a workspace profile. Tokens are stored in
your OS keyring, never in the config file.

A workspace profile can hold three token kinds:
  bot   xoxb-…  drives most commands (default)
  user  xoxp-…  unlocks user-only methods (search, saved items; use --as-user elsewhere)
  app   xapp-…  opens Socket Mode connections for 'slackctl listen'

### Options

```
  -h, --help   help for auth
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
* [slackctl auth login](slackctl_auth_login.md)	 - Store a Slack token and verify it
* [slackctl auth logout](slackctl_auth_logout.md)	 - Remove stored tokens for the active workspace
* [slackctl auth status](slackctl_auth_status.md)	 - Show the active workspace, base URL, and token validity

