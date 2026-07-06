## slackctl listen

Stream events over Socket Mode (DMs, channels) as lines

### Synopsis

Open a Socket Mode connection and stream events as they happen, one line each —
human-readable by default, JSON with --json (for pipes), full envelopes with --raw.

Filters combine as a union: --dms OR --channels; --events then narrows by event type.
With no filters, every subscribed event streams. Every envelope is acknowledged to Slack
immediately (before filtering), so filtered-out events are consumed, not redelivered.

Requires an app-level token (xapp-) with connections:write — store it with
'slackctl auth login --kind app' — plus Socket Mode enabled and the event subscriptions
(message.im, message.channels, reaction_added, …) configured for the app. Slack only
sends message events for conversations the bot is a member of. Runs until Ctrl-C.

```
slackctl listen [flags]
```

### Examples

```
  slackctl listen --dms --json                       # stream DM events as NDJSON
  slackctl listen --dms --channels C0123456,C0456789 --json
  slackctl listen --events message,reaction_added
  slackctl listen --raw | jq .payload.event.type
```

### Options

```
      --channels strings   only events from these conversation ids
      --debug-reconnects   ask Slack to rotate the connection every ~360s (tests reconnect handling)
      --dms                only events from direct messages
      --events strings     only these event types (message,reaction_added,…)
  -h, --help               help for listen
      --json               emit each event as one JSON line (NDJSON)
      --raw                emit full Socket Mode envelopes as NDJSON (includes slash commands and interactivity)
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

