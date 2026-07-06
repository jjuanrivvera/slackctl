## slackctl usergroups

Manage user groups (@mention groups)

### Options

```
  -h, --help   help for usergroups
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
* [slackctl usergroups create](slackctl_usergroups_create.md)	 - Create a user group
* [slackctl usergroups disable](slackctl_usergroups_disable.md)	 - Disable a user group
* [slackctl usergroups enable](slackctl_usergroups_enable.md)	 - Re-enable a disabled user group
* [slackctl usergroups list](slackctl_usergroups_list.md)	 - List user groups
* [slackctl usergroups members](slackctl_usergroups_members.md)	 - List a user group's members
* [slackctl usergroups members-update](slackctl_usergroups_members-update.md)	 - Replace a user group's member list
* [slackctl usergroups update](slackctl_usergroups_update.md)	 - Update a user group's name/handle/description

