package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

func TestTokenKind_Valid(t *testing.T) {
	for _, k := range []TokenKind{KindBot, KindUser, KindApp, KindSession} {
		assert.True(t, k.Valid(), string(k))
	}
	assert.False(t, TokenKind("nope").Valid())
	assert.False(t, TokenKind("").Valid())
}

func TestTokenKind_EnvVar(t *testing.T) {
	assert.Equal(t, "SLACK_BOT_TOKEN", KindBot.EnvVar())
	assert.Equal(t, "SLACK_USER_TOKEN", KindUser.EnvVar())
	assert.Equal(t, "SLACK_APP_TOKEN", KindApp.EnvVar())
	assert.Equal(t, "SLACK_XOXC_TOKEN", KindSession.EnvVar())
}

func TestKey(t *testing.T) {
	assert.Equal(t, "acme", Key("acme", KindBot), "the bot token keeps the bare profile name")
	assert.Equal(t, "acme", Key("acme", ""), "empty kind is treated as bot")
	assert.Equal(t, "acme#user", Key("acme", KindUser))
	assert.Equal(t, "acme#app", Key("acme", KindApp))
	assert.Equal(t, "acme#session", Key("acme", KindSession))
}

func TestSessionCreds_RoundTrip(t *testing.T) {
	keyring.MockInit()
	store := New(t.TempDir())
	require.NoError(t, SetSession(store, "acme", SessionCreds{Token: "xoxc-a", Cookie: "xoxd-b"}))

	got, err := GetSession(store, "acme")
	require.NoError(t, err)
	assert.Equal(t, "xoxc-a", got.Token)
	assert.Equal(t, "xoxd-b", got.Cookie)

	// The blob lives under the session-suffixed key, not the bare profile.
	_, err = store.Get(Key("acme", KindSession))
	require.NoError(t, err)
	_, err = store.Get("acme")
	assert.Error(t, err, "the bot slot must stay empty")
}

func TestGetSession_NotFoundAndCorrupt(t *testing.T) {
	keyring.MockInit()
	store := New(t.TempDir())
	_, err := GetSession(store, "missing")
	require.Error(t, err)

	// A corrupt / incomplete blob is a clear error, not a silent empty pair.
	require.NoError(t, store.Set(Key("acme", KindSession), `{"xoxc":"only-token"}`))
	_, err = GetSession(store, "acme")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "incomplete")

	require.NoError(t, store.Set(Key("beta", KindSession), `not json`))
	_, err = GetSession(store, "beta")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "corrupt")
}
