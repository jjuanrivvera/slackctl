## slackctl conversations export

Export a conversation's full history to JSONL

### Synopsis

Walk a conversation's entire message history (paginating over every page) and write
each message as one JSON object per line. With --threads, each threaded message is followed
by its replies. Bound the range with --oldest/--latest (unix or Slack ts).

```
slackctl conversations export [flags]
```

### Examples

```
  slackctl conversations export --channel C0123456 > history.jsonl
  slackctl conversations export --channel C0123456 --threads --out backup.jsonl
  slackctl conversations export --channel C0123456 --oldest 1720000000.000000
```

### Options

```
      --channel string   conversation id to export (C…/D…/G…)
  -h, --help             help for export
      --latest string    only messages before this ts
      --oldest string    only messages after this ts
      --out string       write to this file instead of stdout
      --threads          also export each thread's replies
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

