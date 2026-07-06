# DECISIONS.md — pinned assumptions (cliwright GOAL.md §11)

One line per decision: question → decision → why. The build loop reads this back every
pass and never silently re-decides.

## Scope & completeness

- **Surface scope** → the core messaging/read/people surface (conversations list/history/
  replies/search/unreads/export/post/mark, channels, users search/info/status/presence,
  usergroups, reactions, saved items, pins, bookmarks, files upload/download, dnd, canvases,
  assistant search) plus a hand-written `listen` event stream — a focused surface, not 308.
- coverage-waiver: the surface is scoped to the core messaging/read/people surface + files +
  dnd + bookmarks + canvases + `listen` (86 of the 308 enumerated methods, 27%). Of the
  remainder: admin.* (106) is Enterprise-Grid-admin-only, files.upload is sunset (Nov 2025 —
  the external-upload flow IS shipped), rtm.*/dialog.*/reminders.*/users.setActive are
  deprecated or retired, and views/workflows/functions/slackLists/apps.datastore serve
  app-platform builders, not a messaging CLI. Widening stays a candidate; the enumeration
  total stays in the manifest so the gap is loud.
- **Enumeration source** → `https://docs.slack.dev/reference/methods.json` (308 entries,
  2026-07-06), cross-checked against sitemap.xml and the Docusaurus route registry (both 308).
  The official `slackapi/slack-api-specs` OpenAPI is stale (174 methods, 10 removed ones) —
  rejected as primary source.

## Architecture

- **Pattern** → generic *method-command* builder (a Pattern-B trigger case: the Web API is
  RPC-on-methods, not CRUD-on-resources). A group = noun, verb = one Web API method,
  declared in a `methodCmd`; shared builder stamps flags, annotations, pagination, rendering.
- **HTTP verbs** → read methods (Kind=read) go as GET with query params; writes as
  `application/x-www-form-urlencoded; charset=utf-8` POST. Slack accepts either for form
  args ("GET querystring, POST body, or a mix"); tying GET to reads keeps the
  only-retry-idempotent rule literal HTTP semantics.
- **JSON args** → non-scalar params (blocks, attachments, metadata) are JSON-serialized into
  the form/query field instead of switching to a JSON body — Slack accepts JSON-in-field for
  exactly these args, which removes the per-method "does it accept a JSON body?" question.
- **Envelope** → success/failure keyed on `ok`, never the HTTP status (Slack returns 200 with
  `ok:false`); the full body is the payload (fields are siblings of `ok`), so each command
  declares a `ResultKey` (channels/messages/members/…) extracted before rendering.
- **Envelope-level rate limit** → `ok:false, error:ratelimited|rate_limited` is treated as a
  429 (penalize the limiter, retry with Retry-After) even when the HTTP status is 200.
