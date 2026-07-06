## slackctl users conversations

List conversations a user is a member of

```
slackctl users conversations [flags]
```

### Examples

```
  slackctl users conversations
  slackctl users conversations --user U0123456 --types public_channel,private_channel
```

### Options

```
      --all                fetch every page (overrides --limit)
      --exclude-archived   omit archived channels
  -h, --help               help for conversations
      --limit int          max items to fetch across pages (default 100)
      --types string       comma-separated: public_channel,private_channel,mpim,im
      --user string        user id (default: the token's own user/bot)
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

