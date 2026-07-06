package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

// Client is the single typed gateway to the Slack Web API. Every command goes through it,
// so auth, retries, rate limiting, and dry-run live in exactly one place (GOAL.md §2).
type Client struct {
	auth    Authenticator
	baseURL string
	http    *http.Client
	limiter *rateLimiter
	retry   retryPolicy

	DryRun    bool
	ShowToken bool
	Verbose   bool

	dryRunW  io.Writer // where the --dry-run curl line is written (default os.Stderr)
	verboseW io.Writer

	recorder Recorder // optional observer of successful calls (local history); nil = no-op
}

// Option configures a Client.
type Option func(*Client)

// DefaultBaseURL is the public Web API root. Method URLs are <base>/<method>.
const DefaultBaseURL = "https://slack.com/api"

// New builds a Client for the given authenticator. Sensible defaults are applied; override
// with Options.
func New(auth Authenticator, opts ...Option) *Client {
	c := &Client{
		auth:    auth,
		baseURL: DefaultBaseURL,
		http:    &http.Client{Timeout: 60 * time.Second},
		// Slack rate-limits per method in tiers (Tier 2 ≈ 20/min … Tier 4 ≈ 100/min) and
		// exposes no quota headers — only 429 + Retry-After. A modest fixed pace plus
		// halve-on-429 stays inside every tier for interactive use.
		limiter:  newRateLimiter(4),
		retry:    defaultRetryPolicy(),
		dryRunW:  os.Stderr,
		verboseW: os.Stderr,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

func WithBaseURL(u string) Option {
	return func(c *Client) {
		if u != "" {
			c.baseURL = strings.TrimRight(u, "/")
		}
	}
}
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.http = h } }
func WithRPS(rps float64) Option           { return func(c *Client) { c.limiter = newRateLimiter(rps) } }
func WithRetryPolicy(p retryPolicy) Option { return func(c *Client) { c.retry = p } }
func WithDryRun(v bool) Option             { return func(c *Client) { c.DryRun = v } }
func WithShowToken(v bool) Option          { return func(c *Client) { c.ShowToken = v } }
func WithVerbose(v bool) Option            { return func(c *Client) { c.Verbose = v } }
func WithDryRunWriter(w io.Writer) Option  { return func(c *Client) { c.dryRunW = w } }

// WithRecorder attaches an observer notified after every successful, non-dry-run call — the
// local message-history hook. internal/api stays generic (it knows nothing about SQLite) by
// depending only on this narrow interface.
func WithRecorder(r Recorder) Option { return func(c *Client) { c.recorder = r } }

// BaseURL returns the configured base URL.
func (c *Client) BaseURL() string { return c.baseURL }

// Close releases resources an attached Recorder holds (e.g. the store's SQLite handle) if it
// implements io.Closer. Safe to defer unconditionally: a nil or non-Closer recorder is a
// no-op. This matters on Windows, where an open handle blocks deleting/renaming the file.
func (c *Client) Close() error {
	if closer, ok := c.recorder.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// maybeRecord offers a successful, non-dry-run result to the recorder. A nil recorder, a
// dry-run, or an empty result are all silent no-ops; the recorder owns method filtering and
// swallowing its own errors.
func (c *Client) maybeRecord(ctx context.Context, method string, params map[string]any, result json.RawMessage) {
	if c.recorder == nil || c.DryRun || len(result) == 0 {
		return
	}
	c.recorder.Record(ctx, method, params, result)
}

// envelope is the part of every Web API response shared across methods. Slack returns the
// method's payload as SIBLING fields of ok (channels, messages, members, ...), so Call hands
// back the whole raw body and commands pick their payload key.
type envelope struct {
	OK       bool   `json:"ok"`
	Error    string `json:"error"`
	Warning  string `json:"warning"`
	Needed   string `json:"needed"`   // missing_scope detail
	Provided string `json:"provided"` // missing_scope detail
	Meta     struct {
		NextCursor string   `json:"next_cursor"`
		Messages   []string `json:"messages"`
	} `json:"response_metadata"`
}

// Call invokes a Web API method and returns the raw response body (the full envelope, since
// Slack nests payloads as siblings of ok). idempotent marks read methods: they are sent as
// GET with query params — which also lets the retry policy safely replay them — while writes
// go as a form-encoded POST and are never auto-retried except on 429 (rejected, not
// processed). In dry-run mode it prints the equivalent curl and returns (nil, nil).
func (c *Client) Call(ctx context.Context, method string, params map[string]any, idempotent bool) (json.RawMessage, error) {
	raw, _, err := c.call(ctx, method, params, idempotent)
	return raw, err
}

// CallInto is Call plus a JSON decode of the response into out.
func (c *Client) CallInto(ctx context.Context, method string, params map[string]any, idempotent bool, out any) error {
	raw, err := c.Call(ctx, method, params, idempotent)
	if err != nil {
		return err
	}
	if len(raw) == 0 || out == nil {
		return nil
	}
	return json.Unmarshal(raw, out)
}

// CallAllPages follows Slack's cursor pagination (response_metadata.next_cursor; empty
// string = done) and returns the pages' resultKey arrays merged into a single JSON array.
// max caps the total items collected; max <= 0 means all. The per-request limit is set by
// the caller via params["limit"] (Slack recommends 100-200).
func (c *Client) CallAllPages(ctx context.Context, method string, params map[string]any, resultKey string, max int) (json.RawMessage, error) {
	if c.DryRun {
		// One representative request is the honest dry-run: the follow-up cursors depend on
		// live responses we will not have.
		_, _, err := c.call(ctx, method, params, true)
		return nil, err
	}
	p := make(map[string]any, len(params)+1)
	for k, v := range params {
		p[k] = v
	}
	var items []json.RawMessage
	for {
		raw, meta, err := c.call(ctx, method, p, true)
		if err != nil {
			return nil, err
		}
		page, err := extractArray(raw, resultKey)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", method, err)
		}
		items = append(items, page...)
		if max > 0 && len(items) >= max {
			items = items[:max]
			break
		}
		if meta == nil || meta.Meta.NextCursor == "" {
			break
		}
		p["cursor"] = meta.Meta.NextCursor
	}
	return marshalArray(items)
}

