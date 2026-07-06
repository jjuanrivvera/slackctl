## slackctl msg update

Edit an existing message

```
slackctl msg update [flags]
```

### Examples

```
  slackctl msg update --channel C0123456 --ts 1720000000.000100 --text "fixed the typo"
```

### Options

```
      --attachments string     new legacy attachments (JSON array)
      --blocks string          new Block Kit blocks (JSON array)
      --channel string         conversation id
  -h, --help                   help for update
      --markdown-text string   new full-markdown body
      --text string            new message text
      --ts string              ts of the message to edit
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

