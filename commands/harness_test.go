package commands

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

// routes maps a Web API method name to its raw response body. Slack nests payloads as
// siblings of ok, so route values are FULL envelopes (`{"ok":true,"channels":[...]}`), not
// a result wrapper. An unknown method answers ok:false so tests exercise the error path.
type routes map[string]string

func newServer(t *testing.T, r routes) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		method := req.URL.Path[strings.LastIndex(req.URL.Path, "/")+1:]
		w.Header().Set("Content-Type", "application/json")
		if body, ok := r[method]; ok {
			_, _ = w.Write([]byte(body))
			return
		}
		_, _ = w.Write([]byte(`{"ok":false,"error":"unknown_method","warning":"` + method + `"}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// run executes slackctl with args against the given server, isolating config (a temp XDG
// dir) and the keyring (in-memory). A bot token is provided via env so commands
// authenticate; SLACK_USER_TOKEN/SLACK_APP_TOKEN cover the user/app-token paths.
func run(t *testing.T, srv *httptest.Server, args ...string) (string, string, error) {
	t.Helper()
	keyring.MockInit()
	t.Setenv("SLACK_BOT_TOKEN", "xoxb-test-token")
	t.Setenv("SLACK_USER_TOKEN", "xoxp-test-token")
	t.Setenv("SLACK_APP_TOKEN", "xapp-1-test-token")
	t.Setenv("SLACKCTL_TOKEN", "")
	t.Setenv("SLACK_XOXC_TOKEN", "")
	t.Setenv("SLACK_XOXD_TOKEN", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("NO_COLOR", "1")

	root := NewRootCmd()
	var out, errb bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errb)
	full := append([]string{}, args...)
	if srv != nil {
		full = append(full, "--base-url", srv.URL)
	}
	root.SetArgs(full)
	err := root.ExecuteContext(t.Context())
	return out.String(), errb.String(), err
}

// runNoToken is like run but with no tokens in the environment (auth/login tests own
// their credential setup) and optional stdin for prompts.
func runNoToken(t *testing.T, srv *httptest.Server, stdin string, args ...string) (string, string, error) {
	t.Helper()
	keyring.MockInit()
	for _, v := range []string{
		"SLACK_BOT_TOKEN", "SLACK_USER_TOKEN", "SLACK_APP_TOKEN", "SLACKCTL_TOKEN",
		"SLACK_XOXC_TOKEN", "SLACK_XOXD_TOKEN",
	} {
		t.Setenv(v, "")
	}
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("NO_COLOR", "1")

	root := NewRootCmd()
	var out, errb bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errb)
	root.SetIn(strings.NewReader(stdin))
	full := append([]string{}, args...)
	if srv != nil {
		full = append(full, "--base-url", srv.URL)
	}
	root.SetArgs(full)
	err := root.ExecuteContext(t.Context())
	return out.String(), errb.String(), err
}

// runIn executes slackctl against a SHARED config dir (so multiple calls see each other's
// writes) without resetting the keyring — the caller controls keyring setup.
func runIn(t *testing.T, dir string, srv *httptest.Server, token string, args ...string) (string, string, error) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("NO_COLOR", "1")
	t.Setenv("SLACKCTL_TOKEN", token)
	for _, v := range []string{"SLACK_BOT_TOKEN", "SLACK_USER_TOKEN", "SLACK_APP_TOKEN", "SLACK_XOXC_TOKEN", "SLACK_XOXD_TOKEN"} {
		t.Setenv(v, "")
	}

	root := NewRootCmd()
	var out, errb bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errb)
	root.SetIn(strings.NewReader(""))
	full := append([]string{}, args...)
	if srv != nil {
		full = append(full, "--base-url", srv.URL)
	}
	root.SetArgs(full)
	err := root.ExecuteContext(t.Context())
	return out.String(), errb.String(), err
}

func mustContain(t *testing.T, s, sub string) {
	t.Helper()
	require.Contains(t, s, sub)
}
