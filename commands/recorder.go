package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/jjuanrivvera/slackctl/internal/store"
)

// storeRecorder implements api.Recorder: it persists the messages flowing through
// message-bearing Web API calls into the local SQLite history, so `slackctl log` can search
// them offline. It is a pure observer — a store error is warned once and swallowed, never
// failing the command that triggered it (a broken history must not break a post).
type storeRecorder struct {
	st        *store.Store
	workspace string
	quiet     bool
	errW      io.Writer
	warned    bool
}

// Record captures the messages from the methods worth persisting: outbound posts and the
// history/replies you fetch. Everything else is ignored.
func (r *storeRecorder) Record(ctx context.Context, method string, params map[string]any, result json.RawMessage) {
	var msgs []store.Message
	switch method {
	case "chat.postMessage", "chat.update":
		msgs = r.fromPostResult(result)
	case "conversations.history", "conversations.replies":
		msgs = r.fromMessageList(paramString(params, "channel"), result)
	default:
		return
	}
	if len(msgs) == 0 {
		return
	}
	if _, err := r.st.RecordBatch(ctx, msgs); err != nil {
		r.warn(err)
	}
}

// Close releases the underlying store handle (api.Client.Close defers this via io.Closer).
func (r *storeRecorder) Close() error { return r.st.Close() }

// fromPostResult builds a message from a chat.postMessage/update result:
// {ok, channel, ts, message:{...}}.
func (r *storeRecorder) fromPostResult(result json.RawMessage) []store.Message {
	var res struct {
		Channel string          `json:"channel"`
		TS      string          `json:"ts"`
		Message json.RawMessage `json:"message"`
	}
	if json.Unmarshal(result, &res) != nil || res.Channel == "" || res.TS == "" {
		return nil
	}
	m := r.parseMessage(res.Message)
	m.Channel = res.Channel
	m.TS = res.TS
	m.Workspace = r.workspace
	m.Raw = res.Message
	return []store.Message{m}
}

// fromMessageList builds messages from a conversations.history/replies result: {ok, messages:[…]}.
func (r *storeRecorder) fromMessageList(channel string, result json.RawMessage) []store.Message {
	if channel == "" {
		return nil
	}
	var res struct {
		Messages []json.RawMessage `json:"messages"`
	}
	if json.Unmarshal(result, &res) != nil {
		return nil
	}
	out := make([]store.Message, 0, len(res.Messages))
	for _, raw := range res.Messages {
		m := r.parseMessage(raw)
		if m.TS == "" {
			continue
		}
		m.Channel = channel
		m.Workspace = r.workspace
		m.Raw = raw
		out = append(out, m)
	}
	return out
}

// parseMessage extracts the common fields from a Slack message object.
func (r *storeRecorder) parseMessage(raw json.RawMessage) store.Message {
	var m struct {
		TS       string `json:"ts"`
		ThreadTS string `json:"thread_ts"`
		User     string `json:"user"`
		Type     string `json:"type"`
		Subtype  string `json:"subtype"`
		Text     string `json:"text"`
	}
	_ = json.Unmarshal(raw, &m)
	return store.Message{
		TS: m.TS, ThreadTS: m.ThreadTS, User: m.User,
		Type: m.Type, Subtype: m.Subtype, Text: m.Text,
	}
}

func (r *storeRecorder) warn(err error) {
	if r.quiet || r.warned {
		return
	}
	r.warned = true // warn once per invocation, not once per message
	fmt.Fprintf(r.errW, "slackctl: warning: local history unavailable (%v) — continuing without it\n", err)
}

// recordEvent persists a single streamed listen event (already the bare event object) into the
// store. Used by `listen` since events arrive over the websocket, not through api.Client.
func recordEvent(ctx context.Context, st *store.Store, workspace, channel string, event json.RawMessage) {
	var m struct {
		Type     string `json:"type"`
		Subtype  string `json:"subtype"`
		TS       string `json:"ts"`
		ThreadTS string `json:"thread_ts"`
		User     string `json:"user"`
		Text     string `json:"text"`
		Channel  string `json:"channel"`
	}
	if json.Unmarshal(event, &m) != nil || m.Type != "message" || m.TS == "" {
		return // only persist actual messages, not typing/presence/reaction frames
	}
	ch := m.Channel
	if ch == "" {
		ch = channel
	}
	_ = st.Record(ctx, store.Message{
		Workspace: workspace, Channel: ch, TS: m.TS, ThreadTS: m.ThreadTS,
		User: m.User, Type: m.Type, Subtype: m.Subtype, Text: m.Text, Raw: event,
	})
}

// paramString reads a string param defensively (params values are `any`).
func paramString(params map[string]any, key string) string {
	if v, ok := params[key].(string); ok {
		return v
	}
	return ""
}
