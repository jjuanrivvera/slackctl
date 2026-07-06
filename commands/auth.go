package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/slackctl/internal/api"
	"github.com/jjuanrivvera/slackctl/internal/auth"
	"github.com/jjuanrivvera/slackctl/internal/config"
)

func init() {
	register(func(root *cobra.Command) {
		authCmd := &cobra.Command{
			Use:   "auth",
			Short: "Manage Slack tokens and verify authentication",
			Long: `Capture, verify, and remove the tokens for a workspace profile. Tokens are stored in
your OS keyring, never in the config file.

A workspace profile can hold three token kinds:
  bot   xoxb-…  drives most commands (default)
  user  xoxp-…  unlocks user-only methods (search, saved items; use --as-user elsewhere)
  app   xapp-…  opens Socket Mode connections for 'slackctl listen'`,
		}
		authCmd.AddCommand(authLoginCmd(), authLogoutCmd(), authStatusCmd())
		root.AddCommand(authCmd)
	})
}

func authLoginCmd() *cobra.Command {
	var token, kindFlag string
	var noVerify bool
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Store a Slack token and verify it",
		Long: `Capture a token from your Slack app (https://api.slack.com/apps → OAuth & Permissions),
verify it against auth.test, and save it to the keyring for the active workspace profile.`,
		Example: `  slackctl auth login                          # prompt for the bot token (hidden input)
  slackctl auth login --token xoxb-...         # non-interactive
  slackctl auth login --kind user              # store a user token (search, saved items)
  slackctl auth login --kind app               # store an app-level token (slackctl listen)
  slackctl auth login --workspace acme         # store under a named workspace profile`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			kind := auth.TokenKind(kindFlag)
			if !kind.Valid() {
				return fmt.Errorf("invalid --kind %q (want bot|user|app)", kindFlag)
			}
			profileName, cfg, err := resolveProfileName(cmd)
			if err != nil {
				return err
			}
			if token == "" {
				token, err = promptSecret(cmd, fmt.Sprintf("Slack %s token: ", kind))
				if err != nil {
					return err
				}
			}
			authr, err := api.NewTokenAuth(token)
			if err != nil {
				return err
			}

			baseFlag, _ := cmd.Flags().GetString("base-url")
			base := config.FirstNonEmpty(baseFlag, api.DefaultBaseURL)
			if err := config.ValidateBaseURL(base); err != nil {
				return err
			}

			prof, _ := cfg.Profile(profileName)
			prof.BaseURL = base
			// App-level tokens can't call auth.test; their verification is opening a Socket
			// Mode connection, which `slackctl listen` does anyway. Verify bot/user tokens.
			if !noVerify && kind != auth.KindApp {
				client := api.New(authr, api.WithBaseURL(base))
				id, err := client.AuthTest(cmd.Context())
				if err != nil {
					return fmt.Errorf("token verification failed: %w", err)
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "verified as %s in %s (%s)\n", id.User, id.Team, id.TeamID)
				prof.Team = id.Team
				prof.TeamID = id.TeamID
				if kind == auth.KindBot {
					prof.AuthMethod = authr.Method()
					prof.UserID = id.UserID
					prof.BotID = id.BotID
				}
			}

			dir, err := config.Dir()
			if err != nil {
				return err
			}
			if err := auth.New(dir).Set(auth.Key(profileName, kind), token); err != nil {
				return fmt.Errorf("store token: %w", err)
			}
			if err := cfg.SetProfile(profileName, prof); err != nil {
				return err
			}
			if cfg.CurrentProfile == "" {
				cfg.CurrentProfile = profileName
			}
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "stored %s token for workspace %q\n", kind, profileName)
			return nil
		},
	}
	cmd.Flags().StringVar(&token, "token", "", "Slack token (omit to be prompted with hidden input)")
	cmd.Flags().StringVar(&kindFlag, "kind", "bot", "token kind: bot|user|app")
	cmd.Flags().BoolVar(&noVerify, "no-verify", false, "skip the auth.test verification call")
	return cmd
}

func authLogoutCmd() *cobra.Command {
	var kindFlag string
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Remove stored tokens for the active workspace",
		Example: `  slackctl auth logout               # remove all tokens for the workspace
  slackctl auth logout --kind user   # remove only the user token`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			profileName, _, err := resolveProfileName(cmd)
			if err != nil {
				return err
			}
			dir, err := config.Dir()
			if err != nil {
				return err
			}
			store := auth.New(dir)
			kinds := []auth.TokenKind{auth.KindBot, auth.KindUser, auth.KindApp}
			if kindFlag != "" {
				kind := auth.TokenKind(kindFlag)
				if !kind.Valid() {
					return fmt.Errorf("invalid --kind %q (want bot|user|app)", kindFlag)
				}
				kinds = []auth.TokenKind{kind}
			}
			var removed int
			for _, k := range kinds {
				if err := store.Delete(auth.Key(profileName, k)); err == nil {
					removed++
				}
			}
			if removed == 0 && kindFlag != "" {
				return fmt.Errorf("no %s token stored for workspace %q", kindFlag, profileName)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "logged out of workspace %q\n", profileName)
			return nil
		},
	}
	cmd.Flags().StringVar(&kindFlag, "kind", "", "remove only this token kind: bot|user|app")
	return cmd
}

func authStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status",
		Aliases: []string{"whoami"},
		Short:   "Show the active workspace, base URL, and token validity",
		Example: `  slackctl auth status
  slackctl auth whoami -o json
  slackctl auth status --as-user   # check the stored user token instead`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			profileName, cfg, err := resolveProfileName(cmd)
			if err != nil {
				return err
			}
			prof, _ := cfg.Profile(profileName)
			base := config.FirstNonEmpty(prof.BaseURL, api.DefaultBaseURL)

			// auth status is a real check: if the token is missing or invalid it exits
			// non-zero (so `slackctl auth status && …` works), while still printing why.
			client, err := clientFromCmd(cmd)
			if err != nil {
				return fmt.Errorf("not authenticated (workspace %q): %w", profileName, err)
			}
			id, err := client.AuthTest(cmd.Context())
			if err != nil {
				return fmt.Errorf("token invalid for workspace %q: %w", profileName, err)
			}
			status := map[string]any{
				"workspace": profileName,
				"base_url":  base,
				"valid":     true,
				"team":      id.Team,
				"team_id":   id.TeamID,
				"user":      id.User,
				"user_id":   id.UserID,
			}
			if id.BotID != "" {
				status["bot_id"] = id.BotID
			}
			return render(cmd, mustJSON(status))
		},
	}
	return cmd
}

func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
