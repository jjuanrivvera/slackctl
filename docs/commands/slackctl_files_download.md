## slackctl files download

Download a file's contents

### Synopsis

Resolve a file's private URL and stream its bytes to --out (or ./<name> by default).

```
slackctl files download [flags]
```

### Examples

```
  slackctl files download --file F0123456
  slackctl files download --file F0123456 --out ~/Downloads/report.pdf
```

### Options

```
      --file string   file id to download (F…)
  -h, --help          help for download
      --out string    destination path (default: the file's name in the cwd)
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

