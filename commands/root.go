// Package commands wires the Slack Web API client (internal/api) into a Cobra command
// tree. Command groups self-register via register() so a fresh tree can be built per process
// (and per test). Shared concerns — client construction and rendering — live here once.
package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/slackctl/internal/api"
	"github.com/jjuanrivvera/slackctl/internal/auth"
	"github.com/jjuanrivvera/slackctl/internal/config"
	"github.com/jjuanrivvera/slackctl/internal/output"
	"github.com/jjuanrivvera/slackctl/internal/store"
)

// registrations are applied to each fresh root command. Command files append to it in init().
var registrations []func(*cobra.Command)

func register(fn func(*cobra.Command)) { registrations = append(registrations, fn) }

const rootLong = `slackctl is a fast, scriptable command-line tool for the Slack Web API.

It wraps the Web API methods (conversations.list, chat.postMessage, search.messages, ...)
behind ergonomic commands with table/json/yaml/csv output, named profiles for multiple
workspaces, OS-keyring token storage, a Socket Mode listener, and an MCP server so AI
agents can drive it safely.

Create an app at https://api.slack.com/apps, install it to your workspace, then:

  slackctl auth login                              # store the bot token in your OS keyring
  slackctl auth status                             # who am I?
  slackctl conversations list                      # channels the token can see
  slackctl msg post --channel C0123456 --text hi   # post a message
  slackctl listen --dms --json                     # stream events (Socket Mode, needs an xapp token)

Every command honors --dry-run (prints the equivalent curl), -o/--output, and --jq.`

// NewRootCmd builds a fresh command tree with all registered groups attached.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "slackctl",
		Short:         "Command-line tool for the Slack Web API",
		Long:          rootLong,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       "", // set by cmd/slackctl via SetVersionTemplate; version cmd has the detail
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			f, _ := cmd.Flags().GetString("output")
			if !output.Format(f).Valid() {
				return fmt.Errorf("invalid --output %q (want table|json|yaml|csv|id)", f)
			}
			return nil
		},
	}
	addGlobalFlags(root)
	for _, fn := range registrations {
		fn(root)
	}
	groupCommands(root)
	return root
}

// commandGroups assigns each top-level command to a gh-style section so `--help` reads like a
// first-party tool instead of one flat alphabetical list.
var commandGroups = map[string]string{
	"conversations": "conversations", "msg": "conversations", "search": "conversations",
	"assistant": "conversations", "listen": "conversations",
	"files": "conversations", "canvases": "conversations",
	"users": "directory", "usergroups": "directory", "team": "directory", "dnd": "directory",
	"reactions": "messages", "pins": "messages", "bookmarks": "messages",
	"saved": "messages", "emoji": "messages",
	"auth": "meta", "config": "meta", "init": "meta", "doctor": "meta",
	"alias": "meta", "api": "meta", "version": "meta", "completion": "meta",
	"mcp": "agents", "agent": "agents",
}

func groupCommands(root *cobra.Command) {
	root.AddGroup(
		&cobra.Group{ID: "conversations", Title: "Conversations, messaging & files:"},
		&cobra.Group{ID: "messages", Title: "Reactions, pins, bookmarks & saved:"},
		&cobra.Group{ID: "directory", Title: "People & workspace:"},
		&cobra.Group{ID: "agents", Title: "AI agents:"},
		&cobra.Group{ID: "meta", Title: "Setup & meta:"},
	)
	for _, c := range root.Commands() {
		if id, ok := commandGroups[c.Name()]; ok {
			c.GroupID = id
		}
	}
}

func addGlobalFlags(root *cobra.Command) {
	pf := root.PersistentFlags()
	pf.StringP("output", "o", "table", "output format: table|json|yaml|csv|id")
	// A slackctl "profile" is one workspace, so the user-facing flag is --workspace.
	// --profile stays as a hidden, still-working alias so generic scripts don't break.
	pf.String("workspace", "", "workspace to use: a named profile/credential (env SLACKCTL_WORKSPACE)")
	pf.String("profile", "", "alias for --workspace")
	_ = pf.MarkHidden("profile")
	pf.String("base-url", "", "Web API base URL (default https://slack.com/api)")
	pf.Bool("as-user", false, "use the stored user token (xoxp-) instead of the bot token")
	pf.Bool("dry-run", false, "print the equivalent curl and make no request")
	pf.Bool("show-token", false, "do not redact the token in --dry-run output")
	pf.BoolP("verbose", "v", false, "log raw API responses to stderr")
	pf.Bool("no-color", false, "disable colored output")
	pf.StringSlice("columns", nil, "explicit, ordered table/csv columns")
	// --quiet has no -q short: the `api` escape hatch uses -q for repeatable key=value params.
	pf.Bool("quiet", false, "suppress notes on stderr")
	pf.String("jq", "", "gojq expression applied to the result before rendering")
	pf.Float64("rps", 0, "client-side requests-per-second cap (0 = default)")
	pf.Bool("no-store", false, "do not record messages to the local history store (see `slackctl log`)")
}

