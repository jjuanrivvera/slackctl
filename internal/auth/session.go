package auth

import (
	"encoding/json"
	"fmt"
)

// SessionCreds is a browser-session credential pair (the scheme slack-mcp-server uses):
// the xoxc- token plus the paired xoxd- "d" cookie. Both are required — an xoxc token
// without its cookie is rejected by Slack. It is stored as one JSON keyring entry under the
// session key so the pair can never be half-written.
type SessionCreds struct {
	Token  string `json:"xoxc"`
	Cookie string `json:"xoxd"`
}

// SetSession stores the xoxc/xoxd pair for a profile as a single JSON secret.
func SetSession(s Store, profile string, creds SessionCreds) error {
	blob, err := json.Marshal(creds)
	if err != nil {
		return err
	}
	return s.Set(Key(profile, KindSession), string(blob))
}

// GetSession loads the stored session pair for a profile. It returns ErrNotFound when none
// is stored, and a descriptive error if the stored blob is corrupt.
func GetSession(s Store, profile string) (SessionCreds, error) {
	raw, err := s.Get(Key(profile, KindSession))
	if err != nil {
		return SessionCreds{}, err
	}
	var creds SessionCreds
	if err := json.Unmarshal([]byte(raw), &creds); err != nil {
		return SessionCreds{}, fmt.Errorf("stored session credentials for %q are corrupt: %w", profile, err)
	}
	if creds.Token == "" || creds.Cookie == "" {
		return SessionCreds{}, fmt.Errorf("stored session credentials for %q are incomplete", profile)
	}
	return creds, nil
}
