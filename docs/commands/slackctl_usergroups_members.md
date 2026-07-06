## slackctl usergroups members

List a user group's members

```
slackctl usergroups members [flags]
```

### Examples

```
  slackctl usergroups members --usergroup S0123456 -o id
```

### Options

```
  -h, --help               help for members
      --include-disabled   include members of disabled groups
      --usergroup string   user group id (S…)
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

* [slackctl usergroups](slackctl_usergroups.md)	 - Manage user groups (@mention groups)

