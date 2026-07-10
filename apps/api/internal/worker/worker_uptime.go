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

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/httpsafe"
	"github.com/checkmeup/checkmeup/internal/slack"
	"github.com/checkmeup/checkmeup/internal/webhook"
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
	if !recordUptimeCheck(ctx, n, m, statusCode, responseTimeMs, isUp, failureReason) {
		return
	}

	prevStatus := m.Status
	if isUp {
		handleUptimeUp(ctx, n, m, prevStatus)
		return
	}
	handleUptimeDown(ctx, n, m, prevStatus, failureReason)
}

func recordUptimeCheck(ctx context.Context, n Notifiers, m db.UptimeMonitor, statusCode int, responseTimeMs int64, isUp bool, failureReason string) bool {
	var codeParam pgtype.Int4
	if statusCode > 0 {
		codeParam = pgtype.Int4{Int32: int32(statusCode), Valid: true}
	}
	if _, err := n.Queries.CreateUptimeCheck(ctx, db.CreateUptimeCheckParams{
		MonitorID:      m.ID,
		StatusCode:     codeParam,
		ResponseTimeMs: int32(responseTimeMs),
		IsUp:           isUp,
		FailureReason:  pgtype.Text{String: failureReason, Valid: failureReason != ""},
	}); err != nil {
		n.Logger.Error("uptime worker: create check", "monitor_id", m.ID, "err", err)
		return false
	}
	return true
}

func handleUptimeUp(ctx context.Context, n Notifiers, m db.UptimeMonitor, prevStatus db.MonitorStatus) {
	if _, err := n.Queries.RecordUptimeCheckUp(ctx, m.ID); err != nil {
		n.Logger.Error("uptime worker: record up", "monitor_id", m.ID, "err", err)
		return
	}
	if prevStatus != db.MonitorStatusDown {
		return
	}
	inc, err := n.Queries.ResolveLatestUptimeIncident(ctx, m.ID)
	if err != nil {
		n.Logger.Error("uptime worker: resolve incident", "monitor_id", m.ID, "err", err)
		return
	}
	if !m.AlertsEnabled {
		return
	}
	downtime := FormatDuration(time.Since(inc.StartedAt.Time))
	DispatchAlert(ctx, n, m.OrgID, MonitorRef{Type: "uptime", ID: m.ID}, buildUptimeRecoveryAlert(m, downtime))
}

func buildUptimeRecoveryAlert(m db.UptimeMonitor, downtime string) AlertMessage {
	return AlertMessage{
		Telegram:     fmt.Sprintf("✅ <b>%s</b> is back up\n\nURL: <code>%s</code>", m.Name, m.Url),
		EmailSubject: fmt.Sprintf("%s recovered", m.Name),
		EmailHTML:    fmt.Sprintf("<p>✅ <b>%s</b> is back up</p><p>URL: <code>%s</code></p>", m.Name, m.Url),
		Webhook: &webhook.Event{
			EventType:        "recovery",
			MonitorName:      m.Name,
			MonitorType:      "uptime",
			DowntimeDuration: downtime,
			Timestamp:        time.Now().UTC().Format(time.RFC3339),
		},
		Slack: slackMsg(slack.RecoveryMessage(m.Name, "uptime", downtime)),
		SMS:   TruncateSMS(fmt.Sprintf("Checkmeup: %s recovered after %s downtime", m.Name, downtime)),
	}
}

func handleUptimeDown(ctx context.Context, n Notifiers, m db.UptimeMonitor, prevStatus db.MonitorStatus, failureReason string) {
	updated, err := n.Queries.RecordUptimeCheckFailure(ctx, m.ID)
	if err != nil {
		n.Logger.Error("uptime worker: record failure", "monitor_id", m.ID, "err", err)
		return
	}
	if updated.ConsecutiveFailures <= m.AlertAfterNFailures || prevStatus == db.MonitorStatusDown {
		return
	}
	if err := n.Queries.MarkUptimeMonitorDown(ctx, m.ID); err != nil {
		n.Logger.Error("uptime worker: mark down", "monitor_id", m.ID, "err", err)
		return
	}
	inc, err := n.Queries.CreateUptimeIncident(ctx, m.ID)
	if err != nil {
		n.Logger.Error("uptime worker: create incident", "monitor_id", m.ID, "err", err)
		return
	}
	alertUptimeIncident(ctx, n, m, inc, failureReason)
}

