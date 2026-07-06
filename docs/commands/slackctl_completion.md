## slackctl completion

Generate a shell completion script

### Synopsis

Output a completion script for your shell. See `slackctl completion <shell> --help` for install instructions.

```
slackctl completion [bash|zsh|fish|powershell]
```

### Examples

```
  source <(slackctl completion bash)
  slackctl completion zsh > "${fpath[1]}/_slackctl"
  slackctl completion fish > ~/.config/fish/completions/slackctl.fish
```

### Options

```
  -h, --help   help for completion
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

