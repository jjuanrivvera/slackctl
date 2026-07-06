package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/njayp/ophis"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// findCmd walks the tree to the command at the given path (e.g. "auth","login").
func findCmd(root *cobra.Command, path ...string) *cobra.Command {
	cur := root
	for _, name := range path {
		var next *cobra.Command
		for _, c := range cur.Commands() {
			if c.Name() == name {
				next = c
				break
			}
		}
		if next == nil {
			return nil
		}
		cur = next
	}
	return cur
}

func TestMCPExcludesSetupCommands(t *testing.T) {
	sel := ophis.ExcludeCmdsContaining(excludedFromMCP...)
	root := NewRootCmd()

	for _, p := range [][]string{
		{"auth", "login"}, {"config", "view"}, {"alias", "set"}, {"api"},
		{"version"}, {"doctor"}, {"listen"}, {"agent", "guard"}, {"init"},
	} {
		cmd := findCmd(root, p...)
		require.NotNil(t, cmd, "command %v should exist", p)
		assert.False(t, sel(cmd), "setup/secret command %v must be excluded from the MCP surface", p)
	}
	for _, p := range [][]string{{"msg", "post"}, {"conversations", "list"}, {"users", "info"}} {
		cmd := findCmd(root, p...)
		require.NotNil(t, cmd)
		assert.True(t, sel(cmd), "API command %v must be exposed as an MCP tool", p)
	}
}

func TestMCPCommandRegistered(t *testing.T) {
	require.NotNil(t, findCmd(NewRootCmd(), "mcp"), "the mcp subtree must be registered")
}

// readShapedMethod mirrors the hook's raw-api gate: a Slack read's final dot-segment is a
// get* name or one of the read nouns. The classification invariant below leans on it.
func readShapedMethod(method string) bool {
	seg := method[strings.LastIndex(method, ".")+1:]
	if strings.HasPrefix(strings.ToLower(seg), "get") {
		return true
	}
	switch strings.ToLower(seg) {
	case "list", "info", "test", "history", "replies", "members", "messages",
		"files", "all", "lookupbyemail", "conversations", "teaminfo", "context", "lookup":
		return true
	}
	return false
}

func TestClassifyAPICommands(t *testing.T) {
	c := classifyAPICommands(false)

	has := func(set []apiCmdInfo, method string) bool {
		for _, x := range set {
			if x.Method == method {
				return true
			}
		}
		return false
	}
	assert.True(t, has(c.Read, "conversations.list"), "conversations.list is read-only")
	assert.True(t, has(c.Read, "search.messages"), "search.messages is read-only")
	assert.True(t, has(c.Write, "chat.postMessage"), "chat.postMessage is a write")
	assert.True(t, has(c.Write, "conversations.mark"), "conversations.mark is a write")
	assert.True(t, has(c.Destructive, "chat.delete"), "chat.delete is destructive")
	assert.True(t, has(c.Destructive, "chat.deleteScheduledMessage"), "cancelling a scheduled message is destructive")
	assert.True(t, has(c.Destructive, "conversations.archive"), "archive is destructive")
	assert.True(t, has(c.Destructive, "conversations.leave"), "leave is destructive")
	assert.True(t, has(c.Destructive, "conversations.kick"), "kick is destructive")
	assert.True(t, has(c.Destructive, "usergroups.disable"), "usergroups.disable is destructive")
	assert.False(t, has(c.Write, "chat.delete"), "chat.delete must not be a mere write")

	// Invariant: nothing in the read (allowed) bucket may mutate remote state. Every Slack
	// read ends in a read-shaped segment (list/info/get*/…); anything else in Read is a
	// write classified as read — the verb-name-collision escape (GOAL.md §3b #4).
	for _, r := range c.Read {
		assert.Truef(t, readShapedMethod(r.Method),
			"read bucket contains non-read-shaped method %s (%s) — a write classified as read", r.Method, r.Path)
	}

	// --all-writes promotes ordinary writes into the hard-block bucket.
	strict := classifyAPICommands(true)
	assert.True(t, has(strict.Destructive, "chat.postMessage"))
	assert.Empty(t, strict.Write)
}

// TestEveryAPICommandIsAnnotated walks the BUILT tree: every runnable leaf outside the
// explicit local/meta groups must carry MCP hint annotations. A hand-built command added
// without markKind would otherwise fall through the agent guard as "local/utility" — the
// annotation-gap escape (GOAL.md §3b #4).
func TestEveryAPICommandIsAnnotated(t *testing.T) {
	localGroups := map[string]bool{
		"auth": true, "config": true, "init": true, "doctor": true, "completion": true,
		"alias": true, "version": true, "mcp": true, "agent": true, "help": true,
		// `api` is the raw escape hatch: excluded from MCP entirely and gated by the hook
		// at the method position, so it carries no per-command annotation.
		"api": true,
	}
	root := NewRootCmd()
	var walk func(c *cobra.Command, top string)
	walk = func(c *cobra.Command, top string) {
		for _, sub := range c.Commands() {
			name := top
			if name == "" {
				name = sub.Name()
			}
			if localGroups[name] {
				continue
			}
			if sub.Runnable() {
				ann := sub.Annotations
				hasHint := ann[annReadOnly] == "true" || ann[annOpenWorld] == "true"
				assert.Truef(t, hasHint, "command %q is un-annotated — the agent guard would misclassify it", sub.CommandPath())
			}
			walk(sub, name)
		}
	}
	walk(root, "")
}

// claudeCodeSettingsMarker separates the hook-script section from the settings fragment
// in the claude-code guard output.
const claudeCodeSettingsMarker = "# ----- merge into .claude/settings.json -----\n"

