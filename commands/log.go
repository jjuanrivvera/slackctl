package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/slackctl/internal/config"
	"github.com/jjuanrivvera/slackctl/internal/store"
)

// logDefaultColumns orders the wide table sensibly; -o json carries every field, and
// --columns overrides this like any other command.
var logDefaultColumns = []string{"ts", "channel", "user", "type", "text"}

func init() {
	register(func(root *cobra.Command) {
		var channel, user, since string
		var limit int

		logCmd := &cobra.Command{
			Use:   "log",
			Short: "Search your local Slack message history",
			Long: `slackctl records the messages that flow through it — every post, every page of
history/replies you fetch, and (with 'listen') every streamed event — into a per-workspace
SQLite database. 'log' searches that store locally: instantly, offline, and without Slack's
user-token-only, rate-limited search API.

Recording is on by default; disable it for a call with --no-store, or globally by never
running message-bearing commands. This command only READS (it never records), so it works
regardless of --no-store.`,
			Example: `  slackctl log                                   # recent messages
  slackctl log --channel C0123456 --since 24h
  slackctl log search "deploy failed"
  slackctl log search "deploy* AND staging" --channel C0123456
  slackctl log stats
  slackctl log prune --older-than 2160h          # drop anything older than 90 days`,
			Args: cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				f, err := buildLogFilter(channel, user, since, limit)
				if err != nil {
					return err
				}
				return withReadStore(cmd, func(st *store.Store) error {
					msgs, err := st.Query(cmd.Context(), f)
					if err != nil {
						return err
					}
					return renderMessages(cmd, msgs)
				})
			},
		}
		bindLogFilterFlags(logCmd, &channel, &user, &since, &limit)
		markKind(logCmd, kindRead)

		search, stats, prune, path := logSearchCmd(), logStatsCmd(), logPruneCmd(), logPathCmd()
		markKind(search, kindRead)
		markKind(stats, kindRead)
		markKind(prune, kindDestructive)
		markKind(path, kindRead)
		logCmd.AddCommand(search, stats, prune, path)
		root.AddCommand(logCmd)
	})
}

func bindLogFilterFlags(cmd *cobra.Command, channel, user, since *string, limit *int) {
	cmd.Flags().StringVar(channel, "channel", "", "filter by conversation id")
	cmd.Flags().StringVar(user, "user", "", "filter by user id")
	cmd.Flags().StringVar(since, "since", "", "only messages at/after this time: a Slack ts, unix seconds, or a Go duration (24h)")
	cmd.Flags().IntVar(limit, "limit", store.DefaultLimit, "max rows to return")
}

func buildLogFilter(channel, user, since string, limit int) (store.Filter, error) {
	sinceTS, err := parseSinceTS(since)
	if err != nil {
		return store.Filter{}, err
	}
	return store.Filter{Channel: channel, User: user, Since: sinceTS, Limit: limit}, nil
}

// parseSinceTS turns a Go duration ("24h" → 24h ago) into a unix-seconds bound, and passes a
// raw Slack ts / unix value through unchanged. Slack ts sorts chronologically as text, so a
// unix-seconds lower bound compares correctly against stored ts values.
func parseSinceTS(s string) (string, error) {
	if s == "" {
		return "", nil
	}
	if d, err := time.ParseDuration(s); err == nil {
		return strconv.FormatInt(time.Now().Add(-d).Unix(), 10), nil
	}
	// A bare number (unix or Slack ts) passes through.
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return s, nil
	}
	return "", fmt.Errorf("invalid --since %q (want a Slack ts, unix seconds, or a Go duration like 24h)", s)
}

func logSearchCmd() *cobra.Command {
	var channel, user, since string
	var limit int
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Full-text search recorded message text",
		Long: `Search uses SQLite FTS5 when available (operators: AND/OR/NOT, prefix*, "phrases");
otherwise it degrades to a substring scan automatically. 'slackctl log stats' reports which
mode is active.`,
		Example: `  slackctl log search "deploy failed"
  slackctl log search "incident* AND prod" --channel C0123456 --since 7d`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f, err := buildLogFilter(channel, user, since, limit)
			if err != nil {
				return err
			}
			return withReadStore(cmd, func(st *store.Store) error {
				msgs, err := st.Search(cmd.Context(), args[0], f)
				if err != nil {
					return err
				}
				return renderMessages(cmd, msgs)
			})
		},
	}
	bindLogFilterFlags(cmd, &channel, &user, &since, &limit)
	return cmd
}

func logStatsCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "stats",
		Short:   "Show what the local history holds",
		Example: "  slackctl log stats\n  slackctl log stats -o json",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withReadStore(cmd, func(st *store.Store) error {
				s, err := st.Stats(cmd.Context())
				if err != nil {
					return err
				}
				return render(cmd, mustJSON(s))
			})
		},
	}
}

func logPruneCmd() *cobra.Command {
	var olderThan string
	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Delete recorded messages older than a duration",
		Example: `  slackctl log prune --older-than 2160h   # 90 days
  slackctl log prune --older-than 720h    # 30 days`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			d, err := time.ParseDuration(olderThan)
			if err != nil {
				return fmt.Errorf("invalid --older-than %q (want a Go duration like 2160h): %w", olderThan, err)
			}
			return withReadStore(cmd, func(st *store.Store) error {
				n, err := st.Prune(cmd.Context(), d)
				if err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "pruned %d message(s) recorded before now minus %s\n", n, olderThan)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&olderThan, "older-than", "", "delete messages recorded before now minus this Go duration (required)")
	_ = cmd.MarkFlagRequired("older-than")
	return cmd
}

func logPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the path to the local history database",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			profileName, _, err := resolveProfileName(cmd)
			if err != nil {
				return err
			}
			dir, err := config.Dir()
			if err != nil {
				return err
			}
			path, err := store.PathFor(dir, profileName)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	}
}

// withReadStore opens the active workspace's store for a read subcommand and closes it after.
// Unlike the write path, a failure here IS a real command error — reading local history is
// what `log` exists to do, so a silent "no messages" would be misleading.
func withReadStore(cmd *cobra.Command, fn func(*store.Store) error) error {
	profileName, _, err := resolveProfileName(cmd)
	if err != nil {
		return err
	}
	dir, err := config.Dir()
	if err != nil {
		return err
	}
	path, err := store.PathFor(dir, profileName)
	if err != nil {
		return err
	}
	st, err := store.Open(path)
	if err != nil {
		return fmt.Errorf("open local history: %w", err)
	}
	defer func() { _ = st.Close() }()
	return fn(st)
}

// renderMessages applies the default columns (unless --columns was set) then renders.
func renderMessages(cmd *cobra.Command, msgs []store.Message) error {
	if !cmd.Flags().Changed("columns") {
		if err := cmd.Flags().Set("columns", strings.Join(logDefaultColumns, ",")); err != nil {
			return err
		}
	}
	return render(cmd, mustJSON(msgs))
}
