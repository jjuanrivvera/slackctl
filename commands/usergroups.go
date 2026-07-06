package commands

// The usergroups family (@eng-style mention groups). usergroups.users.* surfaces as
// `members` / `members-update` so the CLI reads naturally.

func init() {
	registerGroup(group{
		Use:     "usergroups",
		Aliases: []string{"usergroup", "groups"},
		Short:   "Manage user groups (@mention groups)",
		Cmds: []methodCmd{
			{
				Use: "list", Method: "usergroups.list", Kind: kindRead,
				Short:     "List user groups",
				Example:   "  slackctl usergroups list\n  slackctl usergroups list --include-disabled --include-count",
				ResultKey: "usergroups",
				Columns:   []string{"id", "handle", "name", "user_count"},
				Flags: []flagSpec{
					{Name: "include-count", Kind: flagBool, Usage: "include member counts"},
					{Name: "include-disabled", Kind: flagBool, Usage: "include disabled groups"},
					{Name: "include-users", Kind: flagBool, Usage: "include each group's user list"},
				},
			},
			{
				Use: "create", Method: "usergroups.create", Kind: kindWrite,
				Short:     "Create a user group",
				Example:   `  slackctl usergroups create --name "On-call" --handle oncall`,
				ResultKey: "usergroup",
				Columns:   []string{"id", "handle", "name"},
				Flags: []flagSpec{
					{Name: "name", Kind: flagString, Required: true, Usage: "display name"},
					{Name: "handle", Kind: flagString, Usage: "@mention handle (no @)"},
					{Name: "description", Kind: flagString, Usage: "short description"},
					{Name: "channels", Kind: flagStringSlice, Usage: "default channel ids for the group"},
				},
			},
			{
				Use: "update", Method: "usergroups.update", Kind: kindWrite,
				Short:     "Update a user group's name/handle/description",
				Example:   `  slackctl usergroups update --usergroup S0123456 --name "On-call (EMEA)"`,
				ResultKey: "usergroup",
				Flags: []flagSpec{
					{Name: "usergroup", Kind: flagString, Required: true, Usage: "user group id (S…)"},
					{Name: "name", Kind: flagString, Usage: "new display name"},
					{Name: "handle", Kind: flagString, Usage: "new @mention handle"},
					{Name: "description", Kind: flagString, Usage: "new description"},
					{Name: "channels", Kind: flagStringSlice, Usage: "new default channel ids"},
				},
			},
			{
				Use: "enable", Method: "usergroups.enable", Kind: kindWrite,
				Short:     "Re-enable a disabled user group",
				Example:   "  slackctl usergroups enable --usergroup S0123456",
				ResultKey: "usergroup",
				Flags: []flagSpec{
					{Name: "usergroup", Kind: flagString, Required: true, Usage: "user group id (S…)"},
				},
			},
			{
				Use: "disable", Method: "usergroups.disable", Kind: kindDestructive,
				Short:     "Disable a user group",
				Example:   "  slackctl usergroups disable --usergroup S0123456",
				ResultKey: "usergroup",
				Flags: []flagSpec{
					{Name: "usergroup", Kind: flagString, Required: true, Usage: "user group id (S…)"},
				},
			},
			{
				Use: "members", Method: "usergroups.users.list", Kind: kindRead,
				Short:     "List a user group's members",
				Example:   "  slackctl usergroups members --usergroup S0123456 -o id",
				ResultKey: "users",
				Flags: []flagSpec{
					{Name: "usergroup", Kind: flagString, Required: true, Usage: "user group id (S…)"},
					{Name: "include-disabled", Kind: flagBool, Usage: "include members of disabled groups"},
				},
			},
			{
				Use: "members-update", Method: "usergroups.users.update", Kind: kindWrite,
				Short:     "Replace a user group's member list",
				Long:      "Replaces the ENTIRE member list with --users (Slack has no incremental add/remove).",
				Example:   "  slackctl usergroups members-update --usergroup S0123456 --users U111,U222",
				ResultKey: "usergroup",
				Flags: []flagSpec{
					{Name: "usergroup", Kind: flagString, Required: true, Usage: "user group id (S…)"},
					{Name: "users", Kind: flagStringSlice, Required: true, Usage: "the complete new member list"},
				},
			},
		},
	})
}
