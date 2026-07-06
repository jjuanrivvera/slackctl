package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/slackctl/internal/api"
)

// conversationsExportCmd is a beyond-the-API value-add: it walks a conversation's full
// history (and, with --threads, each thread's replies) and writes every message as one JSON
// line to stdout or a file — a git-friendly, greppable archive Slack's UI can't produce.
func conversationsExportCmd() *cobra.Command {
	var channel, oldest, latest, out string
	var withThreads bool
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export a conversation's full history to JSONL",
		Long: `Walk a conversation's entire message history (paginating over every page) and write
each message as one JSON object per line. With --threads, each threaded message is followed
by its replies. Bound the range with --oldest/--latest (unix or Slack ts).`,
		Example: `  slackctl conversations export --channel C0123456 > history.jsonl
  slackctl conversations export --channel C0123456 --threads --out backup.jsonl
  slackctl conversations export --channel C0123456 --oldest 1720000000.000000`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromCmd(cmd)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			if out != "" {
				f, err := os.Create(out) //nolint:gosec // G304: out is the user's chosen --out path
				if err != nil {
					return err
				}
				defer func() { _ = f.Close() }()
				w = f
			}

			params := map[string]any{"channel": channel, "limit": 200}
			if oldest != "" {
				params["oldest"] = oldest
			}
			if latest != "" {
				params["latest"] = latest
			}
			raw, err := client.CallAllPages(cmd.Context(), "conversations.history", params, "messages", 0)
			if err != nil {
				return err
			}
			if raw == nil { // dry-run
				return nil
			}
			var messages []json.RawMessage
			if err := json.Unmarshal(raw, &messages); err != nil {
				return fmt.Errorf("conversations.history: %w", err)
			}

			count := 0
			for _, m := range messages {
				if _, err := fmt.Fprintln(w, string(m)); err != nil {
					return err
				}
				count++
				if withThreads {
					n, err := exportThread(cmd, client, channel, m, w)
					if err != nil {
						return err
					}
					count += n
				}
			}
			if out != "" {
				fmt.Fprintf(cmd.ErrOrStderr(), "exported %d messages to %s\n", count, out)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&channel, "channel", "", "conversation id to export (C…/D…/G…)")
	cmd.Flags().StringVar(&oldest, "oldest", "", "only messages after this ts")
	cmd.Flags().StringVar(&latest, "latest", "", "only messages before this ts")
	cmd.Flags().BoolVar(&withThreads, "threads", false, "also export each thread's replies")
	cmd.Flags().StringVar(&out, "out", "", "write to this file instead of stdout")
	_ = cmd.MarkFlagRequired("channel")
	markKind(cmd, kindRead)
	return cmd
}

// exportThread writes the replies of a threaded parent message (if any). It returns how many
// reply lines it wrote. A message is a thread parent when its thread_ts equals its ts and it
// has a reply_count > 0; we detect that from the raw message.
func exportThread(cmd *cobra.Command, client *api.Client, channel string, msg json.RawMessage, w io.Writer) (int, error) {
	var meta struct {
		TS         string `json:"ts"`
		ThreadTS   string `json:"thread_ts"`
		ReplyCount int    `json:"reply_count"`
	}
	_ = json.Unmarshal(msg, &meta)
	// Only fetch replies for a thread ROOT (thread_ts == ts) that actually has replies —
	// otherwise every message would trigger a redundant conversations.replies call.
	if meta.ReplyCount == 0 || (meta.ThreadTS != "" && meta.ThreadTS != meta.TS) {
		return 0, nil
	}
	raw, err := client.CallAllPages(cmd.Context(), "conversations.replies",
		map[string]any{"channel": channel, "ts": meta.TS, "limit": 200}, "messages", 0)
	if err != nil {
		return 0, err
	}
	var replies []json.RawMessage
	if err := json.Unmarshal(raw, &replies); err != nil {
		return 0, err
	}
	count := 0
	for _, r := range replies {
		// conversations.replies includes the parent as the first item; skip it (already written).
		var rmeta struct {
			TS string `json:"ts"`
		}
		_ = json.Unmarshal(r, &rmeta)
		if rmeta.TS == meta.TS {
			continue
		}
		if _, err := fmt.Fprintln(w, string(r)); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}
