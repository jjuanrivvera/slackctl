## slackctl files

Upload, download, and manage files

### Options

```
  -h, --help   help for files
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
* [slackctl files delete](slackctl_files_delete.md)	 - Delete a file
* [slackctl files download](slackctl_files_download.md)	 - Download a file's contents
* [slackctl files info](slackctl_files_info.md)	 - Show a file's metadata
* [slackctl files list](slackctl_files_list.md)	 - List files
* [slackctl files upload](slackctl_files_upload.md)	 - Upload a file and share it to conversations

