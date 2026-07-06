## slackctl reactions

Add, remove, and list emoji reactions

### Options

```
  -h, --help   help for reactions
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
* [slackctl reactions add](slackctl_reactions_add.md)	 - React to a message
* [slackctl reactions get](slackctl_reactions_get.md)	 - Show the reactions on a message
* [slackctl reactions list](slackctl_reactions_list.md)	 - List items the token's user has reacted to
* [slackctl reactions remove](slackctl_reactions_remove.md)	 - Remove your reaction from a message

