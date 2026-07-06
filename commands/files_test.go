package commands

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilesList(t *testing.T) {
	srv := newServer(t, routes{
		"files.list": `{"ok":true,"files":[{"id":"F1","name":"a.pdf","filetype":"pdf","size":1024,"user":"U1"}]}`,
	})
	out, _, err := run(t, srv, "files", "list", "--channel", "C1")
	require.NoError(t, err)
	mustContain(t, out, "a.pdf")
	assert.NotContains(t, out, `"ok"`)
}

func TestFilesInfo(t *testing.T) {
	srv := newServer(t, routes{
		"files.info": `{"ok":true,"file":{"id":"F1","name":"a.pdf","filetype":"pdf"}}`,
	})
	out, _, err := run(t, srv, "files", "info", "--file", "F1", "-o", "json")
	require.NoError(t, err)
	mustContain(t, out, "a.pdf")
}

func TestFilesDelete(t *testing.T) {
	srv := newServer(t, routes{"files.delete": `{"ok":true}`})
	out, _, err := run(t, srv, "files", "delete", "--file", "F1", "-o", "json")
	require.NoError(t, err)
	mustContain(t, out, "true")
}

// TestFilesUpload_ExternalFlow exercises the full three-step upload against mocks: reserve
// URL → POST bytes → complete.
func TestFilesUpload_ExternalFlow(t *testing.T) {
	var postedBytes []byte
	// The upload endpoint is a separate server (Slack returns its URL from getUploadURLExternal).
	uploadSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.NoError(t, r.ParseMultipartForm(1<<20))
		f, _, err := r.FormFile("file")
		require.NoError(t, err)
		postedBytes, _ = io.ReadAll(f)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK - 1"))
	}))
	t.Cleanup(uploadSrv.Close)

	var completeParams url.Values
	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
		w.Header().Set("Content-Type", "application/json")
		switch method {
		case "files.getUploadURLExternal":
			assert.Equal(t, "hello.txt", r.URL.Query().Get("filename"))
			assert.Equal(t, "5", r.URL.Query().Get("length"))
			_, _ = w.Write([]byte(`{"ok":true,"upload_url":"` + uploadSrv.URL + `/up","file_id":"F99"}`))
		case "files.completeUploadExternal":
			_ = r.ParseForm()
			completeParams = r.PostForm
			_, _ = w.Write([]byte(`{"ok":true,"files":[{"id":"F99","name":"hello.txt","filetype":"text"}]}`))
		default:
			t.Fatalf("unexpected method %s", method)
		}
	}))
	t.Cleanup(apiSrv.Close)

	dir := t.TempDir()
	path := filepath.Join(dir, "hello.txt")
	require.NoError(t, os.WriteFile(path, []byte("hello"), 0o600))

	out, _, err := run(t, apiSrv, "files", "upload", "--file", path, "--channels", "C1,C2", "--comment", "hi", "-o", "json")
	require.NoError(t, err)
	assert.Equal(t, "hello", string(postedBytes), "the file bytes must reach the upload URL")
	// completeUploadExternal carries the file id, channels, and comment.
	require.NotNil(t, completeParams)
	assert.Contains(t, completeParams.Get("files"), "F99")
	assert.Equal(t, "C1,C2", completeParams.Get("channel_id"))
	assert.Equal(t, "hi", completeParams.Get("initial_comment"))
	mustContain(t, out, "hello.txt")
}

func TestFilesUpload_MissingFile(t *testing.T) {
	_, _, err := run(t, nil, "files", "upload", "--file", "/no/such/file", "--channels", "C1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "open upload file")
}

// TestFilesDownload fetches a file's private URL with the credential and writes it out.
func TestFilesDownload(t *testing.T) {
	// The file content is served from a "private URL" that requires the auth header.
	var gotAuth string
	contentSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte("PDF-BYTES"))
	}))
	t.Cleanup(contentSrv.Close)

	apiSrv := newServer(t, routes{
		"files.info": `{"ok":true,"file":{"id":"F1","name":"report.pdf","url_private_download":"` + contentSrv.URL + `/dl"}}`,
	})

	dir := t.TempDir()
	dest := filepath.Join(dir, "out.pdf")
	_, errb, err := run(t, apiSrv, "files", "download", "--file", "F1", "--out", dest)
	require.NoError(t, err, errb)
	assert.Equal(t, "Bearer xoxb-test-token", gotAuth, "the download must carry the credential")
	data, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, "PDF-BYTES", string(data))
	mustContain(t, errb, "wrote 9 bytes")
}

func TestFilesUpload_DryRun(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "x.txt")
	require.NoError(t, os.WriteFile(path, []byte("hi"), 0o600))
	// Dry-run must not require a reachable server; it prints curls and makes no request.
	_, errb, err := run(t, nil, "files", "upload", "--file", path, "--channels", "C1", "--dry-run")
	require.NoError(t, err)
	assert.Contains(t, errb, "files.getUploadURLExternal")
}
