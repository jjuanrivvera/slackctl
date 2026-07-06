## slackctl files list

List files

```
slackctl files list [flags]
```

### Examples

```
  slackctl files list --channel C0123456
  slackctl files list --user U0123456 --types images
```

### Options

```
      --channel string   only files in this conversation
  -h, --help             help for list
      --ts-from string   only files created after this timestamp
      --ts-to string     only files created before this timestamp
      --types string     filter by type: all,spaces,snippets,images,gdocs,zips,pdfs
      --user string      only files from this user
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

* [slackctl files](slackctl_files.md)	 - Upload, download, and manage files

