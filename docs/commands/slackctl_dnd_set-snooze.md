## slackctl dnd set-snooze

Turn on snooze for a number of minutes

```
slackctl dnd set-snooze [flags]
```

### Examples

```
  slackctl dnd set-snooze --minutes 60
```

### Options

```
  -h, --help          help for set-snooze
      --minutes int   snooze duration in minutes
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

* [slackctl dnd](slackctl_dnd.md)	 - Do Not Disturb — snooze and status

