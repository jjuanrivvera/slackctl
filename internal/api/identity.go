package api

import "context"

// Identity is auth.test's answer — the whoami payload. A bot token adds bot_id; Enterprise
// Grid orgs add enterprise_id.
type Identity struct {
	URL          string `json:"url"`
	Team         string `json:"team"`
	User         string `json:"user"`
	TeamID       string `json:"team_id"`
	UserID       string `json:"user_id"`
	BotID        string `json:"bot_id,omitempty"`
	EnterpriseID string `json:"enterprise_id,omitempty"`
}

// AuthTest verifies the token against auth.test and returns who it belongs to. It is the
// verification step behind `auth login`, `auth status`, and `doctor`.
func (c *Client) AuthTest(ctx context.Context) (*Identity, error) {
	var id Identity
	if err := c.CallInto(ctx, "auth.test", nil, true, &id); err != nil {
		return nil, err
	}
	return &id, nil
}
