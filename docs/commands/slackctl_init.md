## slackctl init

First-run wizard: capture tokens, verify, and save a workspace profile

### Synopsis

Interactively set up a workspace profile: paste the bot token (xoxb-), verify it against
auth.test, and optionally add a user token (xoxp-, for search and saved items) and an
app-level token (xapp-, for 'slackctl listen'). Tokens go to the OS keyring.

Create an app and grab tokens at https://api.slack.com/apps (OAuth & Permissions for
xoxb/xoxp; Basic Information → App-Level Tokens for xapp).

```
slackctl init [flags]
```

### Examples

```
  slackctl init
  slackctl init --workspace acme
```

### Options

```
  -h, --help   help for init
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

