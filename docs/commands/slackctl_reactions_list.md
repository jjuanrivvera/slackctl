## slackctl reactions list

List items the token's user has reacted to

```
slackctl reactions list [flags]
```

### Examples

```
  slackctl reactions list --limit 20
```

### Options

```
      --all           fetch every page (overrides --limit)
      --full          return all reactions, not a summary
  -h, --help          help for list
      --limit int     max items to fetch across pages (default 100)
      --user string   list another user's reactions
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

* [slackctl reactions](slackctl_reactions.md)	 - Add, remove, and list emoji reactions

