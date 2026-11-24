package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/deliver"
	"github.com/checkmeup/checkmeup/internal/httpsafe"
)

// ─── uptime monitor checks ───────────────────────────────────────────────────

func checkUptimeMonitors(ctx context.Context, n Notifiers) {
	monitors, err := n.Queries.ListDueUptimeMonitors(ctx)
	if err != nil {
		n.Logger.Error("uptime worker: list due monitors", "err", err)
		return
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, checkConcurrency)
	for _, m := range monitors {
		m := m
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			defer recoverPanic(n.Logger, "uptime check", m.ID)
			checkOneUptimeMonitor(ctx, n, m)
		}()
	}
	wg.Wait()
}

func checkOneUptimeMonitor(ctx context.Context, n Notifiers, m db.UptimeMonitor) {
	client := n.HTTPClient
	if client == nil {
		client = uptimeCheckClient()
	}
	statusCode, responseTimeMs, isUp, failureReason := performHTTPCheck(m, client)
	trackCheck(ctx, n, uptimeTracker{
		q: n.Queries, m: m, statusCode: statusCode, responseTimeMs: responseTimeMs, failureReason: failureReason,
	}, isUp)
}

// uptimeTracker is an uptime monitor's side of trackCheck, bound to one check's result.
type uptimeTracker struct {
	q              *db.Queries
	m              db.UptimeMonitor
	statusCode     int
	responseTimeMs int64
	failureReason  string
}

func (t uptimeTracker) monitor() trackedMonitor {
	return trackedMonitor{
		Type: "uptime", ID: t.m.ID, OrgID: t.m.OrgID, Status: t.m.Status,
		ConsecutiveFailures: t.m.ConsecutiveFailures, AlertAfterNFailures: t.m.AlertAfterNFailures,
		AlertsEnabled: t.m.AlertsEnabled,
	}
}

func (t uptimeTracker) latestCheckFailed(ctx context.Context) (bool, error) {
	return t.q.LatestUptimeCheckFailed(ctx, t.m.ID)
}

func (t uptimeTracker) recordCheck(ctx context.Context, isUp bool) error {
	var codeParam pgtype.Int4
	if t.statusCode > 0 {
		codeParam = pgtype.Int4{Int32: int32(t.statusCode), Valid: true}
	}
	_, err := t.q.CreateUptimeCheck(ctx, db.CreateUptimeCheckParams{
		MonitorID:      t.m.ID,
		StatusCode:     codeParam,
		ResponseTimeMs: int32(t.responseTimeMs),
		IsUp:           isUp,
		FailureReason:  pgtype.Text{String: t.failureReason, Valid: t.failureReason != ""},
	})
	return err
}

func (t uptimeTracker) scheduleRecheck(ctx context.Context) error {
	return t.q.ScheduleUptimeRecheck(ctx, t.m.ID)
}

func (t uptimeTracker) recordUp(ctx context.Context) (pgtype.Timestamptz, error) {
	return t.q.RecordUptimeCheckUp(ctx, t.m.ID)
}

func (t uptimeTracker) recordFailure(ctx context.Context) (int32, error) {
	updated, err := t.q.RecordUptimeCheckFailure(ctx, t.m.ID)
	return updated.ConsecutiveFailures, err
}

func (t uptimeTracker) openIncident(ctx context.Context, alert []byte) (uuid.UUID, error) {
	inc, err := t.q.OpenUptimeIncident(ctx, db.OpenUptimeIncidentParams{MonitorID: t.m.ID, Alert: alert})
	return inc.ID, err
}

func (t uptimeTracker) countAlert(ctx context.Context, incidentID uuid.UUID) error {
	_, err := t.q.IncrementUptimeIncidentAlertCount(ctx, incidentID)
	return err
}

func (t uptimeTracker) downAlert() AlertMessage {
	return buildUptimeDownAlert(t.m, t.failureReason)
}

func (t uptimeTracker) recoveryAlert(downtime string) AlertMessage {
	return buildUptimeRecoveryAlert(t.m, downtime)
}

func buildUptimeRecoveryAlert(m db.UptimeMonitor, downtime string) AlertMessage {
	return backUpAlert("uptime", m.Name, downtime, codeLine("URL", m.Url))
}

