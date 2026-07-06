// Package slackevent holds the event shape shared by both streaming transports — Socket
// Mode (app token) and RTM (session/user token). Both ultimately deliver the same Slack
// event objects (message, reaction_added, …); this package is where the filter metadata and
// its helpers live so the `listen` command treats every source uniformly.
package slackevent

import "encoding/json"

// Meta is the subset of every event used for filtering.
type Meta struct {
	Type        string `json:"type"`
	Channel     string `json:"channel"`
	ChannelType string `json:"channel_type"` // im|mpim|channel|group on message events
	User        string `json:"user"`
	TS          string `json:"ts"`
	Item        struct {
		Channel string `json:"channel"`
	} `json:"item"` // reaction_added/removed carry the channel here
}

// ParseMeta decodes the filter metadata from a raw Slack event object (the shape RTM
// delivers directly, and the inner `event` of an Events API callback).
func ParseMeta(rawEvent json.RawMessage) (Meta, error) {
	var m Meta
	if err := json.Unmarshal(rawEvent, &m); err != nil {
		return Meta{}, err
	}
	return m, nil
}

// ChannelOf returns the event's conversation id wherever the event type carries it.
func (m Meta) ChannelOf() string {
	if m.Channel != "" {
		return m.Channel
	}
	return m.Item.Channel
}

// IsDM reports whether the event happened in a direct message. RTM events omit
// channel_type, so the channel-id prefix ("D…") is the reliable signal there.
func (m Meta) IsDM() bool {
	if m.ChannelType == "im" {
		return true
	}
	ch := m.ChannelOf()
	return len(ch) > 0 && ch[0] == 'D'
}
