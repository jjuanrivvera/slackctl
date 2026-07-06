## slackctl msg

Post, edit, delete, and schedule messages

### Options

```
  -h, --help   help for msg
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

* [slackctl](slackctl.md)	 - Command-line tool for the Slack Web API
* [slackctl msg delete](slackctl_msg_delete.md)	 - Delete a message
* [slackctl msg delete-scheduled](slackctl_msg_delete-scheduled.md)	 - Cancel a scheduled message
* [slackctl msg ephemeral](slackctl_msg_ephemeral.md)	 - Post an ephemeral message (visible to one user)
* [slackctl msg me](slackctl_msg_me.md)	 - Post a /me action message
* [slackctl msg permalink](slackctl_msg_permalink.md)	 - Get a message's permalink URL
* [slackctl msg post](slackctl_msg_post.md)	 - Post a message to a conversation
* [slackctl msg schedule](slackctl_msg_schedule.md)	 - Schedule a message for later
* [slackctl msg scheduled](slackctl_msg_scheduled.md)	 - List scheduled messages
* [slackctl msg update](slackctl_msg_update.md)	 - Edit an existing message

