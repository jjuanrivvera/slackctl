## slackctl users search

Find users by name, display name, or email (client-side)

### Synopsis

Search users by substring across name, real_name, display_name, and email.
Slack has no public users.search method for bot tokens, so this walks users.list and
filters locally — on very large workspaces prefer 'users lookup-email' when you have
the exact address.

```
slackctl users search <query> [flags]
```

### Examples

```
  slackctl users search ada
  slackctl users search "@example.com" -o json
```

### Options

```
  -h, --help   help for search
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

* [slackctl users](slackctl_users.md)	 - Look up and list workspace users

