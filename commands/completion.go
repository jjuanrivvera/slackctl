package commands

import (
	"github.com/spf13/cobra"
)

func init() {
	register(func(root *cobra.Command) {
		cmd := &cobra.Command{
			Use:                   "completion [bash|zsh|fish|powershell]",
			Short:                 "Generate a shell completion script",
			Long:                  "Output a completion script for your shell. See `slackctl completion <shell> --help` for install instructions.",
			DisableFlagsInUseLine: true,
			ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
			Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
			Example: `  source <(slackctl completion bash)
  slackctl completion zsh > "${fpath[1]}/_slackctl"
  slackctl completion fish > ~/.config/fish/completions/slackctl.fish`,
			RunE: func(cmd *cobra.Command, args []string) error {
				out := cmd.OutOrStdout()
				switch args[0] {
				case "bash":
					return root.GenBashCompletionV2(out, true)
				case "zsh":
					return root.GenZshCompletion(out)
				case "fish":
					return root.GenFishCompletion(out, true)
				case "powershell":
					return root.GenPowerShellCompletionWithDesc(out)
				}
				return nil
			},
		}
		root.AddCommand(cmd)
	})
}
