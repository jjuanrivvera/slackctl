## slackctl conversations history

Fetch a conversation's message history

```
slackctl conversations history [flags]
```

### Examples

```
  slackctl conversations history --channel C0123456 --limit 20
  slackctl conversations history --channel C0123456 --oldest 1720000000.000000 --all -o json
```

### Options

```
      --all                    fetch every page (overrides --limit)
      --channel string         conversation id
  -h, --help                   help for history
      --include-all-metadata   include message metadata
      --inclusive              include the oldest/latest ts message itself
      --latest string          only messages before this ts
      --limit int              max items to fetch across pages (default 100)
      --oldest string          only messages after this ts
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

