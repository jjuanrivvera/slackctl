package commands

import (
	"slices"

	"github.com/njayp/ophis"
	"github.com/spf13/cobra"
)

// excludedFromMCP are command-name substrings kept out of the MCP tool surface: setup/meta
// commands an agent should not drive, and the raw `api` escape hatch (which would bypass the
// per-command read-only/write/destructive annotations). The `mcp` and `agent` subtrees are
// excluded too so an agent can neither re-enter the server nor disable its own guardrails.
var excludedFromMCP = []string{
	"agent", "auth", "config", "alias", "init", "doctor", "completion", "version", "api", "log",
	// `listen` is a long-running stream — never expose a blocking command as a tool an
	// agent could call and hang on.
	"listen",
}

// selfUpdatePaths are the EXACT command paths of the self-update command. `update` replaces
// the running binary, so it must never be an MCP tool — but it can't go in excludedFromMCP,
// whose substring match would also drop the real `msg update` / `usergroups update` API
// commands. Excluded by exact path instead.
var selfUpdatePaths = []string{"slackctl update", "slackctl update check"}

// secretFlags must never reach the MCP tool schema: an agent must not read the token or
// switch workspaces. The server uses whatever workspace/profile is active at startup. Both
// the --workspace flag and its hidden --profile alias are excluded, as is --as-user (token
// escalation: bot → human identity).
var secretFlags = []string{"show-token", "workspace", "profile", "base-url", "as-user"}

func init() {
	register(func(root *cobra.Command) {
		excludeMeta := ophis.ExcludeCmdsContaining(excludedFromMCP...)
		// ophis walks the command tree and exposes each runnable leaf as an MCP tool, replaying
		// the cobra command on invocation so tools reuse the same client, keyring, and profile.
		root.AddCommand(ophis.Command(&ophis.Config{
			ToolNamePrefix: "slack",
			Selectors: []ophis.Selector{{
				CmdSelector: func(cmd *cobra.Command) bool {
					return excludeMeta(cmd) && !slices.Contains(selfUpdatePaths, cmd.CommandPath())
				},
				InheritedFlagSelector: ophis.ExcludeFlags(secretFlags...),
			}},
		}))
	})
}
