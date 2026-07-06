## slackctl dnd

Do Not Disturb — snooze and status

### Options

```
  -h, --help   help for dnd
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
* [slackctl dnd end-dnd](slackctl_dnd_end-dnd.md)	 - End the current Do Not Disturb session
* [slackctl dnd end-snooze](slackctl_dnd_end-snooze.md)	 - End the current snooze
* [slackctl dnd info](slackctl_dnd_info.md)	 - Show a user's Do Not Disturb status
* [slackctl dnd set-snooze](slackctl_dnd_set-snooze.md)	 - Turn on snooze for a number of minutes
* [slackctl dnd team-info](slackctl_dnd_team-info.md)	 - Show DND status for several users

