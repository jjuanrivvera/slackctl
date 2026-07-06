package commands

// The reactions family. Emoji names are passed WITHOUT colons (thumbsup, not :thumbsup:).

func init() {
	registerGroup(group{
		Use:     "reactions",
		Aliases: []string{"reaction"},
		Short:   "Add, remove, and list emoji reactions",
		Cmds: []methodCmd{
			{
				Use: "add", Method: "reactions.add", Kind: kindWrite,
				Short:   "React to a message",
				Example: "  slackctl reactions add --channel C0123456 --ts 1720000000.000100 --name thumbsup",
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
					{Name: "ts", Param: "timestamp", Kind: flagString, Required: true, Usage: "message ts to react to"},
					{Name: "name", Kind: flagString, Required: true, Usage: "emoji name without colons (thumbsup)"},
				},
			},
			{
				Use: "remove", Method: "reactions.remove", Kind: kindWrite,
				Short:   "Remove your reaction from a message",
				Example: "  slackctl reactions remove --channel C0123456 --ts 1720000000.000100 --name thumbsup",
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Usage: "conversation id"},
					{Name: "ts", Param: "timestamp", Kind: flagString, Usage: "message ts"},
					{Name: "name", Kind: flagString, Required: true, Usage: "emoji name without colons"},
				},
			},
			{
				Use: "get", Method: "reactions.get", Kind: kindRead,
				Short:   "Show the reactions on a message",
				Example: "  slackctl reactions get --channel C0123456 --ts 1720000000.000100 --full -o json",
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
					{Name: "ts", Param: "timestamp", Kind: flagString, Required: true, Usage: "message ts"},
					{Name: "full", Kind: flagBool, Usage: "return all reactions, not a summary"},
				},
			},
			{
				Use: "list", Method: "reactions.list", Kind: kindRead,
				Short:     "List items the token's user has reacted to",
				Example:   "  slackctl reactions list --limit 20",
				Paginated: true, ResultKey: "items",
				Columns: []string{"type", "channel"},
				Flags: []flagSpec{
					{Name: "user", Kind: flagString, Usage: "list another user's reactions"},
					{Name: "full", Kind: flagBool, Usage: "return all reactions, not a summary"},
				},
			},
		},
	})
}
