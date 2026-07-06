## slackctl msg post

Post a message to a conversation

### Synopsis

Post a message. Reply in a thread with --thread-ts; use --blocks for Block Kit JSON.
Slack recommends keeping --text under 4,000 characters (it truncates at 40,000).

```
slackctl msg post [flags]
```

### Examples

```
  slackctl msg post --channel C0123456 --text "deploy finished ✅"
  slackctl msg post --channel C0123456 --thread-ts 1720000000.000100 --text "replying in thread"
  slackctl msg post --channel C0123456 --blocks '[{"type":"section","text":{"type":"mrkdwn","text":"*hi*"}}]'
```

### Options

```
      --attachments string     legacy attachments (JSON array)
      --blocks string          Block Kit blocks (JSON array)
      --channel string         conversation id (C…/D…/G…)
  -h, --help                   help for post
      --icon-emoji string      override the bot's icon with an emoji (:tada:)
      --icon-url string        override the bot's icon with an image URL
      --markdown-text string   full-markdown message body (alternative to --text)
      --metadata string        message metadata (JSON object)
      --reply-broadcast        also show the thread reply in the channel
      --text string            message text (mrkdwn by default)
      --thread-ts string       parent ts — reply in that thread
      --unfurl-links           unfurl text-based URLs
      --unfurl-media           unfurl media URLs
      --username string        override the bot's display name
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

* [slackctl msg](slackctl_msg.md)	 - Post, edit, delete, and schedule messages

