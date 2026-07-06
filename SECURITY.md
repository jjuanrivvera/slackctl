# Security Policy

## Supported versions

Only the latest release receives security fixes.

## Token handling

- Slack tokens (bot `xoxb-`, user `xoxp-`, app-level `xapp-`) are stored in the OS keyring
  (macOS Keychain, Linux Secret Service, Windows Credential Manager), with an encrypted
  file fallback (`chacha20poly1305`, key held in the config dir with `0600` perms) only
  when no keyring is available.
- Tokens are never written to the config file, logs, or `--dry-run` output; dry-run curls
  redact the `Authorization` header unless `--show-token` is passed explicitly.
- `config view` redacts secrets. Base URLs are validated: cleartext `http://` is rejected
  for non-loopback hosts so a token cannot leak in transit.
- The MCP server never exposes `--show-token`, the workspace selector, `--as-user`, or
  `--base-url` to agents.

## Reporting a vulnerability

Please report vulnerabilities privately via GitHub Security Advisories
(<https://github.com/jjuanrivvera/slackctl/security/advisories/new>). Do not open a public
issue for security reports. You should receive a response within a week.
