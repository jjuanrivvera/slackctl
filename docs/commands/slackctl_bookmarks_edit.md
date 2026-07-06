## slackctl bookmarks edit

Edit a channel bookmark

```
slackctl bookmarks edit [flags]
```

### Examples

```
  slackctl bookmarks edit --channel C0123456 --bookmark Bk123 --title "New Runbook"
```

### Options

```
      --bookmark string   bookmark id (Bk…)
      --channel string    conversation id
      --emoji string      new emoji icon
  -h, --help              help for edit
      --link string       new URL
      --title string      new title
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

* [slackctl bookmarks](slackctl_bookmarks.md)	 - Manage a channel's bookmarks

