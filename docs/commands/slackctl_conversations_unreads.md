## slackctl conversations unreads

Show unread counts across the token's conversations

### Synopsis

List the conversations the token is a member of with their unread message counts
(relative to the token owner's read cursor — use --as-user for YOUR unreads rather
than the bot's). One conversations.info call is made per membership, so on very large
membership lists expect it to take a few seconds under Slack's rate limits.

```
slackctl conversations unreads [flags]
```

### Examples

```
  slackctl conversations unreads --as-user
  slackctl conversations unreads --types im,mpim
  slackctl conversations unreads --include-zero -o json
```

### Options

```
  -h, --help           help for unreads
      --include-zero   also list conversations with nothing unread
      --types string   conversation types to check: public_channel,private_channel,mpim,im
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

* [slackctl conversations](slackctl_conversations.md)	 - Manage channels, DMs, and group conversations

