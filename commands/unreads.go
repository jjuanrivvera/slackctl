package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// unreadsCmd is the documented composite behind `conversations unreads`: Slack has no
// single "unreads" endpoint, so it walks the token's conversation memberships
// (users.conversations) and asks conversations.info for each one's unread_count
// (DECISIONS.md). Meaningful counts need a read cursor, which bots only have where they
// mark; a user token (--as-user) reflects what the human actually hasn't read.
func unreadsCmd() *cobra.Command {
	var types string
	var includeZero bool
	cmd := &cobra.Command{
		Use:   "unreads",
		Short: "Show unread counts across the token's conversations",
		Long: `List the conversations the token is a member of with their unread message counts
(relative to the token owner's read cursor — use --as-user for YOUR unreads rather
than the bot's). One conversations.info call is made per membership, so on very large
membership lists expect it to take a few seconds under Slack's rate limits.`,
		Example: `  slackctl conversations unreads --as-user
  slackctl conversations unreads --types im,mpim
  slackctl conversations unreads --include-zero -o json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromCmd(cmd)
			if err != nil {
				return err
			}
			defer func() { _ = client.Close() }()
			params := map[string]any{"limit": 200}
			if types != "" {
				params["types"] = types
			}
			raw, err := client.CallAllPages(cmd.Context(), "users.conversations", params, "channels", 0)
			if err != nil {
				return err
			}
			if raw == nil { // dry-run prints the representative request and stops
				return nil
			}
			var memberships []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				User string `json:"user"` // DMs carry the counterpart instead of a name
				IsIM bool   `json:"is_im"`
			}
			if err := json.Unmarshal(raw, &memberships); err != nil {
				return fmt.Errorf("users.conversations: %w", err)
			}

			type unread struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Unread      int64  `json:"unread_count"`
				LastRead    string `json:"last_read,omitempty"`
				IsIM        bool   `json:"is_im"`
				CounterUser string `json:"user,omitempty"`
			}
			rows := make([]unread, 0, len(memberships))
			for _, m := range memberships {
				var info struct {
					Channel struct {
						UnreadCount        int64  `json:"unread_count"`
						UnreadCountDisplay int64  `json:"unread_count_display"`
						LastRead           string `json:"last_read"`
					} `json:"channel"`
				}
				err := client.CallInto(cmd.Context(), "conversations.info",
					map[string]any{"channel": m.ID}, true, &info)
				if err != nil {
					return fmt.Errorf("conversations.info %s: %w", m.ID, err)
				}
				count := info.Channel.UnreadCountDisplay
				if count == 0 {
					count = info.Channel.UnreadCount
				}
				if count == 0 && !includeZero {
					continue
				}
				rows = append(rows, unread{
					ID: m.ID, Name: m.Name, Unread: count,
					LastRead: info.Channel.LastRead, IsIM: m.IsIM, CounterUser: m.User,
				})
			}
			out, err := json.Marshal(rows)
			if err != nil {
				return err
			}
			if !cmd.Flags().Changed("columns") {
				_ = cmd.Flags().Set("columns", "id,name,unread_count,user")
			}
			return render(cmd, out)
		},
	}
	cmd.Flags().StringVar(&types, "types", "", "conversation types to check: public_channel,private_channel,mpim,im")
	cmd.Flags().BoolVar(&includeZero, "include-zero", false, "also list conversations with nothing unread")
	markKind(cmd, kindRead)
	return cmd
}
