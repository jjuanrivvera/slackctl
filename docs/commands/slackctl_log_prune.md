## slackctl log prune

Delete recorded messages older than a duration

```
slackctl log prune [flags]
```

### Examples

```
  slackctl log prune --older-than 2160h   # 90 days
  slackctl log prune --older-than 720h    # 30 days
```

### Options

```
  -h, --help                help for prune
      --older-than string   delete messages recorded before now minus this Go duration (required)
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

* [slackctl log](slackctl_log.md)	 - Search your local Slack message history

