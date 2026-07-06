package commands

// The canvases family — Slack Canvases (collaborative docs). Canvas edits are structured
// operations, so --content/--changes/--criteria take JSON (Slack's own edit-op shapes).

func init() {
	registerGroup(group{
		Use:     "canvases",
		Aliases: []string{"canvas"},
		Short:   "Create and manage Canvases",
		Cmds: []methodCmd{
			{
				Use: "create", Method: "canvases.create", Kind: kindWrite,
				Short:   "Create a canvas",
				Example: `  slackctl canvases create --title "Runbook" --content '{"type":"markdown","markdown":"# Runbook"}'`,
				Columns: []string{"canvas_id"},
				Flags: []flagSpec{
					{Name: "title", Kind: flagString, Usage: "canvas title"},
					{Name: "channel", Param: "channel_id", Kind: flagString, Usage: "create a channel canvas in this conversation"},
					{Name: "content", Param: "document_content", Kind: flagJSON, Usage: `document content JSON, e.g. {"type":"markdown","markdown":"…"}`},
				},
			},
			{
				Use: "edit", Method: "canvases.edit", Kind: kindWrite,
				Short:   "Apply edit operations to a canvas",
				Example: `  slackctl canvases edit --canvas F0123456 --changes '[{"operation":"insert_at_end","document_content":{"type":"markdown","markdown":"more"}}]'`,
				Flags: []flagSpec{
					{Name: "canvas", Param: "canvas_id", Kind: flagString, Required: true, Usage: "canvas id"},
					{Name: "changes", Kind: flagJSON, Required: true, Usage: "JSON array of edit operations"},
				},
			},
			{
				Use: "delete", Method: "canvases.delete", Kind: kindDestructive,
				Short:   "Delete a canvas",
				Example: "  slackctl canvases delete --canvas F0123456",
				Flags: []flagSpec{
					{Name: "canvas", Param: "canvas_id", Kind: flagString, Required: true, Usage: "canvas id"},
				},
			},
			{
				Use: "access-set", Method: "canvases.access.set", Kind: kindWrite,
				Short:   "Set who can read/edit a canvas",
				Example: "  slackctl canvases access-set --canvas F0123456 --access-level write --channels C0123456",
				Flags: []flagSpec{
					{Name: "canvas", Param: "canvas_id", Kind: flagString, Required: true, Usage: "canvas id"},
					{Name: "access-level", Kind: flagString, Required: true, Usage: "read|write"},
					{Name: "channels", Param: "channel_ids", Kind: flagStringSlice, Usage: "channel ids to grant access"},
					{Name: "users", Param: "user_ids", Kind: flagStringSlice, Usage: "user ids to grant access"},
				},
			},
			{
				Use: "access-delete", Method: "canvases.access.delete", Kind: kindDestructive,
				Short:   "Revoke access to a canvas",
				Example: "  slackctl canvases access-delete --canvas F0123456 --users U0123456",
				Flags: []flagSpec{
					{Name: "canvas", Param: "canvas_id", Kind: flagString, Required: true, Usage: "canvas id"},
					{Name: "channels", Param: "channel_ids", Kind: flagStringSlice, Usage: "channel ids to revoke"},
					{Name: "users", Param: "user_ids", Kind: flagStringSlice, Usage: "user ids to revoke"},
				},
			},
			{
				Use: "sections-lookup", Method: "canvases.sections.lookup", Kind: kindRead,
				Short:     "Find sections in a canvas by criteria",
				Example:   `  slackctl canvases sections-lookup --canvas F0123456 --criteria '{"contains_text":"TODO"}'`,
				ResultKey: "sections",
				Flags: []flagSpec{
					{Name: "canvas", Param: "canvas_id", Kind: flagString, Required: true, Usage: "canvas id"},
					{Name: "criteria", Kind: flagJSON, Required: true, Usage: "lookup criteria JSON"},
				},
			},
		},
	})
}