func buildUptimeDownAlert(m db.UptimeMonitor, failureReason string) AlertMessage {
	return buildAlert(alertText{
		Event: "down", MonitorType: "uptime", Name: m.Name,
		Icon: "🔴", Headline: " is down",
		Details: []string{codeLine("URL", m.Url), "Reason: " + failureReason},
		Subject: "DOWN: " + m.Name,
		Reason:  failureReason,
		SMS:     fmt.Sprintf("%s is DOWN (%s)", m.Name, failureReason),
	})
}

// maxKeywordCheckBytes caps how much of the response body is read for a
// keyword search or JSON assertion, regardless of Content-Length (PRD keyword-monitoring).
const maxKeywordCheckBytes = 512 * 1024

// sharedUptimeClient is the *http.Client used for every real (non-test)
// uptime check, built once and reused rather than per call — a fresh
// http.Client/Transport per check loses keep-alive connection pooling and
// adds socket/FD churn that grows with monitor count. m.Url is user-supplied,
// so it dials through httpsafe.Dialer to block loopback/private/link-local/
// cloud-metadata targets (SSRF); redirects are still followed (this is a
// generic uptime checker, unlike the one-shot Slack/webhook delivery
// clients), but each hop re-dials through the same Control-equipped Dialer,
// so a redirect into a blocked range is caught too.
//
// No client-level Timeout: PRD configurable-uptime-checks makes the request timeout
// per-monitor (m.MaxResponseTimeMs), enforced via a context deadline in
// performHTTPCheck instead — a fixed client.Timeout here would silently cap
// every monitor at whatever that constant is, regardless of what the
// monitor is actually configured for. deliver.Timeout still bounds the dial
// (connection-establishment) phase specifically, via the Dialer below; the
// context deadline bounds the full request including body read.
var sharedUptimeClient = &http.Client{
	Transport: &http.Transport{
		DialContext:         httpsafe.Dialer(deliver.Timeout).DialContext,
		MaxIdleConnsPerHost: 4,
	},
}

func uptimeCheckClient() *http.Client {
	return sharedUptimeClient
}

// performHTTPCheck runs the monitor's HTTP check and evaluates all configured
// assertions in order: status code → keyword → JSON assertions. The first
// failing condition is the recorded failure reason. client is injected so
// tests can pass an unguarded client to reach a local httptest server — see
// uptimeCheckClient for the hardened client real checks use.
//
// m.MaxResponseTimeMs bounds the whole request (connect + response, via a
// context deadline) rather than being checked post-hoc after a successful
// response — a slow response now aborts the connection at the configured
// ceiling instead of completing and then being judged (PRD configurable-uptime-checks).
func performHTTPCheck(m db.UptimeMonitor, client *http.Client) (statusCode int, responseTimeMs int64, isUp bool, failureReason string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(m.MaxResponseTimeMs)*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, string(m.HttpMethod), m.Url, nil)
	if err != nil {
		return 0, 0, false, "invalid request"
	}

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		return 0, elapsed, false, "timeout / connection error"
	}
	defer func() { _ = resp.Body.Close() }()

	keyword := strings.TrimSpace(m.Keyword.String)
	var assertions []jsonAssertion
	if len(m.JsonAssertions) > 0 {
		_ = json.Unmarshal(m.JsonAssertions, &assertions)
	}

	// The body is read only for a keyword or an assertion, and then only its
	// first maxKeywordCheckBytes: draining it all would download a large file
	// on every check, until the timeout, and fill the worker's link for every
	// other check. Closing it unread costs only the connection's reuse.
	needsBody := keyword != "" || len(assertions) > 0
	var body string
	if needsBody {
		raw, err := io.ReadAll(io.LimitReader(resp.Body, maxKeywordCheckBytes))
		if err != nil {
			// Checking the part that arrived could pass a not_contains keyword
			// on a page cut off by the timeout.
			return resp.StatusCode, elapsed, false, "timeout / connection error"
		}
		body = string(raw)
	}

	if !statusCodeAccepted(resp.StatusCode, m.AcceptedStatusCodes) {
		return resp.StatusCode, elapsed, false, httpStatusDesc(resp.StatusCode)
	}

	if keyword != "" && !keywordCheckPasses(body, keyword, m.KeywordMode, m.KeywordCaseSensitive) {
		return resp.StatusCode, elapsed, false, keywordFailureReason(m.KeywordMode)
	}

	for _, a := range assertions {
		if reason, ok := evaluateJsonAssertion(body, a); !ok {
			return resp.StatusCode, elapsed, false, reason
		}
	}

	return resp.StatusCode, elapsed, true, ""
}

