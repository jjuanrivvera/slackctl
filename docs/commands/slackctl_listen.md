## slackctl listen

Stream events live (Socket Mode or RTM) as lines

### Synopsis

Open a live connection and stream events as they happen, one line each —
human-readable by default, JSON with --json (for pipes), full frames with --raw.

Two transports, auto-selected by the credential you have (--transport to force):
  socket  Socket Mode — needs an app-level token (xapp-, connections:write). Official,
          robust, acks enveloped events. Store one with 'slackctl auth login --kind app'.
  rtm     Real Time Messaging — the legacy WebSocket that works with a user/session token
          (xoxp- or the xoxc+xoxd browser pair). No Slack app required — it streams with a
          browser web-client session. RTM is legacy and unofficial for xoxc tokens, so a
          workspace may block it.

Filters combine as a union: --dms OR --channels; --events then narrows by event type.
With no filters, every received event streams. (Socket Mode only delivers events your app
subscribed to and is a member of; RTM delivers everything the user account can see.)
Runs until Ctrl-C.

```
slackctl listen [flags]
```

### Examples

```
  slackctl listen --dms --json                       # auto transport, DM events as NDJSON
  slackctl listen --transport rtm --json             # force RTM (session/user token)
  slackctl listen --channels C0123456,C0456789 --json
  slackctl listen --events message,reaction_added
  slackctl listen --raw | jq .
```

### Options

```
      --channels strings   only events from these conversation ids
      --debug-reconnects   Socket Mode only: rotate the connection every ~360s (tests reconnect handling)
      --dms                only events from direct messages
      --events strings     only these event types (message,reaction_added,…)
  -h, --help               help for listen
      --json               emit each event as one JSON line (NDJSON)
      --raw                emit full wire frames as NDJSON (Socket Mode envelopes / RTM frames)
      --transport string   event transport: auto|socket|rtm (default "auto")
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

