package commands

// The pins family — channel-pinned messages.

func init() {
	registerGroup(group{
		Use:     "pins",
		Aliases: []string{"pin"},
		Short:   "Pin and unpin messages in a channel",
		Cmds: []methodCmd{
			{
				Use: "list", Method: "pins.list", Kind: kindRead,
				Short:     "List a channel's pinned items",
				Example:   "  slackctl pins list --channel C0123456",
				ResultKey: "items",
				Columns:   []string{"type", "channel", "created"},
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
				},
			},
			{
				Use: "add", Method: "pins.add", Kind: kindWrite,
				Short:   "Pin a message to a channel",
				Example: "  slackctl pins add --channel C0123456 --ts 1720000000.000100",
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
					{Name: "ts", Param: "timestamp", Kind: flagString, Required: true, Usage: "message ts to pin"},
				},
			},
			{
				Use: "remove", Method: "pins.remove", Kind: kindWrite,
				Short:   "Unpin a message from a channel",
				Example: "  slackctl pins remove --channel C0123456 --ts 1720000000.000100",
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
					{Name: "ts", Param: "timestamp", Kind: flagString, Usage: "message ts to unpin"},
				},
			},
		},
	})
}
