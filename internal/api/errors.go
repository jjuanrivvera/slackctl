package api

import (
	"fmt"
	"strings"
)

// APIError is a typed Slack Web API failure. Slack answers HTTP 200 with
// {"ok":false,"error":"snake_case_code",...} for almost every failure — the HTTP status
// only matters for 429s and infrastructure errors — so the string Code is the primary key.
// We surface all of it and, crucially, append an actionable hint keyed by the code so the
// user knows the next move instead of staring at a bare "request failed" (GOAL.md §1).
type APIError struct {
	StatusCode int    // HTTP status (usually 200 even on failure; 429 on rate limit)
	Code       string // Slack error code, e.g. "channel_not_found"
	Warning    string // Slack's non-fatal warning, when present
	Needed     string // missing_scope: the scope the call needed
	Provided   string // missing_scope: the scopes the token actually has
	Body       string // raw body, for --verbose debugging
	Method     string // the Web API method that failed, for context
	RetryAfter int    // seconds from the Retry-After header on a 429, or 0
}

func (e *APIError) Error() string {
	var b strings.Builder
	if e.Method != "" {
		fmt.Fprintf(&b, "%s: ", e.Method)
	}
	code := e.Code
	if code == "" {
		code = "request failed"
	}
	fmt.Fprintf(&b, "Slack API error: %s", code)
	if e.Needed != "" {
		fmt.Fprintf(&b, " (needed: %s)", e.Needed)
	}
	if hint := e.hint(); hint != "" {
		fmt.Fprintf(&b, "\n  hint: %s", hint)
	}
	return b.String()
}

// hint maps Slack's error codes (each method page documents them; the auth family is shared
// by every method) to the concrete next action. Keyed remedies are the difference between a
// CLI you can debug and one you can't. Falls back to the HTTP status for non-envelope
// failures (a proxy 502, a raw 429).
func (e *APIError) hint() string {
	switch e.Code {
	case "token_expired":
		return "the credential has expired — re-run `slackctl auth login` (browser-session xoxc/xoxd tokens rotate, so `--kind session` needs re-capturing periodically)"
	case "not_authed", "invalid_auth", "token_revoked", "account_inactive":
		return "invalid or missing token — run `slackctl auth login` (OAuth tokens at https://api.slack.com/apps, or `--kind session` for a browser session)"
	case "missing_scope":
		if e.Needed != "" {
			return fmt.Sprintf("the token lacks the %q scope — add it under OAuth & Permissions and reinstall the app", e.Needed)
		}
		return "the token lacks a required scope — add it under OAuth & Permissions and reinstall the app"
	case "not_allowed_token_type":
		return "this method needs a different token kind (often a user token, xoxp-) — store one with `slackctl auth login --kind user` and pass --as-user (for `listen`, RTM needs a user/session token; Socket Mode needs an app token)"
	case "method_deprecated", "deprecated_endpoint":
		return "Slack has retired this method — for `listen`, RTM is legacy and may be blocked here; create a Slack app and use `--transport socket` with an app token (`auth login --kind app`)"
	case "no_permission", "ekm_access_denied", "restricted_action":
		return "the token's app or user is not allowed to do this — check app permissions and workspace policy"
	case "channel_not_found":
		return "no such channel — verify the id with `slackctl conversations list` (use the C…/D…/G… id, not the #name)"
	case "not_in_channel":
		return "the bot is not a member — run `slackctl conversations join --channel <id>` (public) or /invite it (private)"
	case "is_archived":
		return "the channel is archived — unarchive it first: `slackctl conversations unarchive --channel <id>`"
	case "message_not_found", "thread_not_found":
		return "no such message — verify the ts with `slackctl conversations history --channel <id>`"
	case "cant_invite_self", "already_in_channel":
		return "the user is already there — nothing to do"
	case "msg_too_long", "too_many_attachments":
		return "message too large — shorten the text or split it"
	case "rate_limited", "ratelimited":
		if e.RetryAfter > 0 {
			return fmt.Sprintf("rate limited — wait %ds and retry (slackctl backs off automatically; lower --rps for steady load)", e.RetryAfter)
		}
		return "rate limited — slow down (lower --rps; slackctl already honors Retry-After)"
	case "method_not_supported_for_channel_type":
		return "this conversation type does not support that method — check the channel kind (public/private/im/mpim)"
	case "user_not_found", "users_not_found":
		return "no such user — verify the id with `slackctl users list` or `slackctl users lookup-email`"
	case "invalid_arguments", "invalid_args", "invalid_form_data":
		return "bad arguments — re-check the flags; `--dry-run` prints the exact request"
	case "internal_error", "fatal_error", "service_unavailable":
		return "Slack server error — usually transient; retry shortly"
	case "invalid_cursor":
		return "the pagination cursor expired — re-run the command without --cursor"
	}
	switch {
	case e.StatusCode == 429:
		return "rate limited — slow down (lower --rps; slackctl already honors Retry-After)"
	case e.StatusCode >= 500:
		return "Slack server error — usually transient; retry shortly"
	case e.StatusCode == 404:
		return "endpoint not found — check the method name (`slackctl api <method>`) or your --base-url"
	case e.StatusCode == 401 || e.StatusCode == 403:
		// A bare 401/403 with no Slack error code usually comes from a proxy in front of the
		// API, not Slack itself — but the fix is still credential/permission-shaped.
		return "unauthorized — check your token/permissions (or a proxy in front of --base-url); run `slackctl auth login`"
	}
	return ""
}

// IsCode reports whether err is an APIError with the given Slack error code.
func IsCode(err error, code string) bool {
	ae, ok := err.(*APIError)
	return ok && ae.Code == code
}