// alertUptimeIncident dispatches the down alert for a freshly-recorded
// incident, honoring the monitor's alerts-enabled flag and per-incident cap.
func alertUptimeIncident(ctx context.Context, n Notifiers, m db.UptimeMonitor, inc db.UptimeIncident, failureReason string) {
	if !m.AlertsEnabled {
		return
	}
	if max := m.MaxAlertsPerIncident; max > 0 && inc.AlertCount >= max {
		return
	}
	if !DispatchAlert(ctx, n, m.OrgID, MonitorRef{Type: "uptime", ID: m.ID}, buildUptimeDownAlert(m, failureReason)) {
		return
	}
	if _, err := n.Queries.IncrementUptimeIncidentAlertCount(ctx, inc.ID); err != nil {
		n.Logger.Error("uptime worker: increment alert count", "incident_id", inc.ID, "err", err)
	}
}

func buildUptimeDownAlert(m db.UptimeMonitor, failureReason string) AlertMessage {
	return AlertMessage{
		Telegram: fmt.Sprintf("🔴 <b>%s</b> is down\n\nURL: <code>%s</code>\nReason: %s",
			m.Name, m.Url, failureReason),
		EmailSubject: fmt.Sprintf("DOWN: %s", m.Name),
		EmailHTML: fmt.Sprintf("<p>🔴 <b>%s</b> is down</p><p>URL: <code>%s</code><br>Reason: %s</p>",
			m.Name, m.Url, failureReason),
		Webhook: &webhook.Event{
			EventType:   "down",
			MonitorName: m.Name,
			MonitorType: "uptime",
			Reason:      failureReason,
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		},
		Slack: slackMsg(slack.DownMessage(m.Name, "uptime", failureReason)),
		SMS:   TruncateSMS(fmt.Sprintf("Checkmeup: %s is DOWN (%s)", m.Name, failureReason)),
	}
}

// maxKeywordCheckBytes caps how much of the response body is read for a
// keyword search or JSON assertion, regardless of Content-Length (US-1102).
const maxKeywordCheckBytes = 512 * 1024

// uptimeCheckClient builds the *http.Client used for real monitor checks.
// m.Url is user-supplied, so it dials through httpsafe.Dialer to block
// loopback/private/link-local/cloud-metadata targets (SSRF); redirects are
// still followed (this is a generic uptime checker, unlike the one-shot
// Slack/webhook delivery clients), but each hop re-dials through the same
// Control-equipped Dialer, so a redirect into a blocked range is caught too.
func uptimeCheckClient() *http.Client {
	dialer := httpsafe.Dialer(10 * time.Second)
	return &http.Client{
		Timeout:   10 * time.Second,
		Transport: &http.Transport{DialContext: dialer.DialContext},
	}
}

// performHTTPCheck runs the monitor's HTTP check and evaluates all configured
// assertions in order: status code → keyword → JSON assertions → response-time
// threshold. The first failing condition is the recorded failure reason.
// client is injected so tests can pass an unguarded client to reach a local
// httptest server — see uptimeCheckClient for the hardened client real
// checks use.
func performHTTPCheck(m db.UptimeMonitor, client *http.Client) (statusCode int, responseTimeMs int64, isUp bool, failureReason string) {
	start := time.Now()
	resp, err := client.Get(m.Url)
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

	needsBody := keyword != "" || len(assertions) > 0
	var body string
	if needsBody {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, maxKeywordCheckBytes))
		body = string(raw)
	} else {
		_, _ = io.Copy(io.Discard, resp.Body)
	}

	if resp.StatusCode != http.StatusOK {
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

	if m.MaxResponseTimeMs.Valid && elapsed > int64(m.MaxResponseTimeMs.Int32) {
		return resp.StatusCode, elapsed, false, "response time exceeded"
	}

	return resp.StatusCode, elapsed, true, ""
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
