package api

import (
	"context"
	"fmt"
)

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
