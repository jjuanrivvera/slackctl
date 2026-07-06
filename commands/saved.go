package commands

// saved wraps stars.* — Slack deprecated stars in favor of the "Later" tab but ships no
// Later API, so stars.* remains the only saved-items surface the platform exposes
// (DECISIONS.md). All three methods are user-token-only.

func init() {
	registerGroup(group{
		Use:     "saved",
		Aliases: []string{"stars"},
		Short:   "Saved items (legacy stars API; needs a user token)",
		Long: `Manage saved items via the legacy stars.* API. Slack deprecated stars when it moved to
the "Later" tab, and Later has no public API yet — so this is still the only saved-items
surface. Items starred here may not appear in the Later tab. All commands need a user
token: slackctl auth login --kind user.`,
		Cmds: []methodCmd{
			{
				Use: "list", Method: "stars.list", Kind: kindRead, Token: tokenUserRequired,
				Short:     "List saved (starred) items",
				Example:   "  slackctl saved list\n  slackctl saved list -o json",
				Paginated: true, ResultKey: "items",
				Columns: []string{"type", "channel"},
			},
			{
				Use: "add", Method: "stars.add", Kind: kindWrite, Token: tokenUserRequired,
				Short:   "Save (star) a message, file, or channel",
				Example: "  slackctl saved add --channel C0123456 --ts 1720000000.000100",
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Usage: "conversation id (with --ts: the message; alone: the channel)"},
					{Name: "ts", Param: "timestamp", Kind: flagString, Usage: "message ts to save"},
					{Name: "file", Kind: flagString, Usage: "file id to save"},
				},
			},
			{
				Use: "remove", Method: "stars.remove", Kind: kindWrite, Token: tokenUserRequired,
				Short:   "Remove a saved (starred) item",
				Example: "  slackctl saved remove --channel C0123456 --ts 1720000000.000100",
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Usage: "conversation id"},
					{Name: "ts", Param: "timestamp", Kind: flagString, Usage: "message ts to unsave"},
					{Name: "file", Kind: flagString, Usage: "file id to unsave"},
				},
			},
		},
	})
}