// openWorkspaceStore opens the workspace's local history DB for the write (recorder) path.
// Failure is never fatal here: nil means "recording is off for this call"; the caller keeps
// going. The read path (`slackctl log`) treats an open failure as a real error instead.
func openWorkspaceStore(cmd *cobra.Command, workspace string) *store.Store {
	dir, err := config.Dir()
	if err != nil {
		warnStoreUnavailable(cmd, err)
		return nil
	}
	path, err := store.PathFor(dir, workspace)
	if err != nil {
		warnStoreUnavailable(cmd, err)
		return nil
	}
	st, err := store.Open(path)
	if err != nil {
		warnStoreUnavailable(cmd, err)
		return nil
	}
	return st
}

func warnStoreUnavailable(cmd *cobra.Command, err error) {
	if quiet, _ := cmd.Flags().GetBool("quiet"); quiet {
		return
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "slackctl: warning: local history unavailable (%v) — continuing without it\n", err)
}

// clientFromCmd builds an API client with the default token for the command: the bot token,
// or the user token when --as-user is set.
func clientFromCmd(cmd *cobra.Command) (*api.Client, error) {
	kind := auth.KindBot
	if asUser, _ := cmd.Flags().GetBool("as-user"); asUser {
		kind = auth.KindUser
	}
	return clientForKind(cmd, kind)
}

// clientForKind builds an API client authenticated with the given token kind. Token
// precedence per kind: $SLACKCTL_TOKEN (explicit override, any kind) > the kind's
// conventional env var (SLACK_BOT_TOKEN / SLACK_USER_TOKEN / SLACK_APP_TOKEN /
// SLACK_XOXC_TOKEN+SLACK_XOXD_TOKEN) > the profile's keyring entry, then — for bot/user
// kinds — a browser-session (xoxc+xoxd) fallback so a web-client credential with no OAuth
// token still works.
func clientForKind(cmd *cobra.Command, kind auth.TokenKind) (*api.Client, error) {
	f := cmd.Flags()
	baseURLFlag, _ := f.GetString("base-url")
	dryRun, _ := f.GetBool("dry-run")
	showToken, _ := f.GetBool("show-token")
	verbose, _ := f.GetBool("verbose")
	rps, _ := f.GetFloat64("rps")

	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	profileName := cfg.ResolveProfileName(resolveWorkspaceFlag(cmd))
	prof, _ := cfg.Profile(profileName)

	authr, err := resolveAuth(profileName, kind)
	if err != nil {
		return nil, err
	}

	baseURL := config.FirstNonEmpty(baseURLFlag, prof.BaseURL, api.DefaultBaseURL)
	if err := config.ValidateBaseURL(baseURL); err != nil {
		return nil, err
	}

	opts := []api.Option{
		api.WithBaseURL(baseURL),
		api.WithDryRun(dryRun),
		api.WithShowToken(showToken),
		api.WithVerbose(verbose),
		api.WithDryRunWriter(cmd.ErrOrStderr()),
	}
	if rps > 0 {
		opts = append(opts, api.WithRPS(rps))
	}
	// Local history recorder: best-effort and additive. A disabled/unavailable store never
	// blocks building a client (a post still works without local history). Dry-run makes no
	// request, so there is nothing to record — skip opening a DB file for it.
	if !dryRun {
		if rec := openRecorder(cmd, profileName); rec != nil {
			opts = append(opts, api.WithRecorder(rec))
		}
	}
	return api.New(authr, opts...), nil
}

// openRecorder opens the active workspace's local history store as a Recorder, honoring
// --no-store. It returns nil (recording disabled) on any failure, warning once on stderr —
// a broken store must never prevent building a client.
func openRecorder(cmd *cobra.Command, workspace string) *storeRecorder {
	if noStore, _ := cmd.Flags().GetBool("no-store"); noStore {
		return nil
	}
	st := openWorkspaceStore(cmd, workspace)
	if st == nil {
		return nil
	}
	quiet, _ := cmd.Flags().GetBool("quiet")
	return &storeRecorder{st: st, workspace: workspace, quiet: quiet, errW: cmd.ErrOrStderr()}
}

