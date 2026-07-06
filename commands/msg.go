package commands

import "github.com/spf13/cobra"

// The msg group wraps the chat.* family — posting, editing, deleting, and scheduling
// messages. `msg post --thread-ts` is the thread-reply path (there is no separate method).

func init() {
	registerGroup(group{
		Use:     "msg",
		Aliases: []string{"chat", "message"},
		Short:   "Post, edit, delete, and schedule messages",
		Cmds: []methodCmd{
			{
				Use: "post", Aliases: []string{"send"}, Method: "chat.postMessage", Kind: kindWrite,
				Short: "Post a message to a conversation",
				Long: `Post a message. Reply in a thread with --thread-ts; use --blocks for Block Kit JSON.
Slack recommends keeping --text under 4,000 characters (it truncates at 40,000).`,
				Example: `  slackctl msg post --channel C0123456 --text "deploy finished ✅"
  slackctl msg post --channel C0123456 --thread-ts 1720000000.000100 --text "replying in thread"
  slackctl msg post --channel C0123456 --blocks '[{"type":"section","text":{"type":"mrkdwn","text":"*hi*"}}]'`,
				Columns: []string{"channel", "ts"},
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id (C…/D…/G…)"},
					{Name: "text", Kind: flagString, Usage: "message text (mrkdwn by default)"},
					{Name: "thread-ts", Kind: flagString, Usage: "parent ts — reply in that thread"},
					{Name: "reply-broadcast", Kind: flagBool, Usage: "also show the thread reply in the channel"},
					{Name: "blocks", Kind: flagJSON, Usage: "Block Kit blocks (JSON array)"},
					{Name: "attachments", Kind: flagJSON, Usage: "legacy attachments (JSON array)"},
					{Name: "markdown-text", Kind: flagString, Usage: "full-markdown message body (alternative to --text)"},
					{Name: "unfurl-links", Kind: flagBool, Usage: "unfurl text-based URLs"},
					{Name: "unfurl-media", Kind: flagBool, Usage: "unfurl media URLs"},
					{Name: "username", Kind: flagString, Usage: "override the bot's display name"},
					{Name: "icon-emoji", Kind: flagString, Usage: "override the bot's icon with an emoji (:tada:)"},
					{Name: "icon-url", Kind: flagString, Usage: "override the bot's icon with an image URL"},
					{Name: "metadata", Kind: flagJSON, Usage: "message metadata (JSON object)"},
				},
			},
			{
				Use: "update", Aliases: []string{"edit"}, Method: "chat.update", Kind: kindWrite,
				Short:   "Edit an existing message",
				Example: `  slackctl msg update --channel C0123456 --ts 1720000000.000100 --text "fixed the typo"`,
				Columns: []string{"channel", "ts", "text"},
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
					{Name: "ts", Kind: flagString, Required: true, Usage: "ts of the message to edit"},
					{Name: "text", Kind: flagString, Usage: "new message text"},
					{Name: "blocks", Kind: flagJSON, Usage: "new Block Kit blocks (JSON array)"},
					{Name: "attachments", Kind: flagJSON, Usage: "new legacy attachments (JSON array)"},
					{Name: "markdown-text", Kind: flagString, Usage: "new full-markdown body"},
				},
			},
			{
				Use: "delete", Method: "chat.delete", Kind: kindDestructive,
				Short:   "Delete a message",
				Example: "  slackctl msg delete --channel C0123456 --ts 1720000000.000100",
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
					{Name: "ts", Kind: flagString, Required: true, Usage: "ts of the message to delete"},
				},
			},
			{
				Use: "ephemeral", Method: "chat.postEphemeral", Kind: kindWrite,
				Short:   "Post an ephemeral message (visible to one user)",
				Example: `  slackctl msg ephemeral --channel C0123456 --user U111 --text "only you can see this"`,
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
					{Name: "user", Kind: flagString, Required: true, Usage: "user id who sees the message"},
					{Name: "text", Kind: flagString, Usage: "message text"},
					{Name: "thread-ts", Kind: flagString, Usage: "show inside this thread"},
					{Name: "blocks", Kind: flagJSON, Usage: "Block Kit blocks (JSON array)"},
				},
			},
			{
				Use: "me", Method: "chat.meMessage", Kind: kindWrite,
				Short:   "Post a /me action message",
				Example: `  slackctl msg me --channel C0123456 --text "is deploying"`,
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
					{Name: "text", Kind: flagString, Required: true, Usage: "action text"},
				},
			},
			{
				Use: "permalink", Method: "chat.getPermalink", Kind: kindRead,
				Short:   "Get a message's permalink URL",
				Example: "  slackctl msg permalink --channel C0123456 --ts 1720000000.000100 -o id",
				Columns: []string{"permalink"},
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
					{Name: "ts", Param: "message_ts", Kind: flagString, Required: true, Usage: "message ts"},
				},
			},
			{
				Use: "schedule", Method: "chat.scheduleMessage", Kind: kindWrite,
				Short:   "Schedule a message for later",
				Example: `  slackctl msg schedule --channel C0123456 --post-at 1735689600 --text "happy new year"`,
				Columns: []string{"scheduled_message_id", "channel", "post_at"},
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
					{Name: "post-at", Kind: flagInt, Required: true, Usage: "unix timestamp to post at"},
					{Name: "text", Kind: flagString, Usage: "message text"},
					{Name: "thread-ts", Kind: flagString, Usage: "schedule as a thread reply"},
					{Name: "blocks", Kind: flagJSON, Usage: "Block Kit blocks (JSON array)"},
				},
			},
			{
				Use: "scheduled", Method: "chat.scheduledMessages.list", Kind: kindRead,
				Short:     "List scheduled messages",
				Example:   "  slackctl msg scheduled\n  slackctl msg scheduled --channel C0123456",
				Paginated: true, ResultKey: "scheduled_messages",
				Columns: []string{"id", "channel_id", "post_at", "text"},
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Usage: "only this conversation"},
					{Name: "oldest", Kind: flagString, Usage: "only messages scheduled after this ts"},
					{Name: "latest", Kind: flagString, Usage: "only messages scheduled before this ts"},
				},
			},
			{
				Use: "delete-scheduled", Method: "chat.deleteScheduledMessage", Kind: kindDestructive,
				Short:   "Cancel a scheduled message",
				Example: "  slackctl msg delete-scheduled --channel C0123456 --id Q0123456",
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Required: true, Usage: "conversation id"},
					{Name: "id", Param: "scheduled_message_id", Kind: flagString, Required: true, Usage: "scheduled message id (from `msg scheduled`)"},
				},
			},
		},
		Extra: []func() *cobra.Command{msgTemplateCmd},
	})
}
