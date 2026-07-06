## slackctl canvases edit

Apply edit operations to a canvas

```
slackctl canvases edit [flags]
```

### Examples

```
  slackctl canvases edit --canvas F0123456 --changes '[{"operation":"insert_at_end","document_content":{"type":"markdown","markdown":"more"}}]'
```

### Options

```
      --canvas string    canvas id
      --changes string   JSON array of edit operations
  -h, --help             help for edit
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

* [slackctl canvases](slackctl_canvases.md)	 - Create and manage Canvases

