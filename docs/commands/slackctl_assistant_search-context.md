## slackctl assistant search-context

Search messages and files (works with a bot token)

### Synopsis

Search across the workspace via the Assistant search API. Unlike 'slackctl search',
this accepts a bot token, so it works without a user/session credential.

```
slackctl assistant search-context [flags]
```

### Examples

```
  slackctl assistant search-context --query "deploy failed"
  slackctl assistant search-context --query "incident" --channel-types public_channel --limit 10
```

### Options

```
      --after string                only results after this date/ts
      --before string               only results before this date/ts
      --channel-types string        comma-separated: public_channel,private_channel,mpim,im
      --content-types string        comma-separated: messages,files
      --context-channel-id string   bias results toward this channel
  -h, --help                        help for search-context
      --include-bots                include messages from bots
      --limit int                   max results to return
      --query string                search query
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

* [slackctl assistant](slackctl_assistant.md)	 - Assistant APIs (bot-token search)

