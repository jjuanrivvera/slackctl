package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/slackctl/internal/api"
	"github.com/jjuanrivvera/slackctl/internal/auth"
)

// The Slack Web API is RPC-method-oriented (conversations.list, chat.postMessage, ...), not
// CRUD on resources, so slackctl uses a generic *method-command* builder instead of a generic
// CRUD resource (DECISIONS.md). A group is a noun (conversations, users, ...) whose verbs each
// map 1:1 to a Web API method. Adding a method is a few declarative lines in a group file —
// zero edits to this shared builder.

// cmdKind classifies a command for retry safety and MCP/agent-guard annotations.
type cmdKind int

const (
	kindRead        cmdKind = iota // read-only: idempotent, safe to auto-retry, MCP readOnlyHint
	kindWrite                      // creates/changes state: MCP openWorldHint
	kindDestructive                // irreversible (delete/archive/kick): MCP destructiveHint
)

// MCP tool annotation keys (the singular MCP hint keys; ophis reads these from cmd.Annotations).
const (
	annReadOnly    = "readOnlyHint"
	annDestructive = "destructiveHint"
	annOpenWorld   = "openWorldHint"
	annIdempotent  = "idempotentHint"
)

type flagKind int

const (
	flagString flagKind = iota
	flagInt
	flagFloat
	flagBool
	flagStringSlice
	flagJSON // value parsed as JSON and sent as-is (objects/arrays, e.g. blocks)
)

// flagSpec declares one CLI flag and the Web API parameter it maps to.
type flagSpec struct {
	Name     string
	Param    string // defaults to Name with '-' → '_'
	Kind     flagKind
	Required bool
	Short    string
	Default  string
	Usage    string
}

func (f flagSpec) param() string {
	if f.Param != "" {
		return f.Param
	}
	return strings.ReplaceAll(f.Name, "-", "_")
}

// methodCmd declares one verb (a Web API method).
type methodCmd struct {
	Use     string
	Aliases []string
	Method  string
	Short   string
	Long    string
	Example string
	Kind    cmdKind
	Flags   []flagSpec
	Columns []string
	// ResultKey names the field of Slack's response envelope that carries the payload
	// (channels, messages, members, ...). Slack nests payloads as siblings of "ok", so
	// rendering the whole envelope would put `ok: true` in every table. Empty = render the
	// full body minus nothing.
	ResultKey string
	// Paginated marks a cursor-paginated list method: the builder adds --limit/--all flags
	// and walks response_metadata.next_cursor through the client's CallAllPages.
	ResultIsArray bool // ResultKey holds an array (list methods) rather than an object
	Paginated     bool
	// Token selects which stored token kind the method needs. Most methods take the bot
	// token; a few are user-token-only (search.messages, stars.*) — Slack rejects bot
	// tokens there with not_allowed_token_type.
	Token tokenNeed
}

// tokenNeed declares a method's token-kind requirement.
type tokenNeed int

const (
	tokenDefault      tokenNeed = iota // bot token (or user token via --as-user)
	tokenUserRequired                  // user token only; fail fast with a hint if missing
)

// group is a noun with its verbs.
type group struct {
	Use     string
	Aliases []string
	Short   string
	Long    string
	Cmds    []methodCmd
	// Extra holds hand-written subcommands for value-adds that aren't a single API method
	// (e.g. `conversations unreads`). They are FACTORIES, not shared instances, so each
	// NewRootCmd builds a fresh command tree and cobra flag state never leaks across tests.
	Extra []func() *cobra.Command
}

// apiCmdInfo records a built API command for the MCP server and agent guard to classify.
type apiCmdInfo struct {
	Path   string // canonical CLI path, e.g. "msg post"
	Method string
	Kind   cmdKind
	// aliasPaths are the alternate CLI paths reachable through cobra aliases —
	// every group-alias × verb-alias combination except the canonical path
	// (e.g. "chat post", "msg send", "chat send"). The agent guard must block these
	// too, or an alias invocation silently bypasses the deny rules that only name the
	// canonical path.
	aliasPaths []string
}

var registeredAPICmds []apiCmdInfo

