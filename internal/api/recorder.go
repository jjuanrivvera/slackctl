package api

import (
	"context"
	"encoding/json"
)

// Recorder observes successful Web API calls so a caller can persist a local message history
// (the `slackctl log` store) without internal/api depending on internal/store — the client
// stays a generic HTTP core (GOAL.md §2). Record is called once per successful, non-dry-run
// call with the same context the command passed in.
//
// Record must not block the caller on its own failures: an implementation filters to the
// methods it cares about and swallows/logs its own errors — a broken store must never fail a
// command.
type Recorder interface {
	Record(ctx context.Context, method string, params map[string]any, result json.RawMessage)
}
