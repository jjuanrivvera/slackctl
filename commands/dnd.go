package commands

// The dnd (Do Not Disturb) family. Reads (info/team-info) take a bot or user token; the
// write verbs act on the token owner's own DND, so they need a user/session token.

func init() {
	registerGroup(group{
		Use:   "dnd",
		Short: "Do Not Disturb — snooze and status",
		Cmds: []methodCmd{
			{
				Use: "info", Method: "dnd.info", Kind: kindRead,
				Short:   "Show a user's Do Not Disturb status",
				Example: "  slackctl dnd info\n  slackctl dnd info --user U0123456",
				Columns: []string{"dnd_enabled", "next_dnd_start_ts", "next_dnd_end_ts", "snooze_enabled"},
				Flags: []flagSpec{
					{Name: "user", Kind: flagString, Usage: "user id (default: the token's own user)"},
				},
			},
			{
				Use: "set-snooze", Method: "dnd.setSnooze", Kind: kindWrite, Token: tokenUserRequired,
				Short:   "Turn on snooze for a number of minutes",
				Example: "  slackctl dnd set-snooze --minutes 60",
				Flags: []flagSpec{
					{Name: "minutes", Param: "num_minutes", Kind: flagInt, Required: true, Usage: "snooze duration in minutes"},
				},
			},
			{
				Use: "end-snooze", Method: "dnd.endSnooze", Kind: kindWrite, Token: tokenUserRequired,
				Short:   "End the current snooze",
				Example: "  slackctl dnd end-snooze",
			},
			{
				Use: "end-dnd", Method: "dnd.endDnd", Kind: kindWrite, Token: tokenUserRequired,
				Short:   "End the current Do Not Disturb session",
				Example: "  slackctl dnd end-dnd",
			},
			{
				Use: "team-info", Method: "dnd.teamInfo", Kind: kindRead,
				Short:     "Show DND status for several users",
				Example:   "  slackctl dnd team-info --users U0123456,U0456789 -o json",
				ResultKey: "users",
				Flags: []flagSpec{
					{Name: "users", Kind: flagStringSlice, Usage: "user ids to check"},
				},
			},
		},
	})
}
