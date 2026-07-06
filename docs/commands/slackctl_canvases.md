## slackctl canvases

Create and manage Canvases

### Options

```
  -h, --help   help for canvases
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
* [slackctl canvases access-delete](slackctl_canvases_access-delete.md)	 - Revoke access to a canvas
* [slackctl canvases access-set](slackctl_canvases_access-set.md)	 - Set who can read/edit a canvas
* [slackctl canvases create](slackctl_canvases_create.md)	 - Create a canvas
* [slackctl canvases delete](slackctl_canvases_delete.md)	 - Delete a canvas
* [slackctl canvases edit](slackctl_canvases_edit.md)	 - Apply edit operations to a canvas
* [slackctl canvases sections-lookup](slackctl_canvases_sections-lookup.md)	 - Find sections in a canvas by criteria

