package commands

// The search family. All three methods are user-token-only (bot tokens get
// not_allowed_token_type) and use Slack's classic page/count pagination — a different
// protocol from the cursor walker, so the flags stay explicit (DECISIONS.md).

func searchFlags() []flagSpec {
	return []flagSpec{
		{Name: "query", Kind: flagString, Required: true, Usage: `search query (supports in:#chan, from:@user, "exact phrase")`},
		{Name: "count", Kind: flagInt, Usage: "results per page (max 100)"},
		{Name: "page", Kind: flagInt, Usage: "page number"},
		{Name: "sort", Kind: flagString, Usage: "sort by: score|timestamp"},
		{Name: "sort-dir", Kind: flagString, Usage: "sort direction: asc|desc"},
		{Name: "highlight", Kind: flagBool, Usage: "wrap matches in highlight markers"},
	}
}

func init() {
	registerGroup(group{
		Use:   "search",
		Short: "Search messages and files (needs a user token)",
		Long: `Search across the workspace with Slack's query modifiers (in:#channel, from:@user,
before:/after:, "exact phrase"). Search methods only accept a user token (xoxp-):
slackctl auth login --kind user.`,
		Cmds: []methodCmd{
			{
				Use: "messages", Method: "search.messages", Kind: kindRead, Token: tokenUserRequired,
				Short: "Search messages",
				Example: `  slackctl search messages --query "deploy failed in:#eng-alerts"
  slackctl search messages --query "from:@ada budget" --sort timestamp --sort-dir desc`,
				ResultKey: "messages.matches",
				Columns:   []string{"ts", "username", "text", "permalink"},
				Flags:     searchFlags(),
			},
			{
				Use: "files", Method: "search.files", Kind: kindRead, Token: tokenUserRequired,
				Short:     "Search files",
				Example:   `  slackctl search files --query "quarterly report"`,
				ResultKey: "files.matches",
				Columns:   []string{"id", "name", "title", "user"},
				Flags:     searchFlags(),
			},
			{
				Use: "all", Method: "search.all", Kind: kindRead, Token: tokenUserRequired,
				Short:   "Search messages and files together",
				Example: `  slackctl search all --query "incident 42" -o json`,
				Flags:   searchFlags(),
			},
		},
	})
}
