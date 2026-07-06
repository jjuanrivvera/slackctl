## slackctl config

Inspect and edit slackctl configuration

### Synopsis

View the config file, switch profiles, and set per-profile options. Secrets live in the keyring and are never shown here.

### Options

```
  -h, --help   help for config
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
* [slackctl config list-profiles](slackctl_config_list-profiles.md)	 - List configured profiles
* [slackctl config path](slackctl_config_path.md)	 - Print the config file path
* [slackctl config set](slackctl_config_set.md)	 - Set a per-profile option (key: base_url)
* [slackctl config use](slackctl_config_use.md)	 - Switch the active profile
* [slackctl config view](slackctl_config_view.md)	 - Show the current configuration (secrets redacted)

