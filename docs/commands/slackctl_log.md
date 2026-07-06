## slackctl log

Search your local Slack message history

### Synopsis

slackctl records the messages that flow through it — every post, every page of
history/replies you fetch, and (with 'listen') every streamed event — into a per-workspace
SQLite database. 'log' searches that store locally: instantly, offline, and without Slack's
user-token-only, rate-limited search API.

Recording is on by default; disable it for a call with --no-store, or globally by never
running message-bearing commands. This command only READS (it never records), so it works
regardless of --no-store.

```
slackctl log [flags]
```

### Examples

```
  slackctl log                                   # recent messages
  slackctl log --channel C0123456 --since 24h
  slackctl log search "deploy failed"
  slackctl log search "deploy* AND staging" --channel C0123456
  slackctl log stats
  slackctl log prune --older-than 2160h          # drop anything older than 90 days
```

### Options

```
      --channel string   filter by conversation id
  -h, --help             help for log
      --limit int        max rows to return (default 50)
      --since string     only messages at/after this time: a Slack ts, unix seconds, or a Go duration (24h)
      --user string      filter by user id
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
* [slackctl log path](slackctl_log_path.md)	 - Print the path to the local history database
* [slackctl log prune](slackctl_log_prune.md)	 - Delete recorded messages older than a duration
* [slackctl log search](slackctl_log_search.md)	 - Full-text search recorded message text
* [slackctl log stats](slackctl_log_stats.md)	 - Show what the local history holds

