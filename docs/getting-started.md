# Getting started

## Install

=== "Go"

    ```sh
    go install github.com/jjuanrivvera/slackctl/cmd/slackctl@latest
    ```

=== "Homebrew"

    ```sh
    brew install jjuanrivvera/slackctl/slackctl-cli
    ```

=== "From source"

    ```sh
    git clone https://github.com/jjuanrivvera/slackctl
    cd slackctl && make build   # → ./bin/slackctl
    ```

Linux packages (`deb`/`rpm`/`apk`) and a Scoop manifest ship with each
[release](https://github.com/jjuanrivvera/slackctl/releases).

## Authenticate

slackctl needs a Slack credential. The quickest path is the interactive wizard:

```sh
slackctl init
```

It captures a **bot token** (`xoxb-…`), verifies it, and optionally stores a user token and an
app-level token. Tokens go to your OS keyring, never a config file.

Don't have (or want) a Slack app? Use your **browser session** instead — see
[Authentication & tokens](authentication.md#no-slack-app-use-a-browser-session).

Verify anytime:

```sh
slackctl auth status     # identity + workspace
slackctl doctor          # config, credentials, connectivity, clock
```

## Your first commands

```sh
# Discover channels (resolve the C… id you'll use elsewhere)
slackctl conversations list

# Read recent history
slackctl conversations history --channel C0123456 --limit 20

# Post a message (reply in a thread with --thread-ts)
slackctl msg post --channel C0123456 --text "hello from slackctl"

# Look someone up
slackctl users lookup-email --email ada@example.com
```

Every command accepts `-o json|yaml|csv|table|id`, a `--jq` filter, and `--dry-run` (which
prints the equivalent `curl` and makes no request). See [Output & filtering](output.md).

## Multiple workspaces

A profile is a workspace. Log in under a name and switch between them:

```sh
slackctl auth login --workspace acme
slackctl config use acme                 # make it the default
slackctl conversations list --workspace other   # one-off override
```

## Next

- [Authentication & tokens](authentication.md)
- [Output & filtering](output.md)
- [The `listen` command](listen.md)
