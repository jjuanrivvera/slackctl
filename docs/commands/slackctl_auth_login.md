## slackctl auth login

Store a Slack token and verify it

### Synopsis

Capture a token from your Slack app (https://api.slack.com/apps → OAuth & Permissions),
verify it against auth.test, and save it to the keyring for the active workspace profile.

```
slackctl auth login [flags]
```

### Examples

```
  slackctl auth login                          # prompt for the bot token (hidden input)
  slackctl auth login --token xoxb-...         # non-interactive
  slackctl auth login --kind user              # store a user token (search, saved items)
  slackctl auth login --kind app               # store an app-level token (slackctl listen)
  slackctl auth login --workspace acme         # store under a named workspace profile
```

### Options

```
  -h, --help           help for login
      --kind string    token kind: bot|user|app (default "bot")
      --no-verify      skip the auth.test verification call
      --token string   Slack token (omit to be prompted with hidden input)
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

