package commands

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// classification buckets every API command by safety level, derived from the live command
// tree's annotations (so it stays correct as commands are added/removed).
type classification struct {
	Read        []apiCmdInfo
	Write       []apiCmdInfo
	Destructive []apiCmdInfo
}

// classifyAPICommands splits the registered API commands into read/write/destructive. When
// allWrites is true, ordinary writes are promoted to the destructive (hard-block) bucket.
func classifyAPICommands(allWrites bool) classification {
	var c classification
	cmds := append([]apiCmdInfo{}, registeredAPICmds...)
	sort.Slice(cmds, func(i, j int) bool { return cmds[i].Path < cmds[j].Path })
	for _, cmd := range cmds {
		switch cmd.Kind {
		case kindRead:
			c.Read = append(c.Read, cmd)
		case kindWrite:
			if allWrites {
				c.Destructive = append(c.Destructive, cmd)
			} else {
				c.Write = append(c.Write, cmd)
			}
		case kindDestructive:
			c.Destructive = append(c.Destructive, cmd)
		}
	}
	return c
}

func init() {
	register(func(root *cobra.Command) {
		agentCmd := &cobra.Command{
			Use:   "agent",
			Short: "AI-agent integration helpers",
			Long:  "Generate safety configuration for AI agents that drive slackctl.",
		}

		var host, out string
		var allWrites bool
		guard := &cobra.Command{
			Use:   "guard --host <claude-code|codex|opencode>",
			Short: "Generate agent-safety config that blocks destructive slackctl operations",
			Long: `Classify every API command (read / write / irreversible) from the live command
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
is denied.`,
			Example: `  slackctl agent guard --host claude-code
  slackctl agent guard --host codex --out ~/.codex/config.toml
  slackctl agent guard --host opencode --all-writes`,
			RunE: func(cmd *cobra.Command, _ []string) error {
				cls := classifyAPICommands(allWrites)
				var content string
				var err error
				switch host {
				case "claude-code", "claude":
					content, err = renderClaudeCode(cls)
				case "codex":
					content, err = renderCodex(cls)
				case "opencode":
					content, err = renderOpenCode(cls)
				default:
					return fmt.Errorf("unknown --host %q (want claude-code|codex|opencode)", host)
				}
				if err != nil {
					return err
				}
				if out != "" {
					if err := os.WriteFile(out, []byte(content), 0o600); err != nil {
						return err
					}
					fmt.Fprintf(cmd.ErrOrStderr(), "wrote %s safety config to %s\n", host, out)
					return nil
				}
				fmt.Fprint(cmd.OutOrStdout(), content)
				return nil
			},
		}
		guard.Flags().StringVar(&host, "host", "", "target agent host: claude-code|codex|opencode (required)")
		guard.Flags().StringVar(&out, "out", "", "write to this file instead of stdout")
		guard.Flags().BoolVar(&allWrites, "all-writes", false, "also hard-block ordinary writes, not just irreversible ops")
		_ = guard.MarkFlagRequired("host")

		agentCmd.AddCommand(guard)
		root.AddCommand(agentCmd)
	})
}

// bashPattern is the Claude-Code/OpenCode Bash permission pattern for a command path.
func bashPattern(path string) string { return "Bash(slackctl " + path + ":*)" }

// mcpToolPattern is the MCP tool name a host gates: slack_<group>_<verb> under the
// slackctl MCP server.
func mcpToolPattern(path string) string {
	return "mcp__slackctl__slack_" + strings.ReplaceAll(path, " ", "_")
}
