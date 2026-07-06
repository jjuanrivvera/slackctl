package slackevent

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMeta(t *testing.T) {
	m, err := ParseMeta(json.RawMessage(`{"type":"message","channel":"D111","channel_type":"im","user":"U1","ts":"5.0"}`))
	require.NoError(t, err)
	assert.Equal(t, "message", m.Type)
	assert.Equal(t, "D111", m.ChannelOf())
	assert.True(t, m.IsDM())

	_, err = ParseMeta(json.RawMessage(`not json`))
	assert.Error(t, err)
}

func TestChannelOf_FallsBackToItem(t *testing.T) {
	m, err := ParseMeta(json.RawMessage(`{"type":"reaction_added","user":"U1","item":{"channel":"C42"}}`))
	require.NoError(t, err)
	assert.Equal(t, "C42", m.ChannelOf(), "reactions carry the channel under item")
	assert.False(t, m.IsDM())
}

func TestIsDM(t *testing.T) {
	// By channel_type (Socket Mode / Events API events).
	assert.True(t, Meta{ChannelType: "im"}.IsDM())
	// By channel-id prefix (RTM events omit channel_type).
	assert.True(t, Meta{Channel: "D0ABC"}.IsDM())
	assert.False(t, Meta{Channel: "C0ABC"}.IsDM())
	assert.False(t, Meta{}.IsDM())
}