// APICommands returns the classification of every API-backed command (for agent guard tests).
func APICommands() []apiCmdInfo { return registeredAPICmds }

// PathString returns the command path; IsRead/IsDestructive expose the classification.
func (a apiCmdInfo) PathString() string  { return a.Path }
func (a apiCmdInfo) IsRead() bool        { return a.Kind == kindRead }
func (a apiCmdInfo) IsDestructive() bool { return a.Kind == kindDestructive }

// AllPaths returns every CLI path that invokes this command: the canonical path
// plus all cobra alias combinations.
func (a apiCmdInfo) AllPaths() []string {
	return append([]string{a.Path}, a.aliasPaths...)
}

// registerGroup adds a group's commands to the root tree and the classification registry.
func registerGroup(g group) {
	groupNames := append([]string{g.Use}, g.Aliases...)
	for _, mc := range g.Cmds {
		verbNames := append([]string{mc.Use}, mc.Aliases...)
		var aliasPaths []string
		for _, gn := range groupNames {
			for _, vn := range verbNames {
				if gn == g.Use && vn == mc.Use {
					continue // the canonical path lives in Path
				}
				aliasPaths = append(aliasPaths, gn+" "+vn)
			}
		}
		registeredAPICmds = append(registeredAPICmds, apiCmdInfo{
			Path:       g.Use + " " + mc.Use,
			Method:     mc.Method,
			Kind:       mc.Kind,
			aliasPaths: aliasPaths,
		})
	}
	register(func(root *cobra.Command) {
		parent := &cobra.Command{
			Use:     g.Use,
			Aliases: g.Aliases,
			Short:   g.Short,
			Long:    g.Long,
		}
		for _, mc := range g.Cmds {
			parent.AddCommand(buildMethodCmd(mc))
		}
		for _, factory := range g.Extra {
			parent.AddCommand(factory())
		}
		root.AddCommand(parent)
	})
}

// buildMethodCmd turns a methodCmd into a cobra command: it binds the declared flags, stamps
// MCP annotations from the Kind, and on run assembles the params, calls the API (walking
// cursors for paginated methods), and renders the result.
func buildMethodCmd(mc methodCmd) *cobra.Command {
	cmd := &cobra.Command{
		Use:     mc.Use,
		Aliases: mc.Aliases,
		Short:   mc.Short,
		Long:    mc.Long,
		Example: mc.Example,
		Args:    cobra.NoArgs,
	}
	markKind(cmd, mc.Kind)
	bindFlags(cmd, mc)
	if mc.Paginated {
		cmd.Flags().Int("limit", 100, "max items to fetch across pages")
		cmd.Flags().Bool("all", false, "fetch every page (overrides --limit)")
	}

	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		params, err := collectParams(cmd, mc)
		if err != nil {
			return err
		}
		client, err := clientForNeed(cmd, mc.Token)
		if err != nil {
			return err
		}
		idempotent := mc.Kind == kindRead
		var raw json.RawMessage
		if mc.Paginated {
			max, _ := cmd.Flags().GetInt("limit")
			if all, _ := cmd.Flags().GetBool("all"); all {
				max = 0
			}
			// Slack recommends 100-200 items per request; the walker caps the total.
			params["limit"] = pageSize(max)
			raw, err = client.CallAllPages(cmd.Context(), mc.Method, params, mc.ResultKey, max)
		} else {
			raw, err = client.Call(cmd.Context(), mc.Method, params, idempotent)
			if err == nil && raw != nil && mc.ResultKey != "" {
				raw, err = extractField(raw, mc.ResultKey)
			}
		}
		if err != nil {
			return err
		}
		if len(mc.Columns) > 0 && !cmd.Flags().Changed("columns") {
			// Apply the command's default columns unless the user overrode --columns.
			if err := cmd.Flags().Set("columns", strings.Join(mc.Columns, ",")); err != nil {
				return err
			}
		}
		if raw == nil { // dry-run made no request
			return nil
		}
		return render(cmd, raw)
	}
	return cmd
}

