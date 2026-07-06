# Local message history & search

slackctl keeps a local, full-text-searchable mirror of the Slack messages it sees — so you
can search your history **instantly, offline, and without Slack's search API** (which is
user-token-only and rate-limited).

## How it fills up

Recording happens automatically as you use slackctl. Anything message-bearing is captured:

- **posts** you send (`msg post`, `msg update`, `msg template`)
- **history/replies** you fetch (`conversations history`, `conversations replies`,
  `conversations export`)
- **live events** you stream (`slackctl listen` records every message event it receives)

So a natural workflow — post, browse, export, or leave `listen` running — builds a searchable
archive you can query any time. The store is a per-workspace SQLite database.

```sh
# Leave a listener running to mirror a channel in real time…
slackctl listen --channels C0123456 &
# …then search what it captured, offline:
slackctl log search "deploy failed"
```

## Searching

```sh
slackctl log                                   # recent messages (newest first)
slackctl log --channel C0123456 --since 24h    # filter by channel and time
slackctl log --user U0123456 -o json
slackctl log search "deploy failed"            # full-text
slackctl log search "incident* AND prod"       # FTS5 operators: AND/OR/NOT, prefix*, "phrase"
```

Search uses SQLite **FTS5** when available (the pure-Go build ships it), with the operators
above. A plain query that isn't valid FTS5 syntax (e.g. `on-call`) transparently falls back to
a substring scan, so ordinary searches never error. `--since` accepts a Go duration (`24h`,
`7d`), a unix time, or a Slack ts.

## Managing the store

```sh
slackctl log stats                    # how many messages, channels, and the FTS mode
slackctl log path                     # where the database lives
slackctl log prune --older-than 2160h # delete anything older than 90 days
```

## Privacy & control

The database holds **message text**, so treat it like any local secret:

- It lives per-workspace under your config dir (`slackctl log path`) with `0600` permissions.
- Disable recording for a single call with `--no-store`, e.g.
  `slackctl msg post --channel C0123456 --text "…" --no-store`.
- `log prune` and deleting the file (`rm "$(slackctl log path)"`) clear it.
- The store is **local-only** — it's never uploaded, and `log` is excluded from the MCP tool
  surface so an AI agent can't browse it unprompted.

A store problem never breaks a command: if the database can't be opened, slackctl warns once
and continues without recording.
