package commands

// The bookmarks family — the links pinned to a channel's header.

func init() {
	registerGroup(group{
		Use:     "bookmarks",
		Aliases: []string{"bookmark"},
		Short:   "Manage a channel's bookmarks",
		Cmds: []methodCmd{
			{
				Use: "list", Method: "bookmarks.list", Kind: kindRead,
				Short:     "List a channel's bookmarks",
				Example:   "  slackctl bookmarks list --channel C0123456",
				ResultKey: "bookmarks",
				Columns:   []string{"id", "title", "link", "type"},
				Flags: []flagSpec{
					{Name: "channel", Param: "channel_id", Kind: flagString, Required: true, Usage: "conversation id"},
				},
			},
			{
				Use: "add", Method: "bookmarks.add", Kind: kindWrite,
				Short:     "Add a link bookmark to a channel",
				Example:   `  slackctl bookmarks add --channel C0123456 --title "Runbook" --link https://wiki/runbook`,
				ResultKey: "bookmark",
				Columns:   []string{"id", "title", "link"},
				Flags: []flagSpec{
					{Name: "channel", Param: "channel_id", Kind: flagString, Required: true, Usage: "conversation id"},
					{Name: "title", Kind: flagString, Required: true, Usage: "bookmark title"},
					{Name: "type", Kind: flagString, Default: "link", Usage: "bookmark type (link)"},
					{Name: "link", Kind: flagString, Usage: "URL to bookmark"},
					{Name: "emoji", Kind: flagString, Usage: "emoji icon, e.g. :book:"},
				},
			},
			{
				Use: "edit", Method: "bookmarks.edit", Kind: kindWrite,
				Short:     "Edit a channel bookmark",
				Example:   `  slackctl bookmarks edit --channel C0123456 --bookmark Bk123 --title "New Runbook"`,
				ResultKey: "bookmark",
				Flags: []flagSpec{
					{Name: "channel", Param: "channel_id", Kind: flagString, Required: true, Usage: "conversation id"},
					{Name: "bookmark", Param: "bookmark_id", Kind: flagString, Required: true, Usage: "bookmark id (Bk…)"},
					{Name: "title", Kind: flagString, Usage: "new title"},
					{Name: "link", Kind: flagString, Usage: "new URL"},
					{Name: "emoji", Kind: flagString, Usage: "new emoji icon"},
				},
			},
			{
				Use: "remove", Method: "bookmarks.remove", Kind: kindDestructive,
				Short:   "Remove a channel bookmark",
				Example: "  slackctl bookmarks remove --channel C0123456 --bookmark Bk123",
				Flags: []flagSpec{
					{Name: "channel", Param: "channel_id", Kind: flagString, Required: true, Usage: "conversation id"},
					{Name: "bookmark", Param: "bookmark_id", Kind: flagString, Required: true, Usage: "bookmark id to remove"},
				},
			},
		},
	})
}
