package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestHookScript_BashExecution exercises the generated hook script with real bash to verify
// the adversarial cases: obfuscation, path-invoked binaries, alias paths, the raw api
// escape hatch, and the benign lookalikes that must stay allowed. Gated on a POSIX shell
// being available so it is safe in the regular suite.
func TestHookScript_BashExecution(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("bash hook tests require a POSIX shell; skipping on windows")
	}
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not found in PATH; skipping hook execution tests")
	}

	// Generate the hook from the real classification so blocked_cmds/blocked_tools are
	// fully populated (canonical + alias paths).
	hookContent := hookScript(classifyAPICommands(false))
	tmpDir := t.TempDir()
	hookFile := filepath.Join(tmpDir, "slackctl-guard.sh")
	if err := os.WriteFile(hookFile, []byte(hookContent), 0o755); err != nil { // #nosec G306 -- hook must be executable
		t.Fatalf("write hook: %v", err)
	}

	bashPayload := func(command string) string {
		b, _ := json.Marshal(map[string]any{
			"tool_name":  "Bash",
			"tool_input": map[string]any{"command": command},
		})
		return string(b)
	}
	mcpPayload := func(toolName string) string {
		b, _ := json.Marshal(map[string]any{
			"tool_name":  toolName,
			"tool_input": map[string]any{},
		})
		return string(b)
	}

	runHook := func(t *testing.T, payload string) string {
		t.Helper()
		cmd := exec.Command(bash, hookFile)
		cmd.Stdin = strings.NewReader(payload)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		// The hook always exits 0; the decision is in the JSON output.
		if err := cmd.Run(); err != nil {
			t.Logf("hook output: %s", out.String())
			t.Fatalf("hook script exited non-zero: %v", err)
		}
		return out.String()
	}

	isDenied := func(output string) bool {
		return strings.Contains(output, `"permissionDecision":"deny"`)
	}

	cases := []struct {
		name       string
		payload    string
		wantDenied bool
	}{
		// --- direct blocked commands ---
		{"msg_delete_denied", bashPayload("slackctl msg delete --channel C1 --ts 1.0"), true},
		{"conversations_archive_denied", bashPayload("slackctl conversations archive --channel C1"), true},
		{"conversations_leave_denied", bashPayload("slackctl conversations leave --channel C1"), true},
		{"usergroups_disable_denied", bashPayload("slackctl usergroups disable --usergroup S1"), true},
		// --- cobra alias paths (bypass without enumeration) ---
		{"chat_alias_delete_denied", bashPayload("slackctl chat delete --channel C1 --ts 1.0"), true},
		{"message_alias_delete_denied", bashPayload("slackctl message delete --channel C1 --ts 1.0"), true},
		{"channels_alias_archive_denied", bashPayload("slackctl channels archive --channel C1"), true},
		{"conv_alias_kick_denied", bashPayload("slackctl conv kick --channel C1 --user U1"), true},
		{"groups_alias_disable_denied", bashPayload("slackctl groups disable --usergroup S1"), true},
		// --- alias minting ---
		{"alias_set_denied", bashPayload(`slackctl alias set kill "msg delete"`), true},
		// --- obfuscation ---
		{"quote_split_denied", bashPayload(`slackctl msg de""lete --channel C1 --ts 1.0`), true},
		{"single_quote_split_denied", bashPayload(`slackctl conversations le''ave --channel C1`), true},
		{"backslash_denied", bashPayload(`slackctl msg de\lete --channel C1 --ts 1.0`), true},
		{"newline_continuation_denied", bashPayload("slackctl msg \\\ndelete --channel C1 --ts 1.0"), true},
		// --- command position after separators ---
		{"after_semicolon_denied", bashPayload("true; slackctl msg delete --channel C1 --ts 1.0"), true},
		{"after_pipe_denied", bashPayload("echo hi | slackctl msg delete --channel C1 --ts 1.0"), true},
		{"after_and_denied", bashPayload("true && slackctl conversations leave --channel C1"), true},
		{"trailing_separator_denied", bashPayload("slackctl conversations archive --channel C1;true"), true},
		{"env_prefix_denied", bashPayload("env SLACK_BOT_TOKEN=x slackctl conversations leave --channel C1"), true},
		// --- path-invoked binaries ---
		{"relative_path_binary_denied", bashPayload("./bin/slackctl msg delete --channel C1 --ts 1.0"), true},
		{"absolute_path_binary_denied", bashPayload("/usr/local/bin/slackctl msg delete --channel C1 --ts 1.0"), true},
		{"absolute_path_api_denied", bashPayload("/usr/local/bin/slackctl api chat.delete"), true},
		// --- raw api escape hatch (method position; only read-shaped segments pass) ---
		{"api_chat_delete_denied", bashPayload("slackctl api chat.delete -q channel=C1 -q ts=1.0"), true},
		{"api_uppercase_denied", bashPayload("slackctl api CHAT.DELETE -q channel=C1"), true},
		{"api_post_message_denied", bashPayload("slackctl api chat.postMessage -q channel=C1 -q text=hi"), true},
		{"api_set_topic_denied", bashPayload("slackctl api conversations.setTopic -q channel=C1 -q topic=x"), true},
		{"api_admin_denied", bashPayload("slackctl api admin.conversations.delete -q channel_id=C1"), true},
		{"api_flag_before_method_denied", bashPayload("slackctl api -q channel=C1 chat.delete"), true},
		{"api_compound_read_then_delete_denied", bashPayload("slackctl api auth.test;slackctl api chat.delete"), true},
		// --- raw api reads stay allowed ---
		{"api_auth_test_allowed", bashPayload("slackctl api auth.test"), false},
		{"api_conversations_list_allowed", bashPayload("slackctl api conversations.list -q types=im"), false},
		{"api_history_allowed", bashPayload("slackctl api conversations.history -q channel=C1"), false},
		{"api_get_presence_allowed", bashPayload("slackctl api users.getPresence -q user=U1"), false},
		{"api_search_allowed", bashPayload("slackctl api search.messages -q query=deploy"), false},
		{"api_delete_in_param_allowed", bashPayload("slackctl api conversations.info -q channel=delete_club"), false},
		// --- benign lookalikes that must stay allowed ---
		{"conversations_list_allowed", bashPayload("slackctl conversations list"), false},
		{"post_with_delete_in_arg_allowed", bashPayload(`slackctl msg post --channel C1 --text "how to delete a message"`), false},
		{"cat_file_allowed", bashPayload("cat msg_delete.go"), false},
		{"other_binary_allowed", bashPayload("myslackctl msg delete --channel C1 --ts 1.0"), false},
		{"other_binary_api_allowed", bashPayload("myslackctl api chat.delete"), false},
		{"saved_remove_is_write_not_blocked", bashPayload("slackctl saved remove --channel C1 --ts 1.0"), false},
		// --- MCP branch ---
		{"mcp_msg_delete_denied", mcpPayload("mcp__slackctl__slack_msg_delete"), true},
		{"mcp_conversations_archive_denied", mcpPayload("mcp__slackctl__slack_conversations_archive"), true},
		{"mcp_conversations_list_allowed", mcpPayload("mcp__slackctl__slack_conversations_list"), false},
		{"mcp_near_miss_allowed", mcpPayload("mcp__slackctl__slack_msg_delete2"), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := runHook(t, tc.payload)
			if denied := isDenied(output); denied != tc.wantDenied {
				t.Errorf("want denied=%v, got denied=%v\noutput: %s", tc.wantDenied, denied, output)
			}
		})
	}
}

