package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

// msgTemplateCmd is a beyond-the-API value-add: render a Go text/template file with
// --set key=value variables and post the result. Handy for parameterized alerts/reports.
// With --blocks, the rendered output is parsed as a Block Kit JSON array.
func msgTemplateCmd() *cobra.Command {
	var channel, file, threadTS string
	var vars []string
	var asBlocks bool
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Render a template file and post it",
		Long: `Render a Go text/template file, substituting --set key=value variables, and post the
result to a conversation. Reference a variable as {{.name}}. With --blocks, the rendered
text is sent as a Block Kit JSON array instead of plain text.`,
		Example: `  slackctl msg template --channel C0123456 --file alert.tmpl --set service=api --set status=down
  slackctl msg template --channel C0123456 --file card.json.tmpl --set title=Deploy --blocks`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			data := map[string]string{}
			for _, kv := range vars {
				k, v, ok := strings.Cut(kv, "=")
				if !ok {
					return fmt.Errorf("invalid --set %q (want key=value)", kv)
				}
				data[k] = v
			}

			body, err := os.ReadFile(file) //nolint:gosec // G304: file is the user's chosen template path
			if err != nil {
				return fmt.Errorf("read template: %w", err)
			}
			tmpl, err := template.New("msg").Option("missingkey=error").Parse(string(body))
			if err != nil {
				return fmt.Errorf("parse template: %w", err)
			}
			var rendered bytes.Buffer
			if err := tmpl.Execute(&rendered, data); err != nil {
				return fmt.Errorf("render template: %w", err)
			}

			params := map[string]any{"channel": channel}
			if threadTS != "" {
				params["thread_ts"] = threadTS
			}
			if asBlocks {
				var blocks any
				if err := json.Unmarshal(rendered.Bytes(), &blocks); err != nil {
					return fmt.Errorf("--blocks: rendered template is not valid JSON: %w", err)
				}
				params["blocks"] = blocks
			} else {
				params["text"] = rendered.String()
			}

			client, err := clientFromCmd(cmd)
			if err != nil {
				return err
			}
			defer func() { _ = client.Close() }()
			raw, err := client.Call(cmd.Context(), "chat.postMessage", params, false)
			if err != nil {
				return err
			}
			if raw == nil { // dry-run
				return nil
			}
			if !cmd.Flags().Changed("columns") {
				_ = cmd.Flags().Set("columns", "channel,ts")
			}
			return render(cmd, raw)
		},
	}
	cmd.Flags().StringVar(&channel, "channel", "", "conversation id to post to")
	cmd.Flags().StringVar(&file, "file", "", "path to the template file")
	cmd.Flags().StringArrayVar(&vars, "set", nil, "template variable key=value (repeatable)")
	cmd.Flags().StringVar(&threadTS, "thread-ts", "", "post as a reply in this thread")
	cmd.Flags().BoolVar(&asBlocks, "blocks", false, "send the rendered output as a Block Kit JSON array")
	_ = cmd.MarkFlagRequired("channel")
	_ = cmd.MarkFlagRequired("file")
	markKind(cmd, kindWrite)
	return cmd
}
