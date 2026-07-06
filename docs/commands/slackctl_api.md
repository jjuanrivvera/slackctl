## slackctl api

Call any Web API method directly (raw escape hatch)

### Synopsis

Invoke an arbitrary Web API method with a JSON body and/or key=value parameters.

This is the documented escape hatch for methods slackctl does not wrap as first-class
commands. It honors --dry-run and -o/--output like every other command. By default a
raw call is treated as a write (a form-encoded POST, never auto-retried); pass
--idempotent for read-only methods so they go as GETs and transient failures retry
safely.

```
slackctl api <method> [-d body] [-q key=value ...] [flags]
```

### Examples

```
  slackctl api auth.test --idempotent
  slackctl api chat.postMessage -q channel=C0123456 -q text="hi from slackctl"
  slackctl api conversations.info -q channel=C0123456 --idempotent
  slackctl api chat.postMessage -d '{"channel":"C0123456","text":"json body"}'
```

### Options

```
  -d, --data string         raw JSON request body
  -h, --help                help for api
      --idempotent          treat as read-only (safe to auto-retry)
  -q, --query stringArray   key=value parameter (repeatable)
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

