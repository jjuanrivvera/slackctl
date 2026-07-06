package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/jjuanrivvera/slackctl/internal/api"
)

// The files family. list/info/delete map 1:1 to files.* methods; upload and download are
// hand-written composites (files.upload was sunset, so upload runs the external-upload flow,
// and download fetches the private URL with the credential).

func init() {
	registerGroup(group{
		Use:     "files",
		Aliases: []string{"file"},
		Short:   "Upload, download, and manage files",
		Cmds: []methodCmd{
			{
				Use: "list", Method: "files.list", Kind: kindRead,
				Short:     "List files",
				Example:   "  slackctl files list --channel C0123456\n  slackctl files list --user U0123456 --types images",
				ResultKey: "files",
				Columns:   []string{"id", "name", "filetype", "size", "user"},
				Flags: []flagSpec{
					{Name: "channel", Kind: flagString, Usage: "only files in this conversation"},
					{Name: "user", Kind: flagString, Usage: "only files from this user"},
					{Name: "types", Kind: flagString, Usage: "filter by type: all,spaces,snippets,images,gdocs,zips,pdfs"},
					{Name: "ts-from", Kind: flagString, Usage: "only files created after this timestamp"},
					{Name: "ts-to", Kind: flagString, Usage: "only files created before this timestamp"},
				},
			},
			{
				Use: "info", Method: "files.info", Kind: kindRead,
				Short:     "Show a file's metadata",
				Example:   "  slackctl files info --file F0123456 -o json",
				ResultKey: "file",
				Columns:   []string{"id", "name", "filetype", "size", "url_private"},
				Flags: []flagSpec{
					{Name: "file", Kind: flagString, Required: true, Usage: "file id (F…)"},
				},
			},
			{
				Use: "delete", Method: "files.delete", Kind: kindDestructive,
				Short:   "Delete a file",
				Example: "  slackctl files delete --file F0123456",
				Flags: []flagSpec{
					{Name: "file", Kind: flagString, Required: true, Usage: "file id to delete"},
				},
			},
		},
		Extra: []func() *cobra.Command{filesUploadCmd, filesDownloadCmd},
	})
}

// filesUploadCmd runs the external-upload flow (getUploadURLExternal → POST → complete).
func filesUploadCmd() *cobra.Command {
	var file, channels, title, comment, threadTS, snippetType string
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload a file and share it to conversations",
		Long: `Upload a local file via Slack's external-upload flow (files.upload was sunset in
November 2025) and optionally share it into one or more conversations.`,
		Example: `  slackctl files upload --file report.pdf --channels C0123456
  slackctl files upload --file diagram.png --channels C0123456,C0456789 --comment "v2"
  slackctl files upload --file snippet.py --channels C0123456 --snippet-type python`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromCmd(cmd)
			if err != nil {
				return err
			}
			res, err := client.UploadFile(cmd.Context(), file, api.UploadOptions{
				Title:          title,
				Channels:       channels,
				InitialComment: comment,
				ThreadTS:       threadTS,
				SnippetType:    snippetType,
			})
			if err != nil {
				return err
			}
			if res == nil || res.Raw == nil { // dry-run
				return nil
			}
			if !cmd.Flags().Changed("columns") {
				_ = cmd.Flags().Set("columns", "id,name,filetype,size")
			}
			return render(cmd, extractFileList(res.Raw))
		},
	}
	cmd.Flags().StringVar(&file, "file", "", "path to the local file to upload")
	cmd.Flags().StringVar(&channels, "channels", "", "comma-separated conversation ids to share into")
	cmd.Flags().StringVar(&title, "title", "", "file title (defaults to the filename)")
	cmd.Flags().StringVar(&comment, "comment", "", "initial comment posted with the file")
	cmd.Flags().StringVar(&threadTS, "thread-ts", "", "share into this thread")
	cmd.Flags().StringVar(&snippetType, "snippet-type", "", "for code snippets: text, python, go, …")
	_ = cmd.MarkFlagRequired("file")
	markKind(cmd, kindWrite)
	return cmd
}

// extractFileList pulls the files array from a completeUploadExternal response for rendering.
func extractFileList(raw json.RawMessage) json.RawMessage {
	var body struct {
		Files json.RawMessage `json:"files"`
	}
	if json.Unmarshal(raw, &body) == nil && len(body.Files) > 0 {
		return body.Files
	}
	return raw
}

// filesDownloadCmd fetches a file's private content with the credential and writes it out.
func filesDownloadCmd() *cobra.Command {
	var fileID, out string
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download a file's contents",
		Long:  "Resolve a file's private URL and stream its bytes to --out (or ./<name> by default).",
		Example: `  slackctl files download --file F0123456
  slackctl files download --file F0123456 --out ~/Downloads/report.pdf`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromCmd(cmd)
			if err != nil {
				return err
			}
			var info struct {
				File struct {
					Name         string `json:"name"`
					URLPrivate   string `json:"url_private_download"`
					URLPrivateUp string `json:"url_private"`
				} `json:"file"`
			}
			if err := client.CallInto(cmd.Context(), "files.info", map[string]any{"file": fileID}, true, &info); err != nil {
				return err
			}
			url := info.File.URLPrivate
			if url == "" {
				url = info.File.URLPrivateUp
			}
			if url == "" && !client.DryRun {
				return fmt.Errorf("file %s has no downloadable URL (it may be an external or hidden file)", fileID)
			}

			dest := out
			if dest == "" {
				dest = info.File.Name
			}
			if dest == "" {
				dest = fileID
			}
			// Confine the destination to a real path and avoid clobbering a directory.
			if fi, err := os.Stat(dest); err == nil && fi.IsDir() {
				dest = filepath.Join(dest, fileID)
			}
			f, err := os.Create(dest) //nolint:gosec // G304: dest is the user's chosen --output path
			if err != nil {
				return err
			}
			defer func() { _ = f.Close() }()
			n, err := client.FetchAuthed(cmd.Context(), url, f)
			if err != nil {
				return err
			}
			if client.DryRun {
				return nil
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "wrote %d bytes to %s\n", n, dest)
			return nil
		},
	}
	cmd.Flags().StringVar(&fileID, "file", "", "file id to download (F…)")
	cmd.Flags().StringVar(&out, "out", "", "destination path (default: the file's name in the cwd)")
	_ = cmd.MarkFlagRequired("file")
	markKind(cmd, kindRead)
	return cmd
}
