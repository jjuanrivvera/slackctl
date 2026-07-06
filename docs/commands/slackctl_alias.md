## slackctl alias

Manage user-defined command aliases

### Synopsis

Define shorthand commands. Aliases are expanded before parsing and can never shadow a built-in.

### Options

```
  -h, --help   help for alias
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
* [slackctl alias list](slackctl_alias_list.md)	 - List aliases
* [slackctl alias remove](slackctl_alias_remove.md)	 - Remove an alias
* [slackctl alias set](slackctl_alias_set.md)	 - Create or update an alias

