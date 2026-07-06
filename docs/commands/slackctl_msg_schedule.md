## slackctl msg schedule

Schedule a message for later

```
slackctl msg schedule [flags]
```

### Examples

```
  slackctl msg schedule --channel C0123456 --post-at 1735689600 --text "happy new year"
```

### Options

```
      --blocks string      Block Kit blocks (JSON array)
      --channel string     conversation id
  -h, --help               help for schedule
      --post-at int        unix timestamp to post at
      --text string        message text
      --thread-ts string   schedule as a thread reply
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

