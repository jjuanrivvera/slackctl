package commands

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDndInfo(t *testing.T) {
	srv := newServer(t, routes{
		"dnd.info": `{"ok":true,"dnd_enabled":true,"next_dnd_start_ts":1720000000,"next_dnd_end_ts":1720003600,"snooze_enabled":false}`,
	})
	out, _, err := run(t, srv, "dnd", "info")
	require.NoError(t, err)
	mustContain(t, out, "DND_ENABLED")
	mustContain(t, out, "true")
}

func TestDndSetSnooze_UsesUserToken(t *testing.T) {
	var authz, minutes string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz = r.Header.Get("Authorization")
		_ = r.ParseForm()
		minutes = r.PostForm.Get("num_minutes")
		_, _ = w.Write([]byte(`{"ok":true,"snooze_enabled":true,"snooze_endtime":1720003600}`))
	}))
	t.Cleanup(srv.Close)
	_, _, err := run(t, srv, "dnd", "set-snooze", "--minutes", "60")
	require.NoError(t, err)
	assert.Equal(t, "Bearer xoxp-test-token", authz, "dnd.setSnooze is user-only")
	assert.Equal(t, "60", minutes)
}

func TestDndSetSnooze_FailsWithoutUserToken(t *testing.T) {
	_, _, err := runNoToken(t, nil, "", "dnd", "set-snooze", "--minutes", "30")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--kind user")
}

func TestUsersSetPresence(t *testing.T) {
	var presence string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		presence = r.PostForm.Get("presence")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)
	_, _, err := run(t, srv, "users", "set-presence", "--presence", "away", "-o", "json")
	require.NoError(t, err)
	assert.Equal(t, "away", presence)
}

func TestUsersSetStatus_BuildsProfileObject(t *testing.T) {
	var authz, profile string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz = r.Header.Get("Authorization")
		_ = r.ParseForm()
		profile = r.PostForm.Get("profile")
		_, _ = w.Write([]byte(`{"ok":true,"profile":{"status_text":"In a meeting","status_emoji":":calendar:"}}`))
	}))
	t.Cleanup(srv.Close)
	out, _, err := run(t, srv, "users", "set-status", "--text", "In a meeting", "--emoji", ":calendar:", "-o", "json")
	require.NoError(t, err)
	assert.Equal(t, "Bearer xoxp-test-token", authz, "profile.set is user-only")
	assert.JSONEq(t, `{"status_text":"In a meeting","status_emoji":":calendar:","status_expiration":0}`, profile)
	mustContain(t, out, "In a meeting")
}

func TestDndTeamInfo(t *testing.T) {
	srv := newServer(t, routes{
		"dnd.teamInfo": `{"ok":true,"users":{"U1":{"dnd_enabled":true},"U2":{"dnd_enabled":false}}}`,
	})
	out, _, err := run(t, srv, "dnd", "team-info", "--users", "U1,U2", "-o", "json")
	require.NoError(t, err)
	mustContain(t, out, "U1")
	mustContain(t, out, "dnd_enabled")
}
