# The `listen` command

`slackctl listen` opens a live connection and streams Slack events as they happen — one line
each, built for pipes. It runs until Ctrl-C.

```sh
slackctl listen --dms --json | jq -r '.text'
```

## Two transports

Live streaming has two mechanisms, and `listen` auto-selects by the credential you have.
Force one with `--transport socket|rtm` (default `auto`).

| Transport | Credential | Notes |
|---|---|---|
| **Socket Mode** | app-level token (`xapp-`) | Official and robust. Only delivers the events your app subscribed to, for conversations the bot is in. |
| **RTM** | user token (`xoxp-`) or a session pair (`xoxc-`+`xoxd-`) | No Slack app required. Delivers everything the user account can see. Legacy/unofficial for `xoxc` tokens. |

`auto` picks Socket Mode when an app token is available, otherwise RTM.

### Socket Mode

Needs an app-level token with `connections:write`, Socket Mode enabled in the app config, and
event subscriptions configured (`message.im`, `message.channels`, `reaction_added`, …). Slack
only pushes subscribed events for conversations the bot belongs to.

```sh
slackctl auth login --kind app
slackctl listen --transport socket --json
```

Each envelope is acknowledged to Slack within its 3-second window *before* filtering, so
filtered-out events are consumed, not redelivered.

### RTM

The legacy Real Time Messaging WebSocket works with a **user or browser-session** credential —
no Slack app to create. This is the path for driving `listen` with a web-client session:

```sh
slackctl auth login --kind session     # or export SLACK_XOXC_TOKEN / SLACK_XOXD_TOKEN
slackctl listen --json                 # auto → RTM
```

!!! warning "RTM is legacy"
    Slack marks RTM as legacy and doesn't officially support it for `xoxc` session tokens. It
    works today but could be disabled, and some Enterprise Grid workspaces block it. For a
    durable listener, create a Slack app (no need to publish it), grab an app-level token, and
    use `--transport socket`.

## Filtering

Filters combine as a union — `--dms` **OR** `--channels` — and `--events` then narrows by
event type. With no filters, every received event streams.

```sh
slackctl listen --dms --json                          # only DMs
slackctl listen --channels C0123,C0456 --json          # only these channels
slackctl listen --dms --channels C0123 --json          # DMs OR that channel
slackctl listen --events message,reaction_added        # only these event types
```

## Output shapes

| Flag | Output |
|---|---|
| *(default)* | one compact human-readable line per event |
| `--json` | one JSON event object per line (NDJSON) |
| `--raw` | full wire frames (Socket Mode envelopes / RTM frames) as NDJSON |

```sh
slackctl listen --raw | jq .          # inspect everything on the wire
slackctl listen --json | jq 'select(.type=="reaction_added") | .reaction'
```

## Reliability

Both transports reconnect with a fresh URL and jittered backoff when a connection drops. RTM
keeps the socket alive with a periodic ping; Socket Mode acks each envelope. Ctrl-C shuts down
cleanly.
