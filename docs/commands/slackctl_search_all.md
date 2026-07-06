## slackctl search all

Search messages and files together

```
slackctl search all [flags]
```

### Examples

```
  slackctl search all --query "incident 42" -o json
```

### Options

```
      --count int         results per page (max 100)
  -h, --help              help for all
      --highlight         wrap matches in highlight markers
      --page int          page number
      --query string      search query (supports in:#chan, from:@user, "exact phrase")
      --sort string       sort by: score|timestamp
      --sort-dir string   sort direction: asc|desc
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

* [slackctl search](slackctl_search.md)	 - Search messages and files (needs a user token)