// statusCodeAccepted reports whether code is one of the monitor's configured
// accepted status codes.
func statusCodeAccepted(code int, accepted []int32) bool {
	for _, c := range accepted {
		if int32(code) == c {
			return true
		}
	}
	return false
}

// keywordCheckPasses reports whether body satisfies the monitor's keyword
// mode. Plain substring search only — no regex, to avoid a ReDoS surface.
func keywordCheckPasses(body, keyword string, mode db.KeywordMode, caseSensitive bool) bool {
	if !caseSensitive {
		body = strings.ToLower(body)
		keyword = strings.ToLower(keyword)
	}
	contains := strings.Contains(body, keyword)
	if mode == db.KeywordModeNotContains {
		return !contains
	}
	return contains
}

func keywordFailureReason(mode db.KeywordMode) string {
	if mode == db.KeywordModeNotContains {
		return "Keyword found"
	}
	return "Keyword not found"
}

func httpStatusDesc(code int) string {
	if code == 0 {
		return "timeout / connection error"
	}
	return fmt.Sprintf("HTTP %d", code)
}

type jsonAssertion struct {
	Path       string `json:"path"`
	Comparator string `json:"comparator"`
	Expected   string `json:"expected"`
}

// evaluateJsonAssertion resolves path in the JSON body and compares to
// expected. Returns ("", true) on pass or (failureReason, false) on fail.
func evaluateJsonAssertion(body string, a jsonAssertion) (string, bool) {
	var root map[string]any
	if err := json.Unmarshal([]byte(body), &root); err != nil {
		return "response is not valid JSON", false
	}
	actual, errMsg, ok := resolveJSONPath(root, a.Path)
	if !ok {
		return errMsg, false
	}
	return applyComparator(a.Path, a.Comparator, actual, a.Expected)
}

func resolveJSONPath(root map[string]any, path string) (value, errMsg string, ok bool) {
	p := strings.TrimPrefix(path, "$.")
	p = strings.TrimPrefix(p, ".")
	var cur any = root
	for _, seg := range strings.Split(p, ".") {
		m, isMap := cur.(map[string]any)
		if !isMap {
			return "", fmt.Sprintf("JSON path %q not found", path), false
		}
		cur, ok = m[seg]
		if !ok {
			return "", fmt.Sprintf("JSON path %q not found", path), false
		}
	}
	return jsonValueToString(cur), "", true
}

func applyComparator(path, comparator, actual, expected string) (string, bool) {
	switch comparator {
	case "greater_than", "less_than":
		return applyNumericComparator(path, comparator, actual, expected)
	default:
		return applyStringComparator(path, comparator, actual, expected)
	}
}

func applyStringComparator(path, comparator, actual, expected string) (string, bool) {
	var passes bool
	switch comparator {
	case "equals":
		passes = actual == expected
	case "not_equals":
		passes = actual != expected
	default: // "contains"
		passes = strings.Contains(actual, expected)
	}
	if !passes {
		return fmt.Sprintf("JSON assertion failed: %q %s %q (got %q)", path, comparator, expected, actual), false
	}
	return "", true
}

func applyNumericComparator(path, comparator, actual, expected string) (string, bool) {
	av, ae, parseOk := parseNumericPair(actual, expected)
	var passes bool
	if parseOk {
		if comparator == "greater_than" {
			passes = av > ae
		} else {
			passes = av < ae
		}
	}
	if !passes {
		return fmt.Sprintf("JSON assertion failed: %q %s %q (got %q)", path, comparator, expected, actual), false
	}
	return "", true
}

func jsonValueToString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	case float64:
		if t == float64(int64(t)) {
			return fmt.Sprintf("%d", int64(t))
		}
		return fmt.Sprintf("%g", t)
	case nil:
		return "null"
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func parseNumericPair(actual, expected string) (float64, float64, bool) {
	av, err1 := strconv.ParseFloat(actual, 64)
	ae, err2 := strconv.ParseFloat(expected, 64)
	return av, ae, err1 == nil && err2 == nil
}
