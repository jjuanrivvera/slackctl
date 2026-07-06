## slackctl config set

Set a per-profile option (key: base_url)

### Synopsis

Set a non-secret option on the active profile. Supported keys: base_url.

```
slackctl config set <key> <value> [flags]
```

### Examples

```
  slackctl config set base_url https://api.telegram.org
  slackctl --workspace acme config set base_url http://localhost:8081
```

### Options

```
  -h, --help   help for set
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

* [slackctl config](slackctl_config.md)	 - Inspect and edit slackctl configuration