// allowsSessionFallback reports whether a browser-session (xoxc+xoxd) credential may stand
// in for the requested kind. Session creds carry the user's own identity, so they satisfy
// bot- and user-kind commands — but NOT the app kind, whose Socket Mode connection
// (apps.connections.open) genuinely needs an app-level xapp token.
func allowsSessionFallback(kind auth.TokenKind) bool {
	return kind == auth.KindBot || kind == auth.KindUser
}

// resolveAuth builds the Authenticator for a profile+kind, honoring env overrides ahead of
// the keyring, and falling back to browser-session (xoxc+xoxd) creds for bot/user kinds
// when no OAuth token is stored.
func resolveAuth(profileName string, kind auth.TokenKind) (api.Authenticator, error) {
	// 1. Explicit single-token env override (any single-token kind).
	if kind != auth.KindSession {
		if t := config.FirstNonEmpty(os.Getenv("SLACKCTL_TOKEN"), os.Getenv(kind.EnvVar())); t != "" {
			return api.NewTokenAuth(t)
		}
	}
	// 2. Session env pair — for an explicit session kind, or as a fallback for bot/user.
	if kind == auth.KindSession || allowsSessionFallback(kind) {
		if xoxc, xoxd := os.Getenv("SLACK_XOXC_TOKEN"), os.Getenv("SLACK_XOXD_TOKEN"); xoxc != "" && xoxd != "" {
			return api.NewSessionAuth(xoxc, xoxd)
		}
	}

	dir, err := config.Dir()
	if err != nil {
		return nil, err
	}
	store := auth.New(dir)

	// 3. Explicit session kind: only the stored pair.
	if kind == auth.KindSession {
		creds, err := auth.GetSession(store, profileName)
		if err != nil {
			return nil, fmt.Errorf("no session credentials for workspace %q — run `slackctl auth login --kind session` (or set $SLACK_XOXC_TOKEN and $SLACK_XOXD_TOKEN)", profileName)
		}
		return api.NewSessionAuth(creds.Token, creds.Cookie)
	}

	// 4. The kind's own OAuth token from the keyring.
	if token, err := store.Get(auth.Key(profileName, kind)); err == nil {
		return api.NewTokenAuth(token)
	}

	// 5. Browser-session fallback for bot/user kinds.
	if allowsSessionFallback(kind) {
		if creds, err := auth.GetSession(store, profileName); err == nil {
			return api.NewSessionAuth(creds.Token, creds.Cookie)
		}
	}

	return nil, fmt.Errorf("no %s token for workspace %q — run `slackctl auth login%s` (or set $%s, or store browser-session creds with `slackctl auth login --kind session`)",
		kind, profileName, loginKindFlag(kind), kind.EnvVar())
}

// loginKindFlag renders the `auth login` flag suffix for a kind, for error hints.
func loginKindFlag(kind auth.TokenKind) string {
	if kind == auth.KindBot {
		return ""
	}
	return " --kind " + string(kind)
}

// render writes data using the format/columns/jq/quiet flags, to the command's streams.
func render(cmd *cobra.Command, data json.RawMessage) error {
	f := cmd.Flags()
	format, _ := f.GetString("output")
	columns, _ := f.GetStringSlice("columns")
	noColor, _ := f.GetBool("no-color")
	quiet, _ := f.GetBool("quiet")
	jq, _ := f.GetString("jq")
	return output.Render(data, output.Options{
		Format:  output.Format(format),
		Columns: columns,
		NoColor: noColor,
		Quiet:   quiet,
		JQ:      jq,
		Out:     cmd.OutOrStdout(),
		Err:     cmd.ErrOrStderr(),
	})
}

// resolveProfileName returns the active profile name from flags/env/config without building
// a client — used by auth/config/doctor commands.
func resolveProfileName(cmd *cobra.Command) (string, *config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", nil, err
	}
	return cfg.ResolveProfileName(resolveWorkspaceFlag(cmd)), cfg, nil
}

// resolveWorkspaceFlag returns the selected workspace from the --workspace flag, falling
// back to the hidden --profile alias, so generic scripts that pass --profile keep working.
func resolveWorkspaceFlag(cmd *cobra.Command) string {
	if ws, _ := cmd.Flags().GetString("workspace"); ws != "" {
		return ws
	}
	prof, _ := cmd.Flags().GetString("profile")
	return prof
}
