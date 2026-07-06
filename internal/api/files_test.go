package api

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadFile_ThreeStepFlow(t *testing.T) {
	var got []byte
	upSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseMultipartForm(1<<20))
		f, _, _ := r.FormFile("file")
		got, _ = io.ReadAll(f)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upSrv.Close)

	var completed bool
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		m := r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]
		switch m {
		case "files.getUploadURLExternal":
			_, _ = w.Write([]byte(`{"ok":true,"upload_url":"` + upSrv.URL + `","file_id":"F1"}`))
		case "files.completeUploadExternal":
			completed = true
			_, _ = w.Write([]byte(`{"ok":true,"files":[{"id":"F1"}]}`))
		default:
			t.Fatalf("unexpected %s", m)
		}
	})

	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	require.NoError(t, os.WriteFile(path, []byte("hello"), 0o600))

	res, err := c.UploadFile(t.Context(), path, UploadOptions{Channels: "C1", Title: "T"})
	require.NoError(t, err)
	assert.Equal(t, "F1", res.FileID)
	assert.Equal(t, "hello", string(got))
	assert.True(t, completed)
}

func TestUploadFile_RejectsDirAndMissing(t *testing.T) {
	c := newTestClient(t, nil)
	_, err := c.UploadFile(t.Context(), t.TempDir(), UploadOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "directory")

	_, err = c.UploadFile(t.Context(), filepath.Join(t.TempDir(), "nope"), UploadOptions{})
	require.Error(t, err)
}

func TestFetchAuthed(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte("bytes"))
	}))
	t.Cleanup(srv.Close)
	c := newTestClient(t, nil)
	var buf bytes.Buffer
	n, err := c.FetchAuthed(t.Context(), srv.URL, &buf)
	require.NoError(t, err)
	assert.Equal(t, int64(5), n)
	assert.Equal(t, "bytes", buf.String())
	assert.Equal(t, "Bearer xoxb-test-token", gotAuth)
}

func TestFetchAuthed_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(srv.Close)
	c := newTestClient(t, nil)
	_, err := c.FetchAuthed(t.Context(), srv.URL, &bytes.Buffer{})
	require.Error(t, err)
}
