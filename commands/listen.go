package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/slackctl/internal/auth"
	"github.com/jjuanrivvera/slackctl/internal/socketmode"
)

// listen is the beyond-the-REST-mold command: a Socket Mode stream for pipe/filter use.
// It needs an app-level token (xapp-) alongside the bot token, and the app must have
// Socket Mode enabled with the wanted events subscribed (message.im, message.channels,
// reaction_added, …) — Slack only delivers subscribed events for conversations the bot
// is in.
func init() {
	register(func(root *cobra.Command) {
		root.AddCommand(listenCmd())
	})
}

func listenCmd() *cobra.Command {
	var (
		dms             bool
		channels        []string
		events          []string
		jsonOut         bool
		rawOut          bool
		debugReconnects bool
	)
	cmd := &cobra.Command{
		Use:   "listen",
		Short: "Stream events over Socket Mode (DMs, channels) as lines",
		Long: `Open a Socket Mode connection and stream events as they happen, one line each —
human-readable by default, JSON with --json (for pipes), full envelopes with --raw.

Filters combine as a union: --dms OR --channels; --events then narrows by event type.
With no filters, every subscribed event streams. Every envelope is acknowledged to Slack
immediately (before filtering), so filtered-out events are consumed, not redelivered.

Requires an app-level token (xapp-) with connections:write — store it with
'slackctl auth login --kind app' — plus Socket Mode enabled and the event subscriptions
(message.im, message.channels, reaction_added, …) configured for the app. Slack only
sends message events for conversations the bot is a member of. Runs until Ctrl-C.`,
		Example: `  slackctl listen --dms --json                       # stream DM events as NDJSON
  slackctl listen --dms --channels C0123456,C0456789 --json
  slackctl listen --events message,reaction_added
  slackctl listen --raw | jq .payload.event.type`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientForKind(cmd, auth.KindApp)
			if err != nil {
				return err
			}
			if client.DryRun {
				return fmt.Errorf("listen opens a live websocket; --dry-run has nothing to print")
			}

			channelSet := map[string]bool{}
			for _, c := range channels {
				channelSet[c] = true
			}
			eventSet := map[string]bool{}
			for _, e := range events {
				eventSet[e] = true
			}

			out := cmd.OutOrStdout()
			sm := socketmode.New(client.OpenSocketURL, cmd.ErrOrStderr())
			sm.DebugReconnects = debugReconnects

			return sm.Run(cmd.Context(), func(env socketmode.Envelope) {
				if rawOut {
					b, _ := json.Marshal(env)
					fmt.Fprintln(out, string(b))
					return
				}
				event, meta, err := socketmode.ParseEvent(env)
				if err != nil {
					return // slash_commands/interactive frames stream only under --raw
				}
				if len(eventSet) > 0 && !eventSet[meta.Type] {
					return
				}
				// --dms / --channels union: match either; no filter = everything.
				if dms || len(channelSet) > 0 {
					if !(dms && meta.IsDM()) && !channelSet[meta.ChannelOf()] {
						return
					}
				}
				if jsonOut {
					fmt.Fprintln(out, string(event))
					return
				}
				fmt.Fprintln(out, humanEventLine(event, meta))
			})
		},
	}
	cmd.Flags().BoolVar(&dms, "dms", false, "only events from direct messages")
	cmd.Flags().StringSliceVar(&channels, "channels", nil, "only events from these conversation ids")
	cmd.Flags().StringSliceVar(&events, "events", nil, "only these event types (message,reaction_added,…)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit each event as one JSON line (NDJSON)")
	cmd.Flags().BoolVar(&rawOut, "raw", false, "emit full Socket Mode envelopes as NDJSON (includes slash commands and interactivity)")
	cmd.Flags().BoolVar(&debugReconnects, "debug-reconnects", false, "ask Slack to rotate the connection every ~360s (tests reconnect handling)")
	markKind(cmd, kindRead)
	return cmd
}

// humanEventLine renders one event compactly for interactive use: ts, type, where, who,
// and the text when there is one.
func humanEventLine(event json.RawMessage, meta socketmode.EventMeta) string {
	var e struct {
		Text     string `json:"text"`
		Reaction string `json:"reaction"`
		Subtype  string `json:"subtype"`
	}
	_ = json.Unmarshal(event, &e)
	parts := []string{meta.TS, meta.Type}
	if e.Subtype != "" {
		parts[1] += "/" + e.Subtype
	}
	if ch := meta.ChannelOf(); ch != "" {
		parts = append(parts, ch)
	}
	if meta.User != "" {
		parts = append(parts, meta.User)
	}
	if e.Reaction != "" {
		parts = append(parts, ":"+e.Reaction+":")
	}
	if e.Text != "" {
		parts = append(parts, "— "+e.Text)
	}
	return strings.Join(parts, "  ")
}
