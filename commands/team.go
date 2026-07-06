package commands

// team + emoji: small read-only workspace lookups.

func init() {
	registerGroup(group{
		Use:   "team",
		Short: "Workspace information",
		Cmds: []methodCmd{
			{
				Use: "info", Method: "team.info", Kind: kindRead,
				Short:     "Show the workspace's name, domain, and icon",
				Example:   "  slackctl team info",
				ResultKey: "team",
				Columns:   []string{"id", "name", "domain"},
				Flags: []flagSpec{
					{Name: "team", Kind: flagString, Usage: "team id (org tokens; default: the token's team)"},
					{Name: "domain", Kind: flagString, Usage: "look up a team by its slack.com subdomain"},
				},
			},
			{
				Use: "profile", Method: "team.profile.get", Kind: kindRead,
				Short:     "Show the workspace's profile field definitions",
				Example:   "  slackctl team profile -o json",
				ResultKey: "profile",
				Flags: []flagSpec{
					{Name: "visibility", Kind: flagString, Usage: "filter fields: all|visible|hidden"},
				},
			},
		},
	})

	registerGroup(group{
		Use:   "emoji",
		Short: "Custom emoji",
		Cmds: []methodCmd{
			{
				Use: "list", Method: "emoji.list", Kind: kindRead,
				Short:     "List the workspace's custom emoji",
				Example:   "  slackctl emoji list -o json | jq 'keys'",
				ResultKey: "emoji",
				Flags: []flagSpec{
					{Name: "include-categories", Kind: flagBool, Usage: "include emoji category metadata"},
				},
			},
		},
	})
}
