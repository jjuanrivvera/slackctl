# DECISIONS.md — pinned assumptions (cliwright GOAL.md §11)

One line per decision: question → decision → why. The build loop reads this back every
pass and never silently re-decides.

## Scope & completeness

- **Surface scope** → Slack-MCP feature parity (conversations list/history/replies/search/
  unreads/post/mark, channels, users search/info, usergroups, reactions, saved items) plus a
  hand-written Socket Mode `listen` — per the commissioning brief, not the full 308-method API.
- coverage-waiver: the surface is user-scoped to Slack-MCP parity + `listen` (60 of the 308
  enumerated methods, 19%). Of the remainder: admin.* (106) is Enterprise-Grid-admin-only,
  files.upload is sunset (Nov 2025) and the external-upload flow is out of scope, rtm.*/dialog.*/
  reminders.*/users.setActive are deprecated or retired, and views/workflows/functions/canvases/
  slackLists/apps.datastore serve app-platform builders, not a messaging CLI. Widening is
  tracked as a v2 candidate; the enumeration total stays in the manifest so the gap is loud.
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
- **Session (browser) auth** → the scheme `slack-mcp-server` uses: an `xoxc-` bearer token
  plus the paired `xoxd-` "d" cookie sent as `Cookie: d=<xoxd>`. It requires no created
  Slack app, carries the user's own identity, and so is a valid FALLBACK for bot- and
  user-kind commands (search/saved included) when no OAuth token is stored — this is the
  "drive slackctl exactly like the Slack MCP" path. It does NOT back the app kind:
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
    same creds a slack-mcp setup already has" path — no Slack app needed. Auto picks this
    when there's no app token. RTM frames ARE the event objects (no envelope, no ack); the
    client keeps the socket alive with a 30s app-level ping and reconnects with a fresh URL.
  - Both feed one shared filter/render path (`internal/slackevent.Meta`), so `--dms`/
    `--channels`/`--events`/`--json`/`--raw` behave identically across transports.
  - **Honest caveat (recorded, not hidden):** RTM is a legacy API and is *not officially
    supported* for xoxc tokens — Slack could disable it and some Enterprise Grid workspaces
    block it. Socket Mode (an xapp token, cheap to create without publishing an app) is the
    durable path; RTM is the no-app convenience path. The `rtm.connect` error hint points
    this out on `method_deprecated`/`not_allowed_token_type`.

## Testing

- **Live smoke test** → skipped: no test workspace/credentials provided (mocks only). Never
  a hard gate (GOAL.md §4 Phase F2).
