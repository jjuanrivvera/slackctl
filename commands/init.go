package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/slackctl/internal/api"
	"github.com/jjuanrivvera/slackctl/internal/auth"
	"github.com/jjuanrivvera/slackctl/internal/config"
)

func init() {
	register(func(root *cobra.Command) {
		cmd := &cobra.Command{
			Use:     "init",
			Aliases: []string{"setup"},
			Short:   "First-run wizard: capture tokens, verify, and save a workspace profile",
			Long: `Interactively set up a workspace profile: paste the bot token (xoxb-), verify it against
auth.test, and optionally add a user token (xoxp-, for search and saved items) and an
app-level token (xapp-, for 'slackctl listen'). Tokens go to the OS keyring.

Create an app and grab tokens at https://api.slack.com/apps (OAuth & Permissions for
xoxb/xoxp; Basic Information → App-Level Tokens for xapp).`,
			Example: `  slackctl init
  slackctl init --workspace acme`,
			RunE: func(cmd *cobra.Command, _ []string) error {
				profileName, cfg, err := resolveProfileName(cmd)
				if err != nil {
					return err
				}
				out := cmd.OutOrStdout()
				fmt.Fprintf(cmd.ErrOrStderr(), "Setting up workspace profile %q.\n", profileName)

				baseFlag, _ := cmd.Flags().GetString("base-url")
				base := config.FirstNonEmpty(baseFlag, api.DefaultBaseURL)
				if err := config.ValidateBaseURL(base); err != nil {
					return err
				}

				token, err := promptSecret(cmd, "Bot token (xoxb-…): ")
				if err != nil {
					return err
				}
				authr, err := api.NewTokenAuth(token)
				if err != nil {
					return err
				}

				client := api.New(authr, api.WithBaseURL(base))
				id, err := client.AuthTest(cmd.Context())
				if err != nil {
					return fmt.Errorf("smoke test failed (token or connectivity): %w", err)
				}

				dir, err := config.Dir()
				if err != nil {
					return err
				}
				store := auth.New(dir)
				if err := store.Set(auth.Key(profileName, auth.KindBot), token); err != nil {
					return err
				}

				// The optional tokens unlock user-only methods and Socket Mode; skipping them
				// keeps the wizard a 30-second path to a working bot.
				userTok, err := promptSecret(cmd, "User token (xoxp-…, optional — enables search/saved; Enter to skip): ")
				if err == nil && userTok != "" {
					if err := store.Set(auth.Key(profileName, auth.KindUser), userTok); err != nil {
						return err
					}
				}
				appTok, err := promptSecret(cmd, "App-level token (xapp-…, optional — enables `slackctl listen`; Enter to skip): ")
				if err == nil && appTok != "" {
					if err := store.Set(auth.Key(profileName, auth.KindApp), appTok); err != nil {
						return err
					}
				}

				if err := cfg.SetProfile(profileName, config.Profile{
					BaseURL:    base,
					AuthMethod: authr.Method(),
					Team:       id.Team,
					TeamID:     id.TeamID,
					UserID:     id.UserID,
					BotID:      id.BotID,
				}); err != nil {
					return err
				}
				cfg.CurrentProfile = profileName
				if err := cfg.Save(); err != nil {
					return err
				}

				fmt.Fprintf(out, "✓ Workspace %q ready — authenticated as %s in %s\n", profileName, id.User, id.Team)
				fmt.Fprintln(out, "  Try: slackctl conversations list")
				return nil
			},
		}
		root.AddCommand(cmd)
	})
}
