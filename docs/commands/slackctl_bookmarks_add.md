## slackctl bookmarks add

Add a link bookmark to a channel

```
slackctl bookmarks add [flags]
```

### Examples

```
  slackctl bookmarks add --channel C0123456 --title "Runbook" --link https://wiki/runbook
```

### Options

```
      --channel string   conversation id
      --emoji string     emoji icon, e.g. :book:
  -h, --help             help for add
      --link string      URL to bookmark
      --title string     bookmark title
      --type string      bookmark type (link) (default "link")
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

* [slackctl bookmarks](slackctl_bookmarks.md)	 - Manage a channel's bookmarks

