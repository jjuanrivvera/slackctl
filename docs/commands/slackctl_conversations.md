## slackctl conversations

Manage channels, DMs, and group conversations

### Synopsis

List, inspect, and manage conversations — public/private channels, DMs (im), and
group DMs (mpim). Channel ids look like C…, DMs D…, groups G….

### Options

```
  -h, --help   help for conversations
```

### Options inherited from parent commands

```
      --as-user            use the stored user token (xoxp-) instead of the bot token
      --base-url string    Web API base URL (default https://slack.com/api)
      --columns strings    explicit, ordered table/csv columns
      --dry-run            print the equivalent curl and make no request
      --jq string          gojq expression applied to the result before rendering
      --no-color           disable colored output
  -o, --output string      output format: table|json|yaml|csv|id (default "table")
      --quiet              suppress notes on stderr
      --rps float          client-side requests-per-second cap (0 = default)
      --show-token         do not redact the token in --dry-run output
  -v, --verbose            log raw API responses to stderr
      --workspace string   workspace to use: a named profile/credential (env SLACKCTL_WORKSPACE)
```

### SEE ALSO

* [slackctl](slackctl.md)	 - Command-line tool for the Slack Web API
* [slackctl conversations archive](slackctl_conversations_archive.md)	 - Archive a conversation
* [slackctl conversations close](slackctl_conversations_close.md)	 - Close a DM or group DM
* [slackctl conversations create](slackctl_conversations_create.md)	 - Create a channel
* [slackctl conversations export](slackctl_conversations_export.md)	 - Export a conversation's full history to JSONL
* [slackctl conversations history](slackctl_conversations_history.md)	 - Fetch a conversation's message history
* [slackctl conversations info](slackctl_conversations_info.md)	 - Show one conversation
* [slackctl conversations invite](slackctl_conversations_invite.md)	 - Invite users to a conversation
* [slackctl conversations join](slackctl_conversations_join.md)	 - Join a public channel
* [slackctl conversations kick](slackctl_conversations_kick.md)	 - Remove a user from a conversation
* [slackctl conversations leave](slackctl_conversations_leave.md)	 - Leave a conversation
* [slackctl conversations list](slackctl_conversations_list.md)	 - List conversations the token can see
* [slackctl conversations mark](slackctl_conversations_mark.md)	 - Move the read cursor in a conversation
* [slackctl conversations members](slackctl_conversations_members.md)	 - List a conversation's member user ids
* [slackctl conversations open](slackctl_conversations_open.md)	 - Open (or resume) a DM or group DM
* [slackctl conversations rename](slackctl_conversations_rename.md)	 - Rename a channel
* [slackctl conversations replies](slackctl_conversations_replies.md)	 - Fetch a thread's replies
* [slackctl conversations set-purpose](slackctl_conversations_set-purpose.md)	 - Set a conversation's purpose
* [slackctl conversations set-topic](slackctl_conversations_set-topic.md)	 - Set a conversation's topic
* [slackctl conversations unarchive](slackctl_conversations_unarchive.md)	 - Unarchive a conversation
* [slackctl conversations unreads](slackctl_conversations_unreads.md)	 - Show unread counts across the token's conversations