// TestHookScript_BashExecutionNoJq exercises the no-jq fallback path with a STRICT PATH: a
// bin dir holding only the POSIX tools the hook needs, so jq is genuinely unreachable
// (merely prepending an empty dir leaves jq resolvable — the test flaw that masked a
// fail-open bug in two audited repos, GOAL.md §3b #3).
func TestHookScript_BashExecutionNoJq(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("bash hook tests require a POSIX shell; skipping on windows")
	}
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash not found in PATH; skipping hook execution tests")
	}

	hookContent := hookScript(classifyAPICommands(false))
	tmpDir := t.TempDir()
	hookFile := filepath.Join(tmpDir, "slackctl-guard.sh")
	if err := os.WriteFile(hookFile, []byte(hookContent), 0o755); err != nil { // #nosec G306 -- hook must be executable
		t.Fatalf("write hook: %v", err)
	}

	binDir := filepath.Join(tmpDir, "nojq-bin")
	if err := os.Mkdir(binDir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for _, tool := range []string{"cat", "tr", "grep", "sed", "printf", "env"} {
		p, lerr := exec.LookPath(tool)
		if lerr != nil {
			continue // shell builtins (printf) need no symlink
		}
		if serr := os.Symlink(p, filepath.Join(binDir, tool)); serr != nil {
			t.Fatalf("symlink %s: %v", tool, serr)
		}
	}

	bashPayload := func(command string) string {
		b, _ := json.Marshal(map[string]any{
			"tool_name":  "Bash",
			"tool_input": map[string]any{"command": command},
		})
		return string(b)
	}

	runHookNoJq := func(t *testing.T, payload string) string {
		t.Helper()
		cmd := exec.Command(bash, hookFile)
		cmd.Stdin = strings.NewReader(payload)
		env := make([]string, 0, len(os.Environ()))
		for _, e := range os.Environ() {
			if !strings.HasPrefix(e, "PATH=") {
				env = append(env, e)
			}
		}
		cmd.Env = append(env, "PATH="+binDir)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		if err := cmd.Run(); err != nil {
			t.Logf("hook output: %s", out.String())
			t.Fatalf("hook script exited non-zero: %v", err)
		}
		return out.String()
	}

	isDenied := func(output string) bool {
		return strings.Contains(output, `"permissionDecision":"deny"`)
	}

	cases := []struct {
		name       string
		payload    string
		wantDenied bool
	}{
		{"nojq_msg_delete_denied", bashPayload("slackctl msg delete --channel C1 --ts 1.0"), true},
		{"nojq_obfuscated_delete_denied", bashPayload(`slackctl msg de""lete --channel C1 --ts 1.0`), true},
		{"nojq_path_binary_denied", bashPayload("./bin/slackctl msg delete --channel C1 --ts 1.0"), true},
		{"nojq_api_delete_denied", bashPayload("slackctl api chat.delete"), true},
		{"nojq_cat_file_allowed", bashPayload("cat msg_delete.go"), false},
		{"nojq_post_allowed", bashPayload(`slackctl msg post --channel C1 --text "delete this later"`), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := runHookNoJq(t, tc.payload)
			if denied := isDenied(output); denied != tc.wantDenied {
				t.Errorf("want denied=%v, got denied=%v\noutput: %s", tc.wantDenied, denied, output)
			}
		})
	}
}