// clientForNeed maps a methodCmd's token requirement to a client. tokenUserRequired ignores
// --as-user (it is already user-only).
func clientForNeed(cmd *cobra.Command, need tokenNeed) (*api.Client, error) {
	if need == tokenUserRequired {
		return clientForKind(cmd, auth.KindUser)
	}
	return clientFromCmd(cmd)
}

// pageSize picks the per-request limit: Slack recommends 100-200; never ask for more than
// the total cap.
func pageSize(max int) int {
	const recommended = 200
	if max > 0 && max < recommended {
		return max
	}
	return recommended
}

// extractField pulls one named field out of a response body, so tables show the payload
// (the channel, the message list) instead of Slack's envelope. A dotted key walks nested
// objects ("messages.matches" for search results).
func extractField(raw json.RawMessage, key string) (json.RawMessage, error) {
	current := raw
	for _, part := range strings.Split(key, ".") {
		var body map[string]json.RawMessage
		if err := json.Unmarshal(current, &body); err != nil {
			return nil, err
		}
		field, ok := body[part]
		if !ok {
			// Some methods omit the field on success (e.g. an empty result); fall back to
			// the full body rather than failing a successful call.
			return raw, nil
		}
		current = field
	}
	return current, nil
}

// markKind stamps MCP annotations so the mcp server and agent guard can gate writes. A write
// sets only openWorldHint; a destructive verb adds destructiveHint; a read sets readOnlyHint.
func markKind(cmd *cobra.Command, kind cmdKind) {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	switch kind {
	case kindRead:
		cmd.Annotations[annReadOnly] = "true"
		cmd.Annotations[annIdempotent] = "true"
	case kindWrite:
		cmd.Annotations[annOpenWorld] = "true"
	case kindDestructive:
		cmd.Annotations[annOpenWorld] = "true"
		cmd.Annotations[annDestructive] = "true"
	}
}

func bindFlags(cmd *cobra.Command, mc methodCmd) {
	f := cmd.Flags()
	for _, fs := range mc.Flags {
		switch fs.Kind {
		case flagString, flagJSON:
			f.StringP(fs.Name, fs.Short, fs.Default, fs.Usage)
		case flagInt:
			f.Int64P(fs.Name, fs.Short, 0, fs.Usage)
		case flagFloat:
			f.Float64P(fs.Name, fs.Short, 0, fs.Usage)
		case flagBool:
			f.BoolP(fs.Name, fs.Short, false, fs.Usage)
		case flagStringSlice:
			f.StringSliceP(fs.Name, fs.Short, nil, fs.Usage)
		}
		if fs.Required {
			_ = cmd.MarkFlagRequired(fs.Name)
		}
	}
}

// collectParams reads the set flags into the Web API params map. Only flags the user actually
// set are sent, so API defaults apply otherwise. Defaulted string flags (Default != "") are
// sent too — they document the CLI's chosen default in the request itself.
func collectParams(cmd *cobra.Command, mc methodCmd) (map[string]any, error) {
	params := map[string]any{}
	f := cmd.Flags()
	for _, fs := range mc.Flags {
		if !f.Changed(fs.Name) && fs.Default == "" {
			continue
		}
		switch fs.Kind {
		case flagString:
			v, _ := f.GetString(fs.Name)
			if v != "" {
				params[fs.param()] = v
			}
		case flagInt:
			v, _ := f.GetInt64(fs.Name)
			params[fs.param()] = v
		case flagFloat:
			v, _ := f.GetFloat64(fs.Name)
			params[fs.param()] = v
		case flagBool:
			v, _ := f.GetBool(fs.Name)
			params[fs.param()] = v
		case flagStringSlice:
			v, _ := f.GetStringSlice(fs.Name)
			params[fs.param()] = strings.Join(v, ",") // Slack takes comma-separated lists
		case flagJSON:
			v, _ := f.GetString(fs.Name)
			if v == "" {
				continue
			}
			var parsed any
			if err := json.Unmarshal([]byte(v), &parsed); err != nil {
				return nil, fmt.Errorf("--%s must be valid JSON: %w", fs.Name, err)
			}
			params[fs.param()] = parsed
		}
	}
	return params, nil
}
