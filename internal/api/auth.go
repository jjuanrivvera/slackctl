package api

import (
	"fmt"
	"net/http"
	"strings"
)

// Authenticator applies a credential to an outgoing Web API request. Slack has one wire
// scheme (an `Authorization: Bearer` header) carried by several token kinds — bot (xoxb-),
// user (xoxp-), and app-level (xapp-, Socket Mode only). We keep the interface from the
// cliwright standard so redaction lives in one place and the design scales to new kinds
// (GOAL.md §1 "auth providers").
type Authenticator interface {
	// Apply sets the credential on the request.
	Apply(req *http.Request)
	// Redacted returns the Authorization value masked, for --dry-run and logs.
	Redacted() string
	// Raw returns the Authorization value for --show-token dry-runs. Never log it elsewhere.
	Raw() string
	// Method is the non-secret auth method name recorded in the profile.
	Method() string
}

// TokenAuth authenticates with a Slack token via the Authorization header.
type TokenAuth struct {
	Token string
}

// NewTokenAuth validates the token shape loosely: Slack tokens start with "xox" (xoxb-,
// xoxp-, xoxe.xoxp-, ...) or "xapp-". Reject only the obviously wrong, not unknown future
// prefixes.
func NewTokenAuth(token string) (*TokenAuth, error) {
	t := strings.TrimSpace(token)
	if t == "" {
		return nil, fmt.Errorf("empty token")
	}
	if !strings.HasPrefix(t, "xox") && !strings.HasPrefix(t, "xapp-") {
		return nil, fmt.Errorf("token does not look like a Slack token (expected an xoxb-/xoxp-/xapp- prefix)")
	}
	return &TokenAuth{Token: t}, nil
}

func (a *TokenAuth) Apply(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+a.Token)
}

// Redacted keeps the token-kind prefix — it is not secret and answers the debugging
// question "was this the bot or the user token?" — and masks the rest.
func (a *TokenAuth) Redacted() string { return "Bearer " + RedactToken(a.Token) }

func (a *TokenAuth) Raw() string { return "Bearer " + a.Token }

// Method reports the token kind from its prefix ("bot-token", "user-token", "app-token").
func (a *TokenAuth) Method() string {
	switch {
	case strings.HasPrefix(a.Token, "xoxb-"):
		return "bot-token"
	case strings.HasPrefix(a.Token, "xoxp-"), strings.HasPrefix(a.Token, "xoxe.xoxp-"):
		return "user-token"
	case strings.HasPrefix(a.Token, "xapp-"):
		return "app-token"
	default:
		return "token"
	}
}

// RedactToken masks a Slack token, keeping only the kind prefix (e.g. "xoxb-****").
func RedactToken(token string) string {
	if i := strings.Index(token, "-"); i > 0 && i <= 10 {
		return token[:i+1] + "****"
	}
	return "****"
}
