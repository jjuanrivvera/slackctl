## slackctl files upload

Upload a file and share it to conversations

### Synopsis

Upload a local file via Slack's external-upload flow (files.upload was sunset in
November 2025) and optionally share it into one or more conversations.

```
slackctl files upload [flags]
```

### Examples

```
  slackctl files upload --file report.pdf --channels C0123456
  slackctl files upload --file diagram.png --channels C0123456,C0456789 --comment "v2"
  slackctl files upload --file snippet.py --channels C0123456 --snippet-type python
```

### Options

```
      --channels string       comma-separated conversation ids to share into
      --comment string        initial comment posted with the file
      --file string           path to the local file to upload
  -h, --help                  help for upload
      --snippet-type string   for code snippets: text, python, go, …
      --thread-ts string      share into this thread
      --title string          file title (defaults to the filename)
```

### Options inherited from parent commands

```
      --as-user                 use the stored user token (xoxp-) instead of the bot token
      --base-url string         Web API base URL (default https://slack.com/api)
      --columns strings         explicit, ordered table/csv columns
      --dry-run                 print the equivalent curl and make no request
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

* [slackctl files](slackctl_files.md)	 - Upload, download, and manage files

