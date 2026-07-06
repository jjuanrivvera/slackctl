## slackctl agent guard

Generate agent-safety config that blocks destructive slackctl operations

### Synopsis

Classify every API command (read / write / irreversible) from the live command
tree and emit host safety config: irreversible operations (msg delete, msg
delete-scheduled, conversations archive/kick/leave, usergroups disable) are hard-blocked,
ordinary writes require approval, and reads are allowed. Cobra alias paths are covered
too — "chat delete" and "message delete" hit the same rails as "msg delete".

For claude-code the output also includes a PreToolUse hook script
(.claude/hooks/slackctl-guard.sh): it strips quote/backslash obfuscation, matches blocked
subcommand paths at the command position even for path-invoked binaries (./bin/slackctl,
/usr/local/bin/slackctl), and gates the raw "slackctl api <method>" escape hatch — a
method passes only when its final dot-segment is a read shape (get*/list/info/test/
history/replies/members/…), which is how every read in Slack's method naming ends.
"slackctl alias set" is denied so an agent cannot mint a new shorthand for a blocked
command.

MCP-only operation is the hard guarantee; the Bash rails are best-effort — the hook
defeats quoting tricks and path prefixes, but not variable indirection
(a=delete; slackctl msg $a) or shell aliases. Conservative false positives are
accepted: a line that merely QUOTES a blocked command (echo "slackctl msg delete")
is denied.

```
slackctl agent guard --host <claude-code|codex|opencode> [flags]
```

### Examples

```
  slackctl agent guard --host claude-code
  slackctl agent guard --host codex --out ~/.codex/config.toml
  slackctl agent guard --host opencode --all-writes
```

### Options

```
      --all-writes    also hard-block ordinary writes, not just irreversible ops
  -h, --help          help for guard
      --host string   target agent host: claude-code|codex|opencode (required)
      --out string    write to this file instead of stdout
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

* [slackctl agent](slackctl_agent.md)	 - AI-agent integration helpers