// extractArray pulls the named array out of a response body.
func extractArray(raw json.RawMessage, key string) ([]json.RawMessage, error) {
	var body map[string]json.RawMessage
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, err
	}
	field, ok := body[key]
	if !ok {
		return nil, fmt.Errorf("response has no %q field", key)
	}
	var items []json.RawMessage
	if err := json.Unmarshal(field, &items); err != nil {
		return nil, fmt.Errorf("%q is not an array: %w", key, err)
	}
	return items, nil
}

func marshalArray(items []json.RawMessage) (json.RawMessage, error) {
	if items == nil {
		items = []json.RawMessage{}
	}
	return json.Marshal(items)
}

// call executes one request and returns the raw body plus the decoded envelope (for
// pagination). It is the single network path: dry-run, backoff, 429 adaptation, and
// envelope parsing all live here.
func (c *Client) call(ctx context.Context, method string, params map[string]any, idempotent bool) (json.RawMessage, *envelope, error) {
	if c.DryRun {
		c.printCurl(method, params, idempotent)
		return nil, nil, nil
	}

	var lastErr error
	for attempt := 0; attempt < c.retry.maxAttempts; attempt++ {
		if err := c.limiter.wait(ctx); err != nil {
			return nil, nil, err
		}

		httpReq, err := c.newRequest(ctx, method, params, idempotent)
		if err != nil {
			return nil, nil, err
		}

		resp, err := c.http.Do(httpReq) //nolint:bodyclose // body is read+closed in parse()
		if err != nil {
			lastErr = fmt.Errorf("%s: %w", method, err)
			if retry, wait := c.retry.decide(attempt, 0, 0, true, idempotent); retry {
				if serr := sleepCtx(ctx, wait); serr != nil {
					return nil, nil, serr
				}
				continue
			}
			return nil, nil, lastErr
		}

		raw, env, apiErr := c.parse(resp, method)
		if apiErr == nil {
			c.limiter.reward()
			// Offer every successful page/call to the recorder; it filters to message-bearing
			// methods (chat.postMessage, conversations.history/replies). CallAllPages routes
			// each page through here, so a full history walk is captured page by page.
			c.maybeRecord(ctx, method, params, raw)
			return raw, env, nil
		}
		lastErr = apiErr
		status := resp.StatusCode
		if status == http.StatusTooManyRequests || apiErr.Code == "ratelimited" || apiErr.Code == "rate_limited" {
			c.limiter.penalize()
			status = http.StatusTooManyRequests // envelope-level rate limits retry like a 429
		}
		if retry, wait := c.retry.decide(attempt, status, apiErr.RetryAfter, false, idempotent); retry {
			if serr := sleepCtx(ctx, wait); serr != nil {
				return nil, nil, serr
			}
			continue
		}
		return nil, nil, lastErr
	}
	return nil, nil, lastErr
}

