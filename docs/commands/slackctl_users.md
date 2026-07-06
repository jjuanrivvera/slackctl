## slackctl users

Look up and list workspace users

### Options

```
  -h, --help   help for users
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
* [slackctl users conversations](slackctl_users_conversations.md)	 - List conversations a user is a member of
* [slackctl users info](slackctl_users_info.md)	 - Show one user
* [slackctl users list](slackctl_users_list.md)	 - List all workspace users
* [slackctl users lookup-email](slackctl_users_lookup-email.md)	 - Find a user by email address
* [slackctl users presence](slackctl_users_presence.md)	 - Show a user's presence (active/away)
* [slackctl users profile](slackctl_users_profile.md)	 - Show a user's profile fields
* [slackctl users search](slackctl_users_search.md)	 - Find users by name, display name, or email (client-side)

