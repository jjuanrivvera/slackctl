package commands

// assistant.search.context is Slack's newer search endpoint and, unlike search.*, it accepts
// a BOT token — so it gives bot-driven setups a search capability without a user token.

func init() {
	registerGroup(group{
		Use:   "assistant",
		Short: "Assistant APIs (bot-token search)",
		Cmds: []methodCmd{
			{
				Use: "search-context", Aliases: []string{"search"}, Method: "assistant.search.context", Kind: kindRead,
				Short: "Search messages and files (works with a bot token)",
				Long: `Search across the workspace via the Assistant search API. Unlike 'slackctl search',
this accepts a bot token, so it works without a user/session credential.`,
				Example: `  slackctl assistant search-context --query "deploy failed"
  slackctl assistant search-context --query "incident" --channel-types public_channel --limit 10`,
				ResultKey: "results.messages",
				Columns:   []string{"ts", "channel_id", "author_user_id", "message"},
				Flags: []flagSpec{
					{Name: "query", Kind: flagString, Required: true, Usage: "search query"},
					{Name: "channel-types", Kind: flagString, Usage: "comma-separated: public_channel,private_channel,mpim,im"},
					{Name: "content-types", Kind: flagString, Usage: "comma-separated: messages,files"},
					{Name: "context-channel-id", Kind: flagString, Usage: "bias results toward this channel"},
					{Name: "include-bots", Kind: flagBool, Usage: "include messages from bots"},
					{Name: "after", Kind: flagString, Usage: "only results after this date/ts"},
					{Name: "before", Kind: flagString, Usage: "only results before this date/ts"},
				},
			},
		},
	})
}