- **Pagination** → cursor methods run through one walker (`CallAllPages`, keyed on
  `response_metadata.next_cursor`; `--limit` caps total items, `--all` fetches everything;
  per-request page size 200 per Slack's 100–200 recommendation). No user-facing `--cursor`
  flag — raw cursor plumbing stays available via `slackctl api`. search.* keeps Slack's
  page/count pagination as explicit `--page/--count` flags (different protocol, no walker).
- **Retry policy** → only reads auto-retry on 5xx/network; 429 retries for everything
  (rejected, not processed), honoring Retry-After. Backoff is AWS full jitter — deliberate.
- **Client-side rate pacing** → fixed 4 rps default + halve-on-429/gradual-restore; Slack
  exposes no quota headers (tiers are per-method), so adaptive-by-headers is impossible.
- **Flexible JSON types** → kept from the house core (ID/Int/Bool/Money/StringOrSlice);
  Slack is string-typed enough that most command paths render raw JSON, but the types guard
  the few numeric fields and keep the fuzz-tested core intact.

## Tokens & auth

- **Token kinds per workspace profile** → bot (xoxb, default), user (xoxp), app (xapp,
  Socket Mode only), and session (xoxc token + xoxd cookie), stored in the keyring under
  `<profile>`, `<profile>#user`, `<profile>#app`, `<profile>#session` ('#' is banned in
  profile names so the suffix cannot collide). The session entry is one JSON blob
  (`{"xoxc":…,"xoxd":…}`) so the pair can never be half-written.
- **Session (browser) auth** → the scheme Slack's own web client uses: an `xoxc-` bearer token
  plus the paired `xoxd-` "d" cookie sent as `Cookie: d=<xoxd>`. It requires no created
  Slack app, carries the user's own identity, and so is a valid FALLBACK for bot- and
  user-kind commands (search/saved included) when no OAuth token is stored — this is the
  "drive slackctl with a browser session, no app" path. It does NOT back the app kind:
  `apps.connections.open` (Socket Mode / `listen`) genuinely needs an app-level xapp token.
  The xoxd value is sent verbatim (browsers store it already URL-encoded; re-encoding
  breaks the session). Auth precedence per kind: `$SLACKCTL_TOKEN`/kind env var >
  `$SLACK_XOXC_TOKEN`+`$SLACK_XOXD_TOKEN` > keyring OAuth token > keyring session pair.
- **Env vars** → the ecosystem's real names: SLACK_BOT_TOKEN / SLACK_USER_TOKEN /
  SLACK_APP_TOKEN, plus SLACKCTL_TOKEN as an explicit any-kind override. Profile selector:
  SLACKCTL_WORKSPACE (SLACKCTL_PROFILE kept as generic alias).
- **User-token-only methods** → search.* and stars.* reject bot tokens
  (not_allowed_token_type); their commands resolve the user token and fail fast with a
  "auth login --kind user" hint. Everything else defaults to the bot token; a global
  `--as-user` switches any command to the user token.
- **App-token verification** → `auth login --kind app` skips auth.test (app-level tokens
  can't call it); its real verification is opening a Socket Mode connection in `listen`.

## Surface details

- **Profile flag** → `--workspace` (a profile IS a workspace), `--profile` kept as a hidden
  alias.
- **Destructive classification** → irreversible-or-disruptive: msg delete/delete-scheduled,
  conversations archive/kick/leave, usergroups disable. Cheap reversible removals
  (reactions remove, pins remove, saved remove, conversations close/unarchive) are ordinary
  writes — still approval-gated for agents, not hard-blocked.
- **saved = stars.*** → Slack deprecated stars (Later tab) but ships no Later API; stars.*
  is still the only saved-items surface, so `saved` wraps it and the help says so.
- **unreads** → composite command (users.conversations + per-channel conversations.info
  unread_count); Slack has no single unreads endpoint for bots. Meaningful counts need a
  user token; the command works with either and says which it used.
- **users search** → client-side filter over users.list (name/real_name/email contains);
  Slack has no public users.search for bot tokens.
- **conversations.history/replies May-2025 limits** (1 req/min, 15 items for NEW
  non-Marketplace *commercially distributed* apps) → documented in the skill/README;
  internal single-workspace apps — the CLI's audience — keep Tier 3, so no special-casing.
- **listen** → hand-written Socket Mode client: apps.connections.open → wss dial →
  ack every envelope (`{"envelope_id":…}`) within 3s BEFORE processing → NDJSON events to
  stdout; reconnect on `disconnect`/socket close with fresh URL; filters: --dms, --channels,
  --events; --raw emits whole envelopes. Uses coder/websocket (maintained, minimal, no CGO).
- **listen ack-before-filter** → envelopes are acked whether or not they match filters;
  filtering is client-side output shaping, not consumption semantics (unacked = redelivered).
- **listen transports** → `listen` streams over EITHER transport so it works with whatever
  credential you have; `--transport auto|socket|rtm` (default auto):
  - **socket** — Socket Mode (`apps.connections.open`, app-level xapp token). Official,
    robust, envelope + 3s ack. Auto picks this when an app token is resolvable.
  - **rtm** — the legacy Real Time Messaging WebSocket (`rtm.connect`), which accepts a
    user/session token (xoxp or the xoxc+xoxd browser pair). This is the "stream with the
    same creds a browser session already has" path — no Slack app needed. Auto picks this
    when there's no app token. RTM frames ARE the event objects (no envelope, no ack); the
    client keeps the socket alive with a 30s app-level ping and reconnects with a fresh URL.
  - Both feed one shared filter/render path (`internal/slackevent.Meta`), so `--dms`/
    `--channels`/`--events`/`--json`/`--raw` behave identically across transports.
  - **Honest caveat (recorded, not hidden):** RTM is a legacy API and is *not officially
    supported* for xoxc tokens — Slack could disable it and some Enterprise Grid workspaces
    block it. Socket Mode (an xapp token, cheap to create without publishing an app) is the
    durable path; RTM is the no-app convenience path. The `rtm.connect` error hint points
    this out on `method_deprecated`/`not_allowed_token_type`.

## v0.2 additions

- **files upload** → `files.upload` was sunset (Nov 2025), so `files upload` runs the
  external-upload flow: `files.getUploadURLExternal` → multipart POST of the bytes to the
  returned URL (unauthenticated; the URL carries the ticket) → `files.completeUploadExternal`
  to share. `files download` resolves `url_private` via `files.info` then GETs it WITH the
  Authorization header (Slack file URLs require it). Both live in `internal/api/files.go`.
- **assistant search-context** → wrapped specifically because `assistant.search.context`
  accepts a BOT token, unlike `search.*` (user-only) — it gives bot-only setups a search path.
- **dnd/status/presence token kinds** → `dnd.set*` and `users.profile.set` have no bot scope
  (user-only) → `tokenUserRequired`; `dnd.info`/`teamInfo`/`users.setPresence` take either.
  `users set-status` is a hand-written Extra: `users.profile.set` takes a `profile` OBJECT, so
  the command builds `{status_text,status_emoji,status_expiration}` from flags.
- **canvas JSON args** → canvas edits are structured operations, so `--content`/`--changes`/
  `--criteria` are `flagJSON` (Slack's own edit-op shapes) rather than invented flat flags.
- **beyond-API composites** → `conversations export` (JSONL dump of history + optional thread
  replies, deduping the parent), `msg template` (Go text/template with `missingkey=error` so an
  unresolved var fails rather than posts a blank; `--blocks` sends the rendered JSON as Block
  Kit), and `listen --since` (backfills `--channels` history before the live stream; a bare
  duration like `1h` resolves to an oldest bound; the channel is injected into history messages
  so filters see the live event shape).
- **listen read-deadline** → both transports bound a single frame read by `ReadTimeout`
  (default 90s; Slack pings a healthy socket well within it). A read that stalls past it means
  the connection went half-open (server gone, no FIN) → reconnect instead of hanging forever.
- **read-shape allowlist** → the agent-guard raw-`api` gate and its classification invariant
  recognize `teaminfo`/`context`/`lookup` as read-shaped final segments (dnd.teamInfo,
  assistant.search.context, canvases.sections.lookup), so those reads pass the hatch while
  every mutating method stays blocked.

## v0.3 — local message store (`log`)

- **Why** → a local SQLite mirror of the messages slackctl sees, full-text searchable offline
  — sidesteps Slack search being user-token-only AND rate-limited. Ports the proven tgctl
  Recorder pattern (stripped during the fork), adapted to Slack message shapes.
- **Driver** → modernc.org/sqlite (pure Go, no cgo) so `CGO_ENABLED=0` cross-compile keeps
  working; FTS5 for full-text, with an automatic LIKE-scan fallback when a build lacks FTS5.
- **Capture points** → the api.Client `Recorder` hook fires on every successful non-dry-run
  call; the storeRecorder persists chat.postMessage/update (sends) and conversations.history/
  replies (reads you fetch). `listen` records each streamed message event directly (events
  don't pass through Call). So normal use — posting, browsing history, exporting, listening —
  fills the store; `log search` then finds it with no API call.
- **Privacy = opt-in-shaped** → recording is ON by default but easy to disable: global
  `--no-store`; per-workspace DB under `<config>/history/<workspace>.db` at 0600 (holds
  message text); `log prune`/`log path` for control. A store error NEVER fails a command
  (warn-once and continue) — a broken history must not break a post.
- **Dedup** → INSERT OR IGNORE on UNIQUE(workspace, channel, ts); re-fetching the same history
  never duplicates, and FTS is only indexed on a genuinely-new row.
- **Search robustness** → a plain query that isn't valid FTS5 (a bare `on-call` reads `-call`
  as an operator) transparently falls back to a LIKE scan for that query, so search never
  errors on ordinary input; power users still get FTS5 operators (AND/OR/NOT, prefix*, phrase).
- **`since` as text** → Slack ts is fixed-width zero-padded, so it sorts chronologically as
  text; a `--since 24h` duration resolves to a unix-seconds lower bound compared directly.
- **Agent surface** → `log` is excluded from the MCP tool surface and the guard's local-groups
  allowlist (it reads potentially-sensitive local message text; not something an agent should
  browse unprompted). Every command that builds a client now `defer`s client.Close() so the
  store's SQLite handle is released (matters on Windows, where an open handle blocks temp
  cleanup).

## Testing

- **Live smoke test** → skipped for unit CI (mocks only); done manually against a real
  workspace at each release. Never a hard gate (GOAL.md §4 Phase F2).
