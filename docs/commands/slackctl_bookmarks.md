## slackctl bookmarks

Manage a channel's bookmarks

### Options

```
  -h, --help   help for bookmarks
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

* [slackctl](slackctl.md)	 - Command-line tool for the Slack Web API
* [slackctl bookmarks add](slackctl_bookmarks_add.md)	 - Add a link bookmark to a channel
* [slackctl bookmarks edit](slackctl_bookmarks_edit.md)	 - Edit a channel bookmark
* [slackctl bookmarks list](slackctl_bookmarks_list.md)	 - List a channel's bookmarks
* [slackctl bookmarks remove](slackctl_bookmarks_remove.md)	 - Remove a channel bookmark

