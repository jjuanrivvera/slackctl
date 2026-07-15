## slackctl

Command-line tool for the Slack Web API

### Synopsis

slackctl is a fast, scriptable command-line tool for the Slack Web API.

It wraps the Web API methods (conversations.list, chat.postMessage, search.messages, ...)
behind ergonomic commands with table/json/yaml/csv output, named profiles for multiple
workspaces, OS-keyring token storage, a Socket Mode listener, and an MCP server so AI
agents can drive it safely.

Create an app at https://api.slack.com/apps, install it to your workspace, then:

  slackctl auth login                              # store the bot token in your OS keyring
  slackctl auth status                             # who am I?
  slackctl conversations list                      # channels the token can see
  slackctl msg post --channel C0123456 --text hi   # post a message
  slackctl listen --dms --json                     # stream events (Socket Mode, needs an xapp token)

Every command honors --dry-run (prints the equivalent curl), -o/--output, and --jq.

### Options

```
      --as-user                 use the stored user token (xoxp-) instead of the bot token
      --base-url string         Web API base URL (default https://slack.com/api)
      --columns strings         explicit, ordered table/csv columns
      --dry-run                 print the equivalent curl and make no request
  -h, --help                    help for slackctl
      --jq string               gojq expression applied to the result before rendering
      --no-color                disable colored output
      --no-store slackctl log   do not record messages to the local history store (see slackctl log)
  -o, --output string           output format: table|json|yaml|csv|id (default "table")
      --quiet                   suppress notes on stderr
      --rps float               client-side requests-per-second cap (0 = default)
      --show-token              do not redact the token in --dry-run output
  -v, --verbose                 log raw API responses to stderr
      --workspace string        workspace to use: a named profile/credential (env SLACKCTL_WORKSPACE)
```

### SEE ALSO

* [slackctl agent](slackctl_agent.md)	 - AI-agent integration helpers
* [slackctl alias](slackctl_alias.md)	 - Manage user-defined command aliases
* [slackctl api](slackctl_api.md)	 - Call any Web API method directly (raw escape hatch)
* [slackctl assistant](slackctl_assistant.md)	 - Assistant APIs (bot-token search)
* [slackctl auth](slackctl_auth.md)	 - Manage Slack tokens and verify authentication
* [slackctl bookmarks](slackctl_bookmarks.md)	 - Manage a channel's bookmarks
* [slackctl canvases](slackctl_canvases.md)	 - Create and manage Canvases
* [slackctl completion](slackctl_completion.md)	 - Generate a shell completion script
* [slackctl config](slackctl_config.md)	 - Inspect and edit slackctl configuration
* [slackctl conversations](slackctl_conversations.md)	 - Manage channels, DMs, and group conversations
* [slackctl dnd](slackctl_dnd.md)	 - Do Not Disturb — snooze and status
* [slackctl doctor](slackctl_doctor.md)	 - Diagnose configuration, credentials, and connectivity
* [slackctl emoji](slackctl_emoji.md)	 - Custom emoji
* [slackctl files](slackctl_files.md)	 - Upload, download, and manage files
* [slackctl init](slackctl_init.md)	 - First-run wizard: capture tokens, verify, and save a workspace profile
* [slackctl listen](slackctl_listen.md)	 - Stream events live (Socket Mode or RTM) as lines
* [slackctl log](slackctl_log.md)	 - Search your local Slack message history
* [slackctl mcp](slackctl_mcp.md)	 - MCP server management
* [slackctl msg](slackctl_msg.md)	 - Post, edit, delete, and schedule messages
* [slackctl pins](slackctl_pins.md)	 - Pin and unpin messages in a channel
* [slackctl reactions](slackctl_reactions.md)	 - Add, remove, and list emoji reactions
* [slackctl saved](slackctl_saved.md)	 - Saved items (legacy stars API; needs a user token)
* [slackctl search](slackctl_search.md)	 - Search messages and files (needs a user token)
* [slackctl team](slackctl_team.md)	 - Workspace information
* [slackctl update](slackctl_update.md)	 - Update slackctl to the latest release
* [slackctl usergroups](slackctl_usergroups.md)	 - Manage user groups (@mention groups)
* [slackctl users](slackctl_users.md)	 - Look up and list workspace users
* [slackctl version](slackctl_version.md)	 - Print version, commit, and build date

