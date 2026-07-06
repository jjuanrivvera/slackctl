package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/slackctl/internal/auth"
)

// The users family. `users search` is a documented composite: Slack has no public
// users.search method, so it filters users.list client-side (DECISIONS.md).

func init() {
	registerGroup(group{
		Use:     "users",
		Aliases: []string{"user"},
		Short:   "Look up and list workspace users",
		Cmds: []methodCmd{
			{
				Use: "list", Method: "users.list", Kind: kindRead,
				Short:     "List all workspace users",
				Example:   "  slackctl users list\n  slackctl users list --all -o csv",
				Paginated: true, ResultKey: "members",
				Columns: []string{"id", "name", "real_name", "is_bot", "deleted"},
				Flags: []flagSpec{
					{Name: "include-locale", Kind: flagBool, Usage: "include each user's locale"},
				},
			},
			{
				Use: "info", Method: "users.info", Kind: kindRead,
				Short:     "Show one user",
				Example:   "  slackctl users info --user U0123456",
				ResultKey: "user",
				Flags: []flagSpec{
					{Name: "user", Kind: flagString, Required: true, Usage: "user id (U…)"},
					{Name: "include-locale", Kind: flagBool, Usage: "include the locale"},
				},
			},
			{
				Use: "lookup-email", Aliases: []string{"by-email"}, Method: "users.lookupByEmail", Kind: kindRead,
				Short:     "Find a user by email address",
				Example:   "  slackctl users lookup-email --email ada@example.com",
				ResultKey: "user",
				Columns:   []string{"id", "name", "real_name"},
				Flags: []flagSpec{
					{Name: "email", Kind: flagString, Required: true, Usage: "email address to look up"},
				},
			},
			{
				Use: "conversations", Method: "users.conversations", Kind: kindRead,
				Short:     "List conversations a user is a member of",
				Example:   "  slackctl users conversations\n  slackctl users conversations --user U0123456 --types public_channel,private_channel",
				Paginated: true, ResultKey: "channels",
				Columns: []string{"id", "name", "is_private", "is_im"},
				Flags: []flagSpec{
					{Name: "user", Kind: flagString, Usage: "user id (default: the token's own user/bot)"},
					{Name: "types", Kind: flagString, Usage: "comma-separated: public_channel,private_channel,mpim,im"},
					{Name: "exclude-archived", Kind: flagBool, Usage: "omit archived channels"},
				},
			},
			{
				Use: "presence", Method: "users.getPresence", Kind: kindRead,
				Short:   "Show a user's presence (active/away)",
				Example: "  slackctl users presence --user U0123456",
				Flags: []flagSpec{
					{Name: "user", Kind: flagString, Usage: "user id (default: the token's own user)"},
				},
			},
			{
				Use: "profile", Method: "users.profile.get", Kind: kindRead,
				Short:     "Show a user's profile fields",
				Example:   "  slackctl users profile --user U0123456 -o json",
				ResultKey: "profile",
				Flags: []flagSpec{
					{Name: "user", Kind: flagString, Usage: "user id (default: the token's own user)"},
					{Name: "include-labels", Kind: flagBool, Usage: "include custom-field labels"},
				},
			},
			{
				Use: "set-presence", Method: "users.setPresence", Kind: kindWrite,
				Short:   "Set your presence (auto or away)",
				Example: "  slackctl users set-presence --presence away",
				Flags: []flagSpec{
					{Name: "presence", Kind: flagString, Required: true, Usage: "auto|away"},
				},
			},
		},
		Extra: []func() *cobra.Command{usersSearchCmd, usersSetStatusCmd},
	})
}

// usersSetStatusCmd sets the token owner's custom status (text + emoji + optional expiry) via
// users.profile.set, which takes a profile object rather than flat fields. User-token only.
func usersSetStatusCmd() *cobra.Command {
	var text, emoji string
	var expiration int64
	cmd := &cobra.Command{
		Use:   "set-status",
		Short: "Set your Slack status (text, emoji, expiry)",
		Long: `Set the custom status on your own profile. Clear it by passing empty --text and
--emoji. Needs a user or session token (a bot has no personal status).`,
		Example: `  slackctl users set-status --text "In a meeting" --emoji :calendar:
  slackctl users set-status --text "Lunch" --emoji :taco: --expiration 1735689600
  slackctl users set-status --text "" --emoji ""     # clear`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientForKind(cmd, auth.KindUser)
			if err != nil {
				return err
			}
			profile := map[string]any{
				"status_text":       text,
				"status_emoji":      emoji,
				"status_expiration": expiration,
			}
			raw, err := client.Call(cmd.Context(), "users.profile.set", map[string]any{"profile": profile}, false)
			if err != nil {
				return err
			}
			if raw == nil { // dry-run
				return nil
			}
			return render(cmd, extractField2(raw, "profile"))
		},
	}
	cmd.Flags().StringVar(&text, "text", "", "status text")
	cmd.Flags().StringVar(&emoji, "emoji", "", "status emoji, e.g. :coffee:")
	cmd.Flags().Int64Var(&expiration, "expiration", 0, "unix timestamp when the status clears (0 = never)")
	markKind(cmd, kindWrite)
	return cmd
}

// extractField2 is extractField for hand-written commands (same dotted-key semantics).
func extractField2(raw json.RawMessage, key string) json.RawMessage {
	out, err := extractField(raw, key)
	if err != nil {
		return raw
	}
	return out
}

// usersSearchCmd filters users.list client-side by name / real name / email substring.
func usersSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Find users by name, display name, or email (client-side)",
		Long: `Search users by substring across name, real_name, display_name, and email.
Slack has no public users.search method for bot tokens, so this walks users.list and
filters locally — on very large workspaces prefer 'users lookup-email' when you have
the exact address.`,
		Example: `  slackctl users search ada
  slackctl users search "@example.com" -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromCmd(cmd)
			if err != nil {
				return err
			}
			raw, err := client.CallAllPages(cmd.Context(), "users.list", map[string]any{"limit": 200}, "members", 0)
			if err != nil {
				return err
			}
			if raw == nil { // dry-run
				return nil
			}
			var members []map[string]json.RawMessage
			if err := json.Unmarshal(raw, &members); err != nil {
				return fmt.Errorf("users.list: %w", err)
			}
			q := strings.ToLower(args[0])
			matches := make([]map[string]json.RawMessage, 0, 8)
			for _, m := range members {
				if userMatches(m, q) {
					matches = append(matches, m)
				}
			}
			out, err := json.Marshal(matches)
			if err != nil {
				return err
			}
			if !cmd.Flags().Changed("columns") {
				_ = cmd.Flags().Set("columns", "id,name,real_name,is_bot")
			}
			return render(cmd, out)
		},
	}
	markKind(cmd, kindRead)
	return cmd
}

// userMatches checks the query against the fields people actually search by.
func userMatches(m map[string]json.RawMessage, q string) bool {
	var s struct {
		Name     string `json:"name"`
		RealName string `json:"real_name"`
		Profile  struct {
			DisplayName string `json:"display_name"`
			RealName    string `json:"real_name"`
			Email       string `json:"email"`
		} `json:"profile"`
	}
	b, _ := json.Marshal(m)
	_ = json.Unmarshal(b, &s)
	for _, field := range []string{s.Name, s.RealName, s.Profile.DisplayName, s.Profile.RealName, s.Profile.Email} {
		if field != "" && strings.Contains(strings.ToLower(field), q) {
			return true
		}
	}
	return false
}
