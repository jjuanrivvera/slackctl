package auth

import "fmt"

// TokenKind names the Slack token kinds a workspace profile can hold. One workspace often
// needs more than one: the bot token (xoxb-) drives most methods, a user token (xoxp-)
// unlocks user-only methods (search.messages, stars.*), and an app-level token (xapp-)
// opens Socket Mode connections for `slackctl listen`.
type TokenKind string

const (
	KindBot  TokenKind = "bot"
	KindUser TokenKind = "user"
	KindApp  TokenKind = "app"
)

// Valid reports whether k is a known kind.
func (k TokenKind) Valid() bool {
	switch k {
	case KindBot, KindUser, KindApp:
		return true
	}
	return false
}

// EnvVar is the conventional Slack environment variable for this kind — the ecosystem's
// real names (SLACK_BOT_TOKEN, ...), not an invented SLACKCTL_* scheme (GOAL.md §1).
func (k TokenKind) EnvVar() string {
	switch k {
	case KindUser:
		return "SLACK_USER_TOKEN"
	case KindApp:
		return "SLACK_APP_TOKEN"
	default:
		return "SLACK_BOT_TOKEN"
	}
}

// Key namespaces a keyring/file entry per profile AND kind. The bot token keeps the bare
// profile name so entries written by earlier versions keep working; '#' cannot appear in a
// profile name (config.ValidateProfileName), so the suffix can never collide with a profile.
func Key(profile string, kind TokenKind) string {
	if kind == KindBot || kind == "" {
		return profile
	}
	return fmt.Sprintf("%s#%s", profile, kind)
}
