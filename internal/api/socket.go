package api

import (
	"context"
	"fmt"
	"net/http"
)

// DialHeaders returns the credential headers a WebSocket handshake must carry beyond the
// URL ticket. For a browser-session credential this is the `Cookie: d=<xoxd>` header, which
// the RTM gateway re-validates on the WebSocket UPGRADE (not just on the rtm.connect HTTP
// call) — without it the gateway answers invalid_auth. For a plain token it is empty (the
// URL ticket authenticates), so this is safe to apply for every transport.
func (c *Client) DialHeaders() http.Header {
	h := http.Header{}
	for k, v := range c.auth.ExtraHeaders(false) {
		h.Set(k, v)
	}
	return h
}

// OpenSocketURL asks apps.connections.open for a fresh Socket Mode wss:// URL. It requires
// an app-level token (xapp-, scope connections:write); Slack's URLs are single-use tickets,
// so callers fetch one per (re)connection attempt.
func (c *Client) OpenSocketURL(ctx context.Context) (string, error) {
	var out struct {
		URL string `json:"url"`
	}
	// Documented as POST; not idempotent in our retry sense (each call mints a ticket) —
	// the listener's reconnect loop is the retry layer.
	if err := c.CallInto(ctx, "apps.connections.open", nil, false, &out); err != nil {
		return "", err
	}
	if out.URL == "" {
		return "", fmt.Errorf("apps.connections.open returned no url")
	}
	return out.URL, nil
}

// OpenRTMURL asks rtm.connect for a fresh Real Time Messaging wss:// URL. RTM works with a
// user/session token (xoxc+xoxd) — the credential a slack-mcp-style setup already has — so it
// backs `slackctl listen` when no app-level token is available. RTM is a legacy API and is
// not officially supported for xoxc tokens; a workspace may return method_deprecated or
// not_allowed_token_type, which the caller surfaces with its hint. URLs are single-use and
// must be connected within ~30s, so callers fetch one per (re)connection.
func (c *Client) OpenRTMURL(ctx context.Context) (string, error) {
	var out struct {
		URL string `json:"url"`
	}
	if err := c.CallInto(ctx, "rtm.connect", nil, false, &out); err != nil {
		return "", err
	}
	if out.URL == "" {
		return "", fmt.Errorf("rtm.connect returned no url")
	}
	return out.URL, nil
}
