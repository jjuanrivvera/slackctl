## slackctl users set-status

Set your Slack status (text, emoji, expiry)

### Synopsis

Set the custom status on your own profile. Clear it by passing empty --text and
--emoji. Needs a user or session token (a bot has no personal status).

```
slackctl users set-status [flags]
```

### Examples

```
  slackctl users set-status --text "In a meeting" --emoji :calendar:
  slackctl users set-status --text "Lunch" --emoji :taco: --expiration 1735689600
  slackctl users set-status --text "" --emoji ""     # clear
```

### Options

```
      --emoji string     status emoji, e.g. :coffee:
      --expiration int   unix timestamp when the status clears (0 = never)
  -h, --help             help for set-status
      --text string      status text
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

* [slackctl users](slackctl_users.md)	 - Look up and list workspace users

