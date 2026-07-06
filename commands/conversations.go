package commands

import "github.com/spf13/cobra"

// The conversations family — Slack's unified channel/DM/group API. Verbs map 1:1 to
// conversations.* methods (api-manifest.json); `unreads` is a documented composite
// (users.conversations + conversations.info), see unreads.go.

func init() {
	registerGroup(group{
		Use:     "conversations",
		Aliases: []string{"conv", "channels"},
		Short:   "Manage channels, DMs, and group conversations",
		Long: `List, inspect, and manage conversations — public/private channels, DMs (im), and
group DMs (mpim). Channel ids look like C…, DMs D…, groups G….`,
		Cmds: []methodCmd{
			{
				Use: "list", Method: "conversations.list", Kind: kindRead,
				Short:     "List conversations the token can see",
				Example:   "  slackctl conversations list\n  slackctl conversations list --types public_channel,private_channel --all\n  slackctl conversations list --types im -o json",
				Paginated: true, ResultKey: "channels",
				Columns: []string{"id", "name", "is_private", "is_archived", "num_members"},
				Flags: []flagSpec{
					{Name: "types", Kind: flagString, Default: "public_channel", Usage: "comma-separated: public_channel,private_channel,mpim,im"},
					{Name: "exclude-archived", Kind: flagBool, Usage: "omit archived channels"},
				},
			},
			{
				Use: "info", Method: "conversations.info", Kind: kindRead,
				Short:     "Show one conversation",
				Example:   "  slackctl conversations info --channel C0123456",
				ResultKey: "channel",
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id (C…/D…/G…)"},
					{Name: "include-locale", Kind: flagBool, Usage: "include the locale"},
					{Name: "include-num-members", Kind: flagBool, Usage: "include the member count"},
				},
			},
			{
				Use: "history", Method: "conversations.history", Kind: kindRead,
				Short:     "Fetch a conversation's message history",
				Example:   "  slackctl conversations history --channel C0123456 --limit 20\n  slackctl conversations history --channel C0123456 --oldest 1720000000.000000 --all -o json",
				Paginated: true, ResultKey: "messages",
				Columns: []string{"ts", "user", "type", "text"},
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
					{Name: "oldest", Kind: flagString, Usage: "only messages after this ts"},
					{Name: "latest", Kind: flagString, Usage: "only messages before this ts"},
					{Name: "inclusive", Kind: flagBool, Usage: "include the oldest/latest ts message itself"},
					{Name: "include-all-metadata", Kind: flagBool, Usage: "include message metadata"},
				},
			},
			{
				Use: "replies", Method: "conversations.replies", Kind: kindRead,
				Short:     "Fetch a thread's replies",
				Example:   "  slackctl conversations replies --channel C0123456 --ts 1720000000.000100",
				Paginated: true, ResultKey: "messages",
				Columns: []string{"ts", "user", "type", "text"},
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
					{Name: "ts", Kind: flagString, Required: true, Usage: "parent message ts (the thread root)"},
					{Name: "oldest", Kind: flagString, Usage: "only replies after this ts"},
					{Name: "latest", Kind: flagString, Usage: "only replies before this ts"},
					{Name: "inclusive", Kind: flagBool, Usage: "include the oldest/latest ts message itself"},
				},
			},
			{
				Use: "members", Method: "conversations.members", Kind: kindRead,
				Short:     "List a conversation's member user ids",
				Example:   "  slackctl conversations members --channel C0123456 --all -o id",
				Paginated: true, ResultKey: "members",
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
				},
			},
			{
				Use: "create", Method: "conversations.create", Kind: kindWrite,
				Short:     "Create a channel",
				Example:   "  slackctl conversations create --name eng-alerts\n  slackctl conversations create --name secret-plans --private",
				ResultKey: "channel",
				Columns:   []string{"id", "name", "is_private"},
				Flags: []flagSpec{
					{Name: "name", Kind: flagString, Required: true, Usage: "channel name (lowercase, no spaces)"},
					{Name: "private", Param: "is_private", Kind: flagBool, Usage: "create a private channel"},
				},
			},
			{
				Use: "rename", Method: "conversations.rename", Kind: kindWrite,
				Short:     "Rename a channel",
				Example:   "  slackctl conversations rename --channel C0123456 --name eng-alerts-v2",
				ResultKey: "channel",
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
					{Name: "name", Kind: flagString, Required: true, Usage: "new channel name"},
				},
			},
			{
				Use: "archive", Method: "conversations.archive", Kind: kindDestructive,
				Short:   "Archive a conversation",
				Example: "  slackctl conversations archive --channel C0123456",
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
				},
			},
			{
				Use: "unarchive", Method: "conversations.unarchive", Kind: kindWrite,
				Short:   "Unarchive a conversation",
				Example: "  slackctl conversations unarchive --channel C0123456",
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
				},
			},
			{
				Use: "invite", Method: "conversations.invite", Kind: kindWrite,
				Short:     "Invite users to a conversation",
				Example:   "  slackctl conversations invite --channel C0123456 --users U111,U222",
				ResultKey: "channel",
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
					{Name: "users", Kind: flagStringSlice, Required: true, Usage: "user ids to invite (comma-separated or repeated)"},
					{Name: "force", Kind: flagBool, Usage: "continue even if some users can't be invited"},
				},
			},
			{
				Use: "join", Method: "conversations.join", Kind: kindWrite,
				Short:     "Join a public channel",
				Example:   "  slackctl conversations join --channel C0123456",
				ResultKey: "channel",
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
				},
			},
			{
				Use: "leave", Method: "conversations.leave", Kind: kindDestructive,
				Short:   "Leave a conversation",
				Example: "  slackctl conversations leave --channel C0123456",
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
				},
			},
			{
				Use: "kick", Method: "conversations.kick", Kind: kindDestructive,
				Short:   "Remove a user from a conversation",
				Example: "  slackctl conversations kick --channel C0123456 --user U111",
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
					{Name: "user", Kind: flagString, Required: true, Usage: "user id to remove"},
				},
			},
			{
				Use: "set-topic", Method: "conversations.setTopic", Kind: kindWrite,
				Short:     "Set a conversation's topic",
				Example:   `  slackctl conversations set-topic --channel C0123456 --topic "Deploys land here"`,
				ResultKey: "channel",
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
					{Name: "topic", Kind: flagString, Required: true, Usage: "new topic text"},
				},
			},
			{
				Use: "set-purpose", Method: "conversations.setPurpose", Kind: kindWrite,
				Short:     "Set a conversation's purpose",
				Example:   `  slackctl conversations set-purpose --channel C0123456 --purpose "Alerting for the eng org"`,
				ResultKey: "channel",
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
					{Name: "purpose", Kind: flagString, Required: true, Usage: "new purpose text"},
				},
			},
			{
				Use: "mark", Method: "conversations.mark", Kind: kindWrite,
				Short:   "Move the read cursor in a conversation",
				Long:    "Sets the read cursor for whoever owns the token: the bot's cursor with the bot token, yours with --as-user.",
				Example: "  slackctl conversations mark --channel C0123456 --ts 1720000000.000100",
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
					{Name: "ts", Kind: flagString, Required: true, Usage: "timestamp of the most recently seen message"},
				},
			},
			{
				Use: "open", Method: "conversations.open", Kind: kindWrite,
				Short:     "Open (or resume) a DM or group DM",
				Example:   "  slackctl conversations open --users U111\n  slackctl conversations open --users U111,U222,U333",
				ResultKey: "channel",
				Columns:   []string{"id"},
				Flags: []flagSpec{
					{Name: "users", Kind: flagStringSlice, Usage: "user ids: one for a DM, several for a group DM"},
					{Name: "channel", Kind: flagString, Usage: "resume an existing im/mpim by id instead"},
					{Name: "return-im", Kind: flagBool, Usage: "include the full im object in the reply"},
				},
			},
			{
				Use: "close", Method: "conversations.close", Kind: kindWrite,
				Short:   "Close a DM or group DM",
				Example: "  slackctl conversations close --channel D0123456",
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "im/mpim id to close"},
				},
			},
		},
		Extra: []func() *cobra.Command{unreadsCmd, conversationsExportCmd},
	})
}