func TestAgentGuard_ClaudeCode(t *testing.T) {
	out, _, err := run(t, nil, "agent", "guard", "--host", "claude-code")
	require.NoError(t, err)

	idx := strings.Index(out, claudeCodeSettingsMarker)
	require.GreaterOrEqual(t, idx, 0, "claude-code output must contain the settings section marker")
	hook := out[:idx]
	settingsJSON := out[idx+len(claudeCodeSettingsMarker):]

	// Hook script: path-prefix-hardened command matching, api gate, MCP branch.
	assert.Contains(t, hook, "#!/usr/bin/env bash")
	assert.Contains(t, hook, `([^[:space:]]*/)?slackctl`, "hook regex must accept a path-prefixed binary")
	assert.Contains(t, hook, "'msg delete'")
	assert.Contains(t, hook, "'chat delete'", "hook must block alias paths too")
	assert.Contains(t, hook, "'channels archive'", "group-alias × verb cross-product must be enumerated")
	assert.Contains(t, hook, "'alias set'")
	assert.Contains(t, hook, "'mcp__slackctl__slack_msg_delete'")
	assert.Contains(t, hook, "api_is_blocked")

	var settings struct {
		Permissions struct {
			Deny  []string `json:"deny"`
			Ask   []string `json:"ask"`
			Allow []string `json:"allow"`
		} `json:"permissions"`
		Hooks struct {
			PreToolUse []struct {
				Matcher string `json:"matcher"`
			} `json:"PreToolUse"`
		} `json:"hooks"`
	}
	require.NoError(t, json.Unmarshal([]byte(settingsJSON), &settings))
	assert.Contains(t, settings.Permissions.Deny, "Bash(slackctl msg delete:*)")
	assert.Contains(t, settings.Permissions.Deny, "mcp__slackctl__slack_msg_delete")
	// Alias paths must be denied too, or `slackctl chat delete` / `slackctl channels
	// archive` bypass the rules that only name the canonical path.
	assert.Contains(t, settings.Permissions.Deny, "Bash(slackctl chat delete:*)")
	assert.Contains(t, settings.Permissions.Deny, "Bash(slackctl message delete:*)")
	assert.Contains(t, settings.Permissions.Deny, "Bash(slackctl channels archive:*)")
	assert.Contains(t, settings.Permissions.Deny, "Bash(slackctl conv kick:*)")
	// Raw api escape hatch: destructive Web API methods denied by name.
	assert.Contains(t, settings.Permissions.Deny, "Bash(slackctl api chat.delete:*)")
	assert.Contains(t, settings.Permissions.Deny, "Bash(slackctl api conversations.archive:*)")
	// Alias minting is denied.
	assert.Contains(t, settings.Permissions.Deny, "Bash(slackctl alias set:*)")
	assert.Contains(t, settings.Permissions.Ask, "Bash(slackctl msg post:*)")
	assert.Contains(t, settings.Permissions.Allow, "Bash(slackctl conversations list:*)")
	// A destructive op must never appear in allow — canonical or alias path.
	for _, p := range settings.Permissions.Allow {
		assert.NotContains(t, settings.Permissions.Deny, p, "allow and deny must not overlap: %s", p)
	}
	assert.NotContains(t, settings.Permissions.Allow, "Bash(slackctl conversations leave:*)")
	// The hook must be wired for both the Bash and MCP surfaces.
	require.Len(t, settings.Hooks.PreToolUse, 2)
	assert.Equal(t, "Bash", settings.Hooks.PreToolUse[0].Matcher)
	assert.Equal(t, "mcp__slackctl__", settings.Hooks.PreToolUse[1].Matcher)
}

// TestAPICommandAliasPaths pins the alias expansion the guard depends on.
func TestAPICommandAliasPaths(t *testing.T) {
	var msgDelete *apiCmdInfo
	for i := range registeredAPICmds {
		if registeredAPICmds[i].Path == "msg delete" {
			msgDelete = &registeredAPICmds[i]
			break
		}
	}
	require.NotNil(t, msgDelete)
	all := msgDelete.AllPaths()
	assert.Contains(t, all, "msg delete")
	assert.Contains(t, all, "chat delete")
	assert.Contains(t, all, "message delete")
}

func TestAgentGuard_Codex(t *testing.T) {
	out, _, err := run(t, nil, "agent", "guard", "--host", "codex")
	require.NoError(t, err)
	// Codex reads TOP-LEVEL keys; an invented [sandbox] table is silently ignored.
	assert.Contains(t, out, `approval_policy = "on-request"`)
	assert.Contains(t, out, `sandbox_mode = "read-only"`)
	assert.NotContains(t, out, "[sandbox]", "a [sandbox] table is dead config Codex ignores")
	assert.Contains(t, out, "slackctl msg delete")
}

func TestAgentGuard_OpenCode(t *testing.T) {
	out, _, err := run(t, nil, "agent", "guard", "--host", "opencode", "--all-writes")
	require.NoError(t, err)
	var cfg struct {
		Permission map[string]string `json:"permission"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &cfg))
	assert.Equal(t, "deny", cfg.Permission["Bash(slackctl msg delete:*)"])
	// With --all-writes, chat.postMessage is hard-denied too.
	assert.Equal(t, "deny", cfg.Permission["Bash(slackctl msg post:*)"])
	assert.Equal(t, "allow", cfg.Permission["Bash(slackctl conversations list:*)"])
}

func TestAgentGuard_UnknownHost(t *testing.T) {
	_, _, err := run(t, nil, "agent", "guard", "--host", "bogus")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown --host")
}

func TestAgentGuard_WriteToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	_, _, err := run(t, nil, "agent", "guard", "--host", "claude-code", "--out", path)
	require.NoError(t, err)
	data, rerr := os.ReadFile(path)
	require.NoError(t, rerr)
	assert.Contains(t, string(data), "permissions")
}
