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
	var token, cookie, kindFlag string
	var noVerify bool
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Store a Slack token and verify it",
		Long: `Capture a credential and save it to the keyring for the active workspace profile,
verifying it against auth.test first (except app-level tokens, which can't call it).

Token kinds:
  bot      xoxb-…            OAuth bot token (default; from OAuth & Permissions)
  user     xoxp-…            OAuth user token (search, saved items)
  app      xapp-…            app-level token for 'slackctl listen' (Socket Mode)
  session  xoxc-… + xoxd-…   browser web-client pair (xoxc token + xoxd cookie);
                             no app needed. Acts as your user identity, so it backs
                             bot- and user-kind commands too.`,
		Example: `  slackctl auth login                          # prompt for the bot token (hidden input)
  slackctl auth login --token xoxb-...         # non-interactive
  slackctl auth login --kind user              # store a user token (search, saved items)
  slackctl auth login --kind app               # store an app-level token (slackctl listen)
  slackctl auth login --kind session           # store an xoxc token + xoxd cookie
  slackctl auth login --workspace acme         # store under a named workspace profile`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			kind := auth.TokenKind(kindFlag)
			if !kind.Valid() {
				return fmt.Errorf("invalid --kind %q (want bot|user|app|session)", kindFlag)
			}
			profileName, cfg, err := resolveProfileName(cmd)
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

			if kind == auth.KindSession {
				return runSessionLogin(cmd, cfg, prof, profileName, base, token, cookie, noVerify)
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
			if err := saveProfile(cfg, prof, profileName); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "stored %s token for workspace %q\n", kind, profileName)
			return nil
		},
	}
	cmd.Flags().StringVar(&token, "token", "", "Slack token (omit to be prompted with hidden input)")
	cmd.Flags().StringVar(&cookie, "cookie", "", "xoxd cookie value for --kind session (omit to be prompted)")
	cmd.Flags().StringVar(&kindFlag, "kind", "bot", "token kind: bot|user|app|session")
	cmd.Flags().BoolVar(&noVerify, "no-verify", false, "skip the auth.test verification call")
	return cmd
}

// runSessionLogin captures and stores a browser-session pair (xoxc token + xoxd cookie),
// verifying the pair against auth.test so a wrong/expired session fails at login, not later.
func runSessionLogin(cmd *cobra.Command, cfg *config.Config, prof config.Profile, profileName, base, token, cookie string, noVerify bool) error {
	var err error
	if token == "" {
		token, err = promptSecret(cmd, "Slack session token (xoxc-…): ")
		if err != nil {
			return err
		}
	}
	if cookie == "" {
		cookie, err = promptSecret(cmd, "Slack d cookie (xoxd-…): ")
		if err != nil {
			return err
		}
	}
	authr, err := api.NewSessionAuth(token, cookie)
	if err != nil {
		return err
	}
	if !noVerify {
		client := api.New(authr, api.WithBaseURL(base))
		id, err := client.AuthTest(cmd.Context())
		if err != nil {
			return fmt.Errorf("session verification failed (token or cookie wrong/expired): %w", err)
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "verified as %s in %s (%s)\n", id.User, id.Team, id.TeamID)
		prof.Team = id.Team
		prof.TeamID = id.TeamID
		prof.AuthMethod = authr.Method()
		prof.UserID = id.UserID
	}
	dir, err := config.Dir()
	if err != nil {
		return err
	}
	if err := auth.SetSession(auth.New(dir), profileName, auth.SessionCreds{Token: token, Cookie: cookie}); err != nil {
		return fmt.Errorf("store session credentials: %w", err)
	}
	if err := saveProfile(cfg, prof, profileName); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "stored session credentials for workspace %q\n", profileName)
	return nil
}

// saveProfile writes prof under profileName, making it the current profile if none is set.
func saveProfile(cfg *config.Config, prof config.Profile, profileName string) error {
	if err := cfg.SetProfile(profileName, prof); err != nil {
		return err
	}
	if cfg.CurrentProfile == "" {
		cfg.CurrentProfile = profileName
	}
	return cfg.Save()
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
			kinds := []auth.TokenKind{auth.KindBot, auth.KindUser, auth.KindApp, auth.KindSession}
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
			defer func() { _ = client.Close() }()
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
