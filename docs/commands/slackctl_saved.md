## slackctl saved

Saved items (legacy stars API; needs a user token)

### Synopsis

Manage saved items via the legacy stars.* API. Slack deprecated stars when it moved to
the "Later" tab, and Later has no public API yet — so this is still the only saved-items
surface. Items starred here may not appear in the Later tab. All commands need a user
token: slackctl auth login --kind user.

### Options

```
  -h, --help   help for saved
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
* [slackctl saved add](slackctl_saved_add.md)	 - Save (star) a message, file, or channel
* [slackctl saved list](slackctl_saved_list.md)	 - List saved (starred) items
* [slackctl saved remove](slackctl_saved_remove.md)	 - Remove a saved (starred) item

