## slackctl search

Search messages and files (needs a user token)

### Synopsis

Search across the workspace with Slack's query modifiers (in:#channel, from:@user,
before:/after:, "exact phrase"). Search methods only accept a user token (xoxp-):
slackctl auth login --kind user.

### Options

```
  -h, --help   help for search
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
* [slackctl search all](slackctl_search_all.md)	 - Search messages and files together
* [slackctl search files](slackctl_search_files.md)	 - Search files
* [slackctl search messages](slackctl_search_messages.md)	 - Search messages

