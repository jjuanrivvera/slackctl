## slackctl conversations replies

Fetch a thread's replies

```
slackctl conversations replies [flags]
```

### Examples

```
  slackctl conversations replies --channel C0123456 --ts 1720000000.000100
```

### Options

```
      --all              fetch every page (overrides --limit)
      --channel string   conversation id
  -h, --help             help for replies
      --inclusive        include the oldest/latest ts message itself
      --latest string    only replies before this ts
      --limit int        max items to fetch across pages (default 100)
      --oldest string    only replies after this ts
      --ts string        parent message ts (the thread root)
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

