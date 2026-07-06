# Authentication & tokens

slackctl stores every credential in your **OS keyring** (macOS Keychain, Linux Secret
Service, Windows Credential Manager), with an encrypted-file fallback for headless hosts.
Tokens are never written to a config file, and `--dry-run` redacts them.

## Token kinds

A workspace profile can hold up to four credential kinds. Store one with
`slackctl auth login --kind <kind>`:

| Kind | Looks like | Where to get it | Unlocks |
|---|---|---|---|
| `bot` (default) | `xoxb-…` | App → OAuth & Permissions → Bot User OAuth Token | almost everything |
| `user` | `xoxp-…` | App → OAuth & Permissions → User OAuth Token | `search`, `saved`, `--as-user`, `listen` via RTM |
| `app` | `xapp-…` | App → Basic Information → App-Level Tokens (scope `connections:write`) | `listen` via Socket Mode |
| `session` | `xoxc-…` + `xoxd-…` | your logged-in browser | everything **except** Socket Mode — no Slack app needed |

```sh
slackctl auth login                 # bot token (default)
slackctl auth login --kind user     # user token
slackctl auth login --kind app      # app-level token
slackctl auth login --kind session  # browser session (xoxc + xoxd)
slackctl auth logout                # remove all tokens for the workspace
```

### Which token does a command use?

Most commands use the **bot** token. `search` and `saved` are user-only (Slack rejects bot
tokens there), so they use the **user** token automatically. Add `--as-user` to run any other
command with the user token instead. `listen` picks `app` or `user`/`session` depending on the
transport — see [The `listen` command](listen.md).

### Bot scopes

Request the scopes matching what you'll call, under **OAuth & Permissions**, then reinstall the
app:

`channels:read` `groups:read` `im:read` `mpim:read` (listing) · `channels:history`
`groups:history` `im:history` `mpim:history` (history/replies/export) · `chat:write` (posting) ·
`channels:manage` `groups:write` `im:write` `mpim:write` (create/invite/topic/mark) ·
`channels:join` · `reactions:read` `reactions:write` · `users:read` `users:read.email`
`users:write` (presence) · `usergroups:read` `usergroups:write` · `pins:read` `pins:write` ·
`bookmarks:read` `bookmarks:write` · `files:read` `files:write` (upload/download) ·
`dnd:read` · `canvases:read` `canvases:write` · `emoji:read` · `team:read`.

For the user token: `search:read`, `stars:read`, `stars:write`, `dnd:write` (snooze),
`users.profile:write` (set status). `assistant search-context` works with a bot token
(`search:read.public` + friends) — unlike `search`, which needs a user token.

## No Slack app? Use a browser session

A Slack **web-client session** — an `xoxc-` token plus the paired `xoxd-` `d` cookie, the
credentials your browser holds — authenticates as *you*, with no app to create. It backs
bot- and user-kind commands (search and saved items included). The only thing it can't do is
Socket Mode; `listen` uses [RTM](listen.md#rtm) instead.

```sh
slackctl auth login --kind session      # paste the xoxc token, then the xoxd cookie
# or env-only:
export SLACK_XOXC_TOKEN=xoxc-…
export SLACK_XOXD_TOKEN=xoxd-…
slackctl auth status
slackctl conversations list
slackctl listen --dms --json            # streams over RTM
```

!!! note "Where the values live"
    Both halves come from a logged-in browser session. slackctl sends the `d` cookie
    verbatim (browsers store it URL-encoded — re-encoding breaks the session) and never logs
    or commits either value.

## Environment variables

Env vars override the keyring, so they're handy for CI:

| Variable | Purpose |
|---|---|
| `SLACK_BOT_TOKEN` / `SLACK_USER_TOKEN` / `SLACK_APP_TOKEN` | per-kind tokens |
| `SLACK_XOXC_TOKEN` + `SLACK_XOXD_TOKEN` | a browser session pair |
| `SLACKCTL_TOKEN` | explicit override for any single-token kind |
| `SLACKCTL_WORKSPACE` | select the active workspace profile |

**Precedence** (per kind): `SLACKCTL_TOKEN` / the kind's env var → the session env pair →
the keyring OAuth token → the keyring session pair.

## Workspaces (profiles)

A profile *is* a workspace. Name them and switch freely:

```sh
slackctl auth login --workspace acme
slackctl config list-profiles
slackctl config use acme                     # set the default
slackctl conversations list --workspace other  # one-off
```

`--profile` is a hidden alias for `--workspace`, so generic scripts keep working.

## Verify & troubleshoot

```sh
slackctl auth status     # identity, workspace, validity
slackctl doctor          # config + credentials + connectivity + clock, exits non-zero on failure
```

Common errors and their fixes:

| Error | Fix |
|---|---|
| `not_authed` / `invalid_auth` | `slackctl auth login` (check the workspace) |
| `missing_scope` | the message names the scope — add it and reinstall the app |
| `not_allowed_token_type` | user-only method: `auth login --kind user`, then `--as-user` |
| `not_in_channel` | `slackctl conversations join --channel <id>` (public) or /invite the bot |
