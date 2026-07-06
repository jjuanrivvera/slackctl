package api

import (
	"fmt"
	"net/http"
	"strings"
)

// Authenticator applies a credential to an outgoing Web API request. Slack has one primary
// wire scheme (an `Authorization: Bearer` header) carried by several token kinds — bot
// (xoxb-), user (xoxp-), app-level (xapp-, Socket Mode only) — plus the browser-session
// scheme (xoxc- bearer + a `Cookie: d=xoxd-…` header) that tools like slack-mcp-server
// use. We keep the interface from the cliwright standard so redaction lives in one place
// and the design scales across kinds (GOAL.md §1 "auth providers").
type Authenticator interface {
	// Apply sets the credential on the request.
	Apply(req *http.Request)
	// Redacted returns the Authorization value masked, for --dry-run and logs.
	Redacted() string
	// Raw returns the Authorization value for --show-token dry-runs. Never log it elsewhere.
	Raw() string
	// ExtraHeaders returns credential headers beyond Authorization (the session cookie),
	// masked unless redact is false — so the --dry-run curl reproduces the real request.
	ExtraHeaders(redact bool) map[string]string
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

// ExtraHeaders is empty for plain tokens — the bearer header is the whole credential.
func (a *TokenAuth) ExtraHeaders(bool) map[string]string { return nil }

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

// SessionAuth authenticates as a real Slack user with browser-session credentials: an
// xoxc- token in the Authorization header plus the paired xoxd- cookie. This is the
// scheme slack-mcp-server uses; an xoxc token WITHOUT its cookie is rejected by Slack
// with invalid_auth. It carries the user's own identity, so it satisfies user-token-only
// methods (search.*, stars.*) too.
type SessionAuth struct {
	Token  string // xoxc-…
	Cookie string // the d cookie value, xoxd-…
}

// NewSessionAuth validates the pair's shape: both halves are required.
func NewSessionAuth(token, cookie string) (*SessionAuth, error) {
	t, c := strings.TrimSpace(token), strings.TrimSpace(cookie)
	if !strings.HasPrefix(t, "xoxc-") {
		return nil, fmt.Errorf("session token must start with xoxc- (got %q prefix)", RedactToken(t))
	}
	if c == "" {
		return nil, fmt.Errorf("session auth needs the xoxd- browser cookie (the d cookie) alongside the xoxc token")
	}
	return &SessionAuth{Token: t, Cookie: c}, nil
}

func (a *SessionAuth) Apply(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+a.Token)
	// The d cookie value is sent exactly as the browser stores it (often already
	// URL-encoded); re-encoding it breaks the session.
	req.Header.Set("Cookie", "d="+a.Cookie)
}

func (a *SessionAuth) Redacted() string { return "Bearer " + RedactToken(a.Token) }
func (a *SessionAuth) Raw() string      { return "Bearer " + a.Token }

func (a *SessionAuth) ExtraHeaders(redact bool) map[string]string {
	if redact {
		return map[string]string{"Cookie": "d=" + RedactToken(a.Cookie)}
	}
	return map[string]string{"Cookie": "d=" + a.Cookie}
}

// Method reports "session-token": user-grade identity from a browser session.
func (a *SessionAuth) Method() string { return "session-token" }

// RedactToken masks a Slack token, keeping only the kind prefix (e.g. "xoxb-****").
func RedactToken(token string) string {
	if i := strings.Index(token, "-"); i > 0 && i <= 10 {
		return token[:i+1] + "****"
	}
	return "****"
}