// newRequest builds the HTTP request. Reads are GETs with query params; writes are
// form-encoded POSTs. Non-scalar values (blocks, attachments, metadata) are serialized as
// JSON strings inside the form/query field — Slack accepts JSON-in-field for exactly these
// nested arguments, which sidesteps the per-method "does it accept a JSON body?" question
// (DECISIONS.md).
func (c *Client) newRequest(ctx context.Context, method string, params map[string]any, idempotent bool) (*http.Request, error) {
	vals := formValues(params)
	endpoint := c.baseURL + "/" + method

	var req *http.Request
	var err error
	if idempotent {
		u := endpoint
		if enc := vals.Encode(); enc != "" {
			u += "?" + enc
		}
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	} else {
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(vals.Encode()))
		if err == nil {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
		}
	}
	if err != nil {
		return nil, err
	}
	c.auth.Apply(req)
	return req, nil
}

// formValues renders params into url.Values, JSON-encoding non-scalars.
func formValues(params map[string]any) url.Values {
	vals := url.Values{}
	for k, v := range params {
		vals.Set(k, scalarString(v))
	}
	return vals
}

// parse reads the response body, decodes the Web API envelope, and turns a non-ok response
// into a typed *APIError. Slack signals failure through ok:false (usually with HTTP 200),
// so the envelope — not the status — is the source of truth; the status matters for 429s
// and non-JSON infrastructure errors.
func (c *Client) parse(resp *http.Response, method string) (json.RawMessage, *envelope, *APIError) {
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 32<<20)) // cap at 32MiB to bound memory

	if c.Verbose {
		_, _ = fmt.Fprintf(c.verboseW, "← %s %d %s\n", method, resp.StatusCode, truncate(strings.TrimSpace(string(raw)), 2000))
	}

	var env envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		// Non-JSON (e.g. an HTML 502 from a proxy): synthesize an APIError off the status.
		return nil, nil, &APIError{
			StatusCode: resp.StatusCode,
			Code:       "",
			Body:       fmt.Sprintf("non-JSON response: %s", truncate(strings.TrimSpace(string(raw)), 200)),
			Method:     method,
			RetryAfter: retryAfterHeader(resp),
		}
	}

	if env.OK {
		if env.Warning != "" && c.Verbose {
			_, _ = fmt.Fprintf(c.verboseW, "⚠ %s warning: %s\n", method, env.Warning)
		}
		return raw, &env, nil
	}

	return nil, nil, &APIError{
		StatusCode: resp.StatusCode,
		Code:       env.Error,
		Warning:    env.Warning,
		Needed:     env.Needed,
		Provided:   env.Provided,
		Body:       string(raw),
		Method:     method,
		RetryAfter: retryAfterHeader(resp),
	}
}

// printCurl writes the equivalent curl command for --dry-run. The token is redacted unless
// --show-token is set, so a dry-run is always safe to paste into a bug report.
func (c *Client) printCurl(method string, params map[string]any, idempotent bool) {
	authz := c.auth.Redacted()
	if c.ShowToken {
		authz = c.auth.Raw()
	}
	endpoint := c.baseURL + "/" + method

	var b strings.Builder
	b.WriteString("curl -sS")
	if idempotent {
		vals := formValues(params)
		u := endpoint
		if enc := vals.Encode(); enc != "" {
			u += "?" + enc
		}
		b.WriteByte(' ')
		b.WriteString(shellQuote(u))
	} else {
		b.WriteString(" -X POST ")
		b.WriteString(shellQuote(endpoint))
	}
	b.WriteString(" -H ")
	b.WriteString(shellQuote("Authorization: " + authz))
	extra := c.auth.ExtraHeaders(!c.ShowToken)
	for _, k := range sortedHeaderKeys(extra) {
		b.WriteString(" -H ")
		b.WriteString(shellQuote(k + ": " + extra[k]))
	}
	if !idempotent {
		for _, k := range sortedKeys(params) {
			b.WriteString(" --data-urlencode ")
			b.WriteString(shellQuote(k + "=" + scalarString(params[k])))
		}
	}
	_, _ = fmt.Fprintln(c.dryRunW, b.String())
}

func sortedHeaderKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func retryAfterHeader(resp *http.Response) int {
	if v := resp.Header.Get("Retry-After"); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
		}
	}
	return 0
}

// scalarString renders a param value for a form/query field: strings pass through, booleans
// and numbers use their JSON form, and non-scalars (blocks, attachments) become JSON text.
func scalarString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case nil:
		return ""
	case json.RawMessage:
		return string(t)
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

// shellQuote single-quote-escapes a string so a dry-run curl line is copy-pasteable and safe.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
