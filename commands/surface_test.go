package commands

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVerbs_MockedAPI exercises every declared verb against the mocked Web API: each
// command must reach its method, render the result, and exit cleanly. Together with
// commands_test.go this walks the complete manifest surface.
func TestVerbs_MockedAPI(t *testing.T) {
	cases := []struct {
		name   string
		method string
		body   string
		want   string
		args   []string
	}{
		// conversations — the verbs not covered elsewhere
		{"conversations replies", "conversations.replies", `{"ok":true,"messages":[{"ts":"1.1","user":"U1","type":"message","text":"reply"}]}`, "reply",
			[]string{"conversations", "replies", "--channel", "C1", "--ts", "1.0"}},
		{"conversations members", "conversations.members", `{"ok":true,"members":["U1","U2"]}`, "U2",
			[]string{"conversations", "members", "--channel", "C1", "-o", "json"}},
		{"conversations create", "conversations.create", `{"ok":true,"channel":{"id":"C9","name":"eng-alerts","is_private":false}}`, "eng-alerts",
			[]string{"conversations", "create", "--name", "eng-alerts"}},
		{"conversations rename", "conversations.rename", `{"ok":true,"channel":{"id":"C9","name":"eng-v2"}}`, "eng-v2",
			[]string{"conversations", "rename", "--channel", "C9", "--name", "eng-v2"}},
		{"conversations archive", "conversations.archive", `{"ok":true}`, "true",
			[]string{"conversations", "archive", "--channel", "C9", "-o", "json"}},
		{"conversations unarchive", "conversations.unarchive", `{"ok":true}`, "true",
			[]string{"conversations", "unarchive", "--channel", "C9", "-o", "json"}},
		{"conversations invite", "conversations.invite", `{"ok":true,"channel":{"id":"C9"}}`, "C9",
			[]string{"conversations", "invite", "--channel", "C9", "--users", "U1,U2", "-o", "json"}},
		{"conversations leave", "conversations.leave", `{"ok":true}`, "true",
			[]string{"conversations", "leave", "--channel", "C9", "-o", "json"}},
		{"conversations kick", "conversations.kick", `{"ok":true}`, "true",
			[]string{"conversations", "kick", "--channel", "C9", "--user", "U1", "-o", "json"}},
		{"conversations set-topic", "conversations.setTopic", `{"ok":true,"channel":{"id":"C9","topic":{"value":"deploys"}}}`, "deploys",
			[]string{"conversations", "set-topic", "--channel", "C9", "--topic", "deploys", "-o", "json"}},
		{"conversations set-purpose", "conversations.setPurpose", `{"ok":true,"channel":{"id":"C9"}}`, "C9",
			[]string{"conversations", "set-purpose", "--channel", "C9", "--purpose", "alerts", "-o", "json"}},
		{"conversations mark", "conversations.mark", `{"ok":true}`, "true",
			[]string{"conversations", "mark", "--channel", "C9", "--ts", "1.0", "-o", "json"}},
		{"conversations open", "conversations.open", `{"ok":true,"channel":{"id":"D42"}}`, "D42",
			[]string{"conversations", "open", "--users", "U1"}},
		{"conversations close", "conversations.close", `{"ok":true}`, "true",
			[]string{"conversations", "close", "--channel", "D42", "-o", "json"}},

		// msg — the verbs not covered elsewhere
		{"msg update", "chat.update", `{"ok":true,"channel":"C1","ts":"1.0","text":"fixed"}`, "fixed",
			[]string{"msg", "update", "--channel", "C1", "--ts", "1.0", "--text", "fixed"}},
		{"msg delete", "chat.delete", `{"ok":true,"channel":"C1","ts":"1.0"}`, "1.0",
			[]string{"msg", "delete", "--channel", "C1", "--ts", "1.0", "-o", "json"}},
		{"msg ephemeral", "chat.postEphemeral", `{"ok":true,"message_ts":"1.5"}`, "1.5",
			[]string{"msg", "ephemeral", "--channel", "C1", "--user", "U1", "--text", "psst", "-o", "json"}},
		{"msg me", "chat.meMessage", `{"ok":true,"channel":"C1","ts":"1.6"}`, "1.6",
			[]string{"msg", "me", "--channel", "C1", "--text", "is deploying", "-o", "json"}},
		{"msg permalink", "chat.getPermalink", `{"ok":true,"permalink":"https://acme.slack.com/archives/C1/p1"}`, "archives",
			[]string{"msg", "permalink", "--channel", "C1", "--ts", "1.0"}},
		{"msg schedule", "chat.scheduleMessage", `{"ok":true,"scheduled_message_id":"Q1","channel":"C1","post_at":1735689600}`, "Q1",
			[]string{"msg", "schedule", "--channel", "C1", "--post-at", "1735689600", "--text", "hny"}},
		{"msg scheduled", "chat.scheduledMessages.list", `{"ok":true,"scheduled_messages":[{"id":"Q1","channel_id":"C1","post_at":1735689600,"text":"hny"}]}`, "Q1",
			[]string{"msg", "scheduled"}},
		{"msg delete-scheduled", "chat.deleteScheduledMessage", `{"ok":true}`, "true",
			[]string{"msg", "delete-scheduled", "--channel", "C1", "--id", "Q1", "-o", "json"}},

		// search files / all
		{"search files", "search.files", `{"ok":true,"files":{"matches":[{"id":"F1","name":"report.pdf","title":"Q report","user":"U1"}]}}`, "report.pdf",
			[]string{"search", "files", "--query", "report"}},
		{"search all", "search.all", `{"ok":true,"messages":{"matches":[]},"files":{"matches":[]}}`, "ok",
			[]string{"search", "all", "--query", "x", "-o", "json"}},

		// users
		{"users list", "users.list", `{"ok":true,"members":[{"id":"U1","name":"ada","real_name":"Ada","is_bot":false,"deleted":false}]}`, "ada",
			[]string{"users", "list"}},
		{"users info", "users.info", `{"ok":true,"user":{"id":"U1","name":"ada"}}`, "ada",
			[]string{"users", "info", "--user", "U1", "-o", "json"}},
		{"users lookup-email", "users.lookupByEmail", `{"ok":true,"user":{"id":"U1","name":"ada","real_name":"Ada"}}`, "U1",
			[]string{"users", "lookup-email", "--email", "ada@example.com"}},
		{"users conversations", "users.conversations", `{"ok":true,"channels":[{"id":"C1","name":"general","is_private":false,"is_im":false}]}`, "general",
			[]string{"users", "conversations"}},
		{"users presence", "users.getPresence", `{"ok":true,"presence":"active"}`, "active",
			[]string{"users", "presence", "--user", "U1", "-o", "json"}},
		{"users profile", "users.profile.get", `{"ok":true,"profile":{"real_name":"Ada","email":"ada@example.com"}}`, "Ada",
			[]string{"users", "profile", "--user", "U1", "-o", "json"}},

		// usergroups
		{"usergroups list", "usergroups.list", `{"ok":true,"usergroups":[{"id":"S1","handle":"oncall","name":"On-call","user_count":3}]}`, "oncall",
			[]string{"usergroups", "list"}},
		{"usergroups create", "usergroups.create", `{"ok":true,"usergroup":{"id":"S2","handle":"eng","name":"Engineers"}}`, "S2",
			[]string{"usergroups", "create", "--name", "Engineers", "--handle", "eng"}},
		{"usergroups update", "usergroups.update", `{"ok":true,"usergroup":{"id":"S2","name":"Engs"}}`, "Engs",
			[]string{"usergroups", "update", "--usergroup", "S2", "--name", "Engs", "-o", "json"}},
		{"usergroups enable", "usergroups.enable", `{"ok":true,"usergroup":{"id":"S2"}}`, "S2",
			[]string{"usergroups", "enable", "--usergroup", "S2", "-o", "json"}},
		{"usergroups disable", "usergroups.disable", `{"ok":true,"usergroup":{"id":"S2"}}`, "S2",
			[]string{"usergroups", "disable", "--usergroup", "S2", "-o", "json"}},
		{"usergroups members", "usergroups.users.list", `{"ok":true,"users":["U1","U2"]}`, "U2",
			[]string{"usergroups", "members", "--usergroup", "S2", "-o", "json"}},
		{"usergroups members-update", "usergroups.users.update", `{"ok":true,"usergroup":{"id":"S2","user_count":2}}`, "S2",
			[]string{"usergroups", "members-update", "--usergroup", "S2", "--users", "U1,U2", "-o", "json"}},

		// reactions
		{"reactions add", "reactions.add", `{"ok":true}`, "true",
			[]string{"reactions", "add", "--channel", "C1", "--ts", "1.0", "--name", "thumbsup", "-o", "json"}},
		{"reactions remove", "reactions.remove", `{"ok":true}`, "true",
			[]string{"reactions", "remove", "--channel", "C1", "--ts", "1.0", "--name", "thumbsup", "-o", "json"}},
		{"reactions get", "reactions.get", `{"ok":true,"type":"message","message":{"reactions":[{"name":"thumbsup","count":2}]}}`, "thumbsup",
			[]string{"reactions", "get", "--channel", "C1", "--ts", "1.0", "-o", "json"}},
		{"reactions list", "reactions.list", `{"ok":true,"items":[{"type":"message","channel":"C1"}]}`, "message",
			[]string{"reactions", "list"}},

		// saved (user token) — list covered; add/remove here
		{"saved add", "stars.add", `{"ok":true}`, "true",
			[]string{"saved", "add", "--channel", "C1", "--ts", "1.0", "-o", "json"}},
		{"saved remove", "stars.remove", `{"ok":true}`, "true",
			[]string{"saved", "remove", "--channel", "C1", "--ts", "1.0", "-o", "json"}},

		// pins
		{"pins list", "pins.list", `{"ok":true,"items":[{"type":"message","channel":"C1","created":1720000000}]}`, "message",
			[]string{"pins", "list", "--channel", "C1"}},
		{"pins add", "pins.add", `{"ok":true}`, "true",
			[]string{"pins", "add", "--channel", "C1", "--ts", "1.0", "-o", "json"}},
		{"pins remove", "pins.remove", `{"ok":true}`, "true",
			[]string{"pins", "remove", "--channel", "C1", "--ts", "1.0", "-o", "json"}},

		// emoji + team
		{"emoji list", "emoji.list", `{"ok":true,"emoji":{"party_parrot":"https://emoji.example/pp.gif"}}`, "party_parrot",
			[]string{"emoji", "list", "-o", "json"}},
		{"team info", "team.info", `{"ok":true,"team":{"id":"T1","name":"Acme","domain":"acme"}}`, "Acme",
			[]string{"team", "info"}},
		{"team profile", "team.profile.get", `{"ok":true,"profile":{"fields":[{"id":"Xf1","label":"Role"}]}}`, "Role",
			[]string{"team", "profile", "-o", "json"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := newServer(t, routes{tc.method: tc.body})
			out, errb, err := run(t, srv, tc.args...)
			require.NoError(t, err, errb)
			assert.Contains(t, out, tc.want)
		})
	}
}

// TestManifestVerbsAllReachable mirrors scripts/spec-check.sh inside the test suite: every
// registered method command path must resolve in a fresh tree.
func TestManifestVerbsAllReachable(t *testing.T) {
	root := NewRootCmd()
	for _, info := range APICommands() {
		cmd := findCmd(root, strings.Fields(info.PathString())...)
		assert.NotNilf(t, cmd, "registered command %q not reachable in the tree", info.PathString())
	}
}
