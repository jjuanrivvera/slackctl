## slackctl reactions remove

Remove your reaction from a message

```
slackctl reactions remove [flags]
```

### Examples

```
  slackctl reactions remove --channel C0123456 --ts 1720000000.000100 --name thumbsup
```

### Options

```
      --channel string   conversation id
  -h, --help             help for remove
      --name string      emoji name without colons
      --ts string        message ts
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

