## slackctl conversations open

Open (or resume) a DM or group DM

```
slackctl conversations open [flags]
```

### Examples

```
  slackctl conversations open --users U111
  slackctl conversations open --users U111,U222,U333
```

### Options

```
      --channel string   resume an existing im/mpim by id instead
  -h, --help             help for open
      --return-im        include the full im object in the reply
      --users strings    user ids: one for a DM, several for a group DM
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

* [slackctl conversations](slackctl_conversations.md)	 - Manage channels, DMs, and group conversations

