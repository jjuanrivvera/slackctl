## slackctl saved add

Save (star) a message, file, or channel

```
slackctl saved add [flags]
```

### Examples

```
  slackctl saved add --channel C0123456 --ts 1720000000.000100
```

### Options

```
      --channel string   conversation id (with --ts: the message; alone: the channel)
      --file string      file id to save
  -h, --help             help for add
      --ts string        message ts to save
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

* [slackctl saved](slackctl_saved.md)	 - Saved items (legacy stars API; needs a user token)

