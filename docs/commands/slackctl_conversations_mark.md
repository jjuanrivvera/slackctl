## slackctl conversations mark

Move the read cursor in a conversation

### Synopsis

Sets the read cursor for whoever owns the token: the bot's cursor with the bot token, yours with --as-user.

```
slackctl conversations mark [flags]
```

### Examples

```
  slackctl conversations mark --channel C0123456 --ts 1720000000.000100
```

### Options

```
      --channel string   conversation id
  -h, --help             help for mark
      --ts string        timestamp of the most recently seen message
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

