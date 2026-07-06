## slackctl conversations list

List conversations the token can see

```
slackctl conversations list [flags]
```

### Examples

```
  slackctl conversations list
  slackctl conversations list --types public_channel,private_channel --all
  slackctl conversations list --types im -o json
```

### Options

```
      --all                fetch every page (overrides --limit)
      --exclude-archived   omit archived channels
  -h, --help               help for list
      --limit int          max items to fetch across pages (default 100)
      --types string       comma-separated: public_channel,private_channel,mpim,im (default "public_channel")
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

