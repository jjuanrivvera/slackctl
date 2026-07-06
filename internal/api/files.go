package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// UploadResult is the outcome of a completed external file upload.
type UploadResult struct {
	FileID string `json:"file_id"`
	Raw    json.RawMessage
}

// UploadOptions carries the sharing parameters for a file upload.
type UploadOptions struct {
	Filename       string // defaults to the base name of the path
	Title          string
	Channels       string // comma-separated channel ids to share into
	InitialComment string
	ThreadTS       string
	SnippetType    string // for code snippets (e.g. "text", "python")
}

// UploadFile runs Slack's three-step external upload flow (files.upload was sunset Nov 2025):
//  1. files.getUploadURLExternal — reserve an upload URL + file id for (filename, length);
//  2. POST the file bytes to that URL (multipart, field "file");
//  3. files.completeUploadExternal — finalize and share the file.
//
// It returns the completed file id. path must be a readable regular file.
func (c *Client) UploadFile(ctx context.Context, path string, opts UploadOptions) (*UploadResult, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("open upload file: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("%s is a directory, not a file", path)
	}
	filename := opts.Filename
	if filename == "" {
		filename = filepath.Base(path)
	}

	// Step 1: reserve the upload URL.
	getParams := map[string]any{"filename": filename, "length": info.Size()}
	if opts.SnippetType != "" {
		getParams["snippet_type"] = opts.SnippetType
	}
	var reserve struct {
		UploadURL string `json:"upload_url"`
		FileID    string `json:"file_id"`
	}
	if err := c.CallInto(ctx, "files.getUploadURLExternal", getParams, true, &reserve); err != nil {
		return nil, err
	}
	if c.DryRun {
		// Step 1 made no request (it printed a curl); print the follow-up steps' curls too,
		// against the reserved-URL placeholder, so the whole flow is visible without a request.
		_ = c.uploadBytes(ctx, "<upload_url from files.getUploadURLExternal>", path, filename)
		_, _ = c.Call(ctx, "files.completeUploadExternal", map[string]any{"files": "<file_id>"}, false)
		return nil, nil
	}
	if reserve.UploadURL == "" || reserve.FileID == "" {
		return nil, fmt.Errorf("files.getUploadURLExternal returned no upload_url/file_id")
	}

	// Step 2: POST the bytes to the reserved URL. This is a plain upload endpoint, not the
	// Web API base, and it is NOT authenticated with the token (the URL carries the ticket).
	if err := c.uploadBytes(ctx, reserve.UploadURL, path, filename); err != nil {
		return nil, err
	}

	// Step 3: complete + share. files carries a JSON array of {id,title}.
	files := []map[string]string{{"id": reserve.FileID}}
	if opts.Title != "" {
		files[0]["title"] = opts.Title
	}
	completeParams := map[string]any{"files": files}
	if opts.Channels != "" {
		completeParams["channel_id"] = opts.Channels
	}
	if opts.InitialComment != "" {
		completeParams["initial_comment"] = opts.InitialComment
	}
	if opts.ThreadTS != "" {
		completeParams["thread_ts"] = opts.ThreadTS
	}
	raw, err := c.Call(ctx, "files.completeUploadExternal", completeParams, false)
	if err != nil {
		return nil, err
	}
	return &UploadResult{FileID: reserve.FileID, Raw: raw}, nil
}

// uploadBytes streams path to the reserved upload URL as multipart/form-data (field "file").
func (c *Client) uploadBytes(ctx context.Context, uploadURL, path, filename string) error {
	if c.DryRun {
		_, _ = fmt.Fprintf(c.dryRunW, "curl -sS -X POST %s -F file=@%s\n", shellQuote(uploadURL), shellQuote(path))
		return nil
	}
	f, err := os.Open(path) //nolint:gosec // G304: path is a user-supplied local file to upload
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		return err
	}
	if _, err := io.Copy(fw, f); err != nil {
		return err
	}
	if err := mw.Close(); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("upload POST: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<10))
		return &APIError{StatusCode: resp.StatusCode, Code: "upload_failed", Method: "files.upload",
			Body: fmt.Sprintf("upload POST returned %d: %s", resp.StatusCode, truncate(string(body), 200))}
	}
	return nil
}

// FetchAuthed streams a Slack-hosted private URL (url_private / url_private_download) to w,
// applying the credential. Slack file URLs require the same Authorization header as the API.
func (c *Client) FetchAuthed(ctx context.Context, url string, w io.Writer) (int64, error) {
	if c.DryRun {
		authz := c.auth.Redacted()
		if c.ShowToken {
			authz = c.auth.Raw()
		}
		_, _ = fmt.Fprintf(c.dryRunW, "curl -sS %s -H %s\n", shellQuote(url), shellQuote("Authorization: "+authz))
		return 0, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	c.auth.Apply(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, fmt.Errorf("download %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<10))
		return 0, &APIError{StatusCode: resp.StatusCode, Code: "download_failed", Method: "files.download",
			Body: fmt.Sprintf("download returned %d: %s", resp.StatusCode, truncate(string(body), 200))}
	}
	return io.Copy(w, resp.Body)
}
