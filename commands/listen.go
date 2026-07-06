package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/slackctl/internal/auth"
	"github.com/jjuanrivvera/slackctl/internal/config"
	"github.com/jjuanrivvera/slackctl/internal/rtm"
	"github.com/jjuanrivvera/slackctl/internal/slackevent"
	"github.com/jjuanrivvera/slackctl/internal/socketmode"
)

// listen is the beyond-the-REST-mold command: a live event stream for pipe/filter use. It
// has two transports so it works with WHATEVER credential you have (DECISIONS.md):
//   - Socket Mode  — needs an app-level token (xapp-); acks enveloped events.
//   - RTM          — the legacy WebSocket; works with a user/session token (xoxc+xoxd),
//     the credential a slack-mcp-style setup already has.
//
// --transport auto picks Socket Mode when an app token is present, else RTM.
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
		transport       string
		debugReconnects bool
	)
	cmd := &cobra.Command{
		Use:   "listen",
		Short: "Stream events live (Socket Mode or RTM) as lines",
		Long: `Open a live connection and stream events as they happen, one line each —
human-readable by default, JSON with --json (for pipes), full frames with --raw.

Two transports, auto-selected by the credential you have (--transport to force):
  socket  Socket Mode — needs an app-level token (xapp-, connections:write). Official,
          robust, acks enveloped events. Store one with 'slackctl auth login --kind app'.
  rtm     Real Time Messaging — the legacy WebSocket that works with a user/session token
          (xoxp- or the xoxc+xoxd browser pair). No Slack app required: this is the path
          that streams with the same credentials a slack-mcp setup already uses. RTM is
          legacy and unofficial for xoxc tokens — a workspace may block it.

Filters combine as a union: --dms OR --channels; --events then narrows by event type.
With no filters, every received event streams. (Socket Mode only delivers events your app
subscribed to and is a member of; RTM delivers everything the user account can see.)
Runs until Ctrl-C.`,
		Example: `  slackctl listen --dms --json                       # auto transport, DM events as NDJSON
  slackctl listen --transport rtm --json             # force RTM (session/user token)
  slackctl listen --channels C0123456,C0456789 --json
  slackctl listen --events message,reaction_added
  slackctl listen --raw | jq .`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			mode, err := resolveTransport(cmd, transport)
			if err != nil {
				return err
			}

			channelSet := toSet(channels)
			eventSet := toSet(events)
			out := cmd.OutOrStdout()

			// emit applies the shared filter + render to a bare event object (the same shape
			// from both transports), so RTM and Socket Mode behave identically downstream.
			emit := func(event json.RawMessage, meta slackevent.Meta) {
				if len(eventSet) > 0 && !eventSet[meta.Type] {
					return
				}
				if dms || len(channelSet) > 0 {
					dmMatch := dms && meta.IsDM()
					if !dmMatch && !channelSet[meta.ChannelOf()] {
						return
					}
				}
				if jsonOut {
					fmt.Fprintln(out, string(event))
					return
				}
				fmt.Fprintln(out, humanEventLine(event, meta))
			}

			switch mode {
			case "socket":
				return runSocketMode(cmd, out, rawOut, debugReconnects, emit)
			default: // "rtm"
				return runRTM(cmd, out, rawOut, emit)
			}
		},
	}
	cmd.Flags().BoolVar(&dms, "dms", false, "only events from direct messages")
	cmd.Flags().StringSliceVar(&channels, "channels", nil, "only events from these conversation ids")
	cmd.Flags().StringSliceVar(&events, "events", nil, "only these event types (message,reaction_added,…)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit each event as one JSON line (NDJSON)")
	cmd.Flags().BoolVar(&rawOut, "raw", false, "emit full wire frames as NDJSON (Socket Mode envelopes / RTM frames)")
	cmd.Flags().StringVar(&transport, "transport", "auto", "event transport: auto|socket|rtm")
	cmd.Flags().BoolVar(&debugReconnects, "debug-reconnects", false, "Socket Mode only: rotate the connection every ~360s (tests reconnect handling)")
	markKind(cmd, kindRead)
	return cmd
}

// resolveTransport turns the --transport choice into a concrete "socket" or "rtm". auto
// picks Socket Mode when an app token is available, else RTM.
func resolveTransport(cmd *cobra.Command, choice string) (string, error) {
	switch choice {
	case "socket", "rtm":
		return choice, nil
	case "auto", "":
		if hasAppToken(cmd) {
			return "socket", nil
		}
		return "rtm", nil
	default:
		return "", fmt.Errorf("unknown --transport %q (want auto|socket|rtm)", choice)
	}
}

// hasAppToken reports whether an app-level token is resolvable, without erroring — the
// signal for auto transport selection. It checks the app env vars and the profile's app
// keyring entry, but NOT a session/bot credential (those don't open Socket Mode).
func hasAppToken(cmd *cobra.Command) bool {
	if strings.HasPrefix(os.Getenv("SLACKCTL_TOKEN"), "xapp-") || os.Getenv("SLACK_APP_TOKEN") != "" {
		return true
	}
	profileName, _, err := resolveProfileName(cmd)
	if err != nil {
		return false
	}
	dir, err := config.Dir()
	if err != nil {
		return false
	}
	_, err = auth.New(dir).Get(auth.Key(profileName, auth.KindApp))
	return err == nil
}

// runSocketMode streams via Socket Mode (app token). It acks every envelope before emitting.
func runSocketMode(cmd *cobra.Command, out io.Writer, rawOut, debugReconnects bool, emit func(json.RawMessage, slackevent.Meta)) error {
	client, err := clientForKind(cmd, auth.KindApp)
	if err != nil {
		return err
	}
	if client.DryRun {
		return fmt.Errorf("listen opens a live websocket; --dry-run has nothing to print")
	}
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
		emit(event, meta)
	})
}

// runRTM streams via the legacy RTM WebSocket (user/session token). RTM frames ARE the event
// objects, so no envelope unwrapping is needed.
func runRTM(cmd *cobra.Command, out io.Writer, rawOut bool, emit func(json.RawMessage, slackevent.Meta)) error {
	// RTM needs a user-grade credential; the bot token can't call rtm.connect. KindUser
	// resolves a user token or falls back to the browser-session pair.
	client, err := clientForKind(cmd, auth.KindUser)
	if err != nil {
		return err
	}
	if client.DryRun {
		return fmt.Errorf("listen opens a live websocket; --dry-run has nothing to print")
	}
	rc := rtm.New(client.OpenRTMURL, cmd.ErrOrStderr())
	// Session credentials must send the d cookie on the WebSocket handshake too, or the RTM
	// gateway answers invalid_auth even though rtm.connect succeeded.
	rc.Header = client.DialHeaders()
	return rc.Run(cmd.Context(), func(frame json.RawMessage) {
		if rawOut {
			fmt.Fprintln(out, string(frame))
			return
		}
		meta, err := slackevent.ParseMeta(frame)
		if err != nil {
			return
		}
		emit(frame, meta)
	})
}

func toSet(items []string) map[string]bool {
	if len(items) == 0 {
		return nil
	}
	m := make(map[string]bool, len(items))
	for _, it := range items {
		m[it] = true
	}
	return m
}

// humanEventLine renders one event compactly for interactive use: ts, type, where, who,
// and the text when there is one.
func humanEventLine(event json.RawMessage, meta slackevent.Meta) string {
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
