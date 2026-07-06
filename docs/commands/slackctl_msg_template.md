## slackctl msg template

Render a template file and post it

### Synopsis

Render a Go text/template file, substituting --set key=value variables, and post the
result to a conversation. Reference a variable as {{.name}}. With --blocks, the rendered
text is sent as a Block Kit JSON array instead of plain text.

```
slackctl msg template [flags]
```

### Examples

```
  slackctl msg template --channel C0123456 --file alert.tmpl --set service=api --set status=down
  slackctl msg template --channel C0123456 --file card.json.tmpl --set title=Deploy --blocks
```

### Options

```
      --blocks             send the rendered output as a Block Kit JSON array
      --channel string     conversation id to post to
      --file string        path to the template file
  -h, --help               help for template
      --set stringArray    template variable key=value (repeatable)
      --thread-ts string   post as a reply in this thread
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

* [slackctl msg](slackctl_msg.md)	 - Post, edit, delete, and schedule messages

