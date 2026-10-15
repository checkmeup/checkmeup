package worker

// Tests for naming a monitor's project in its alerts (ADR alerting, PRD
// projects): projectAlert's rendering of every alert variant, and
// DispatchAlert's lookup of the project.

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/email"
	"github.com/checkmeup/checkmeup/internal/slack"
	"github.com/checkmeup/checkmeup/internal/telegram"
	"github.com/checkmeup/checkmeup/internal/twilio"
	"github.com/checkmeup/checkmeup/internal/webhook"
)

// everyAlertVariant builds each down and recovery alert the worker and the
// ping handler send, for a monitor named "Site".
func everyAlertVariant() map[string]AlertMessage {
	past := pgtype.Timestamptz{Time: time.Now().Add(-10 * time.Minute), Valid: true}
	expires := time.Now().Add(5 * 24 * time.Hour)
	expired := time.Now().Add(-24 * time.Hour)
	threshold := func(alert func(a30, a14, a7 *bool) *AlertMessage) AlertMessage {
		var a30, a14, a7 bool
		return *alert(&a30, &a14, &a7)
	}
	ssl := db.SslMonitor{Name: "Site", Hostname: "example.com"}
	domain := db.DomainMonitor{Name: "Site", Domain: "example.com"}
	return map[string]AlertMessage{
		"cron down":           buildOverdueAlert(db.CronMonitor{Name: "Site", Schedule: "every 1h", NextPingAt: past}),
		"cron recovery":       CronRecoveryAlert("Site", "5m"),
		"cron stuck":          buildStuckRunAlert(db.ListStuckCronRunsRow{MonitorName: "Site", StartedAt: past, MonitorMaxDurationMins: pgtype.Int4{Int32: 5, Valid: true}}),
		"cron stuck recovery": StuckRunRecoveryAlert("Site", "it finished", "12m"),
		"uptime down":         buildUptimeDownAlert(db.UptimeMonitor{Name: "Site", Url: "https://example.com"}, "HTTP 500"),
		"uptime recovery":     buildUptimeRecoveryAlert(db.UptimeMonitor{Name: "Site", Url: "https://example.com"}, "5m"),
		"port down":           buildPortDownAlert(db.PortMonitor{Name: "Site", Host: "example.com", Port: 443}, "connection refused"),
		"port recovery":       buildPortRecoveryAlert(db.PortMonitor{Name: "Site", Host: "example.com", Port: 443}, "5m"),
		"dns down":            buildDNSDownAlert(db.DnsMonitor{Name: "Site", Hostname: "example.com"}, "", "NXDOMAIN"),
		"dns recovery":        buildDNSRecoveryAlert(db.DnsMonitor{Name: "Site", Hostname: "example.com"}, "1.2.3.4", "5m"),
		"ssl expiring": threshold(func(a30, a14, a7 *bool) *AlertMessage {
			return sslThresholdAlert(ssl, 5, expires, a30, a14, a7)
		}),
		"ssl expired": threshold(func(a30, a14, a7 *bool) *AlertMessage {
			return sslThresholdAlert(ssl, -1, expired, a30, a14, a7)
		}),
		"domain expiring": threshold(func(a30, a14, a7 *bool) *AlertMessage {
			return domainThresholdAlert(domain, 5, expires, a30, a14, a7)
		}),
		"domain expired": threshold(func(a30, a14, a7 *bool) *AlertMessage {
			return domainThresholdAlert(domain, -1, expired, a30, a14, a7)
		}),
	}
}

// wantInsertedOnce checks got is want with prefix inserted once, at marker.
func wantInsertedOnce(t *testing.T, field, got, want, prefix, marker string) {
	t.Helper()
	if !strings.Contains(got, prefix+marker) {
		t.Errorf("%s = %q, want %q before %q", field, got, prefix, marker)
	}
	if strings.Replace(got, prefix, "", 1) != want {
		t.Errorf("%s = %q, want %q with only %q added", field, got, want, prefix)
	}
}

func TestProjectAlert_NamesTheProjectOnEveryChannel(t *testing.T) {
	for name, msg := range everyAlertVariant() {
		t.Run(name, func(t *testing.T) {
			got := projectAlert(msg, "p-1", "Acme")

			wantInsertedOnce(t, "Telegram", got.Telegram, msg.Telegram, "[Acme] ", "<b>Site</b>")
			wantInsertedOnce(t, "EmailHTML", got.EmailHTML, msg.EmailHTML, "[Acme] ", "<b>Site</b>")
			if got.EmailSubject != "[Acme] "+msg.EmailSubject {
				t.Errorf("EmailSubject = %q, want %q", got.EmailSubject, "[Acme] "+msg.EmailSubject)
			}
			wantInsertedOnce(t, "Slack text", got.Slack.Text, msg.Slack.Text, "[Acme] ", "Site")
			if got.SMS != TruncateSMS("Checkmeup: [Acme] "+strings.TrimPrefix(msg.SMS, "Checkmeup: ")) {
				t.Errorf("SMS = %q, want %q with [Acme] after Checkmeup:", got.SMS, msg.SMS)
			}
			if got.Webhook.ProjectID != "p-1" || got.Webhook.ProjectName != "Acme" {
				t.Errorf("Webhook = %+v, want projectId p-1 and projectName Acme", got.Webhook)
			}
			if msg.Webhook.ProjectName != "" {
				t.Error("want the original message's webhook event left as it was")
			}
		})
	}
}

func TestProjectAlert_SlackHeader(t *testing.T) {
	got := projectAlert(AlertMessage{Slack: slackMsg(slack.DownMessage("Site", "uptime", "HTTP 500"))}, "p-1", "A&B")
	body, err := json.Marshal(got.Slack)
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Blocks []struct {
			Type string `json:"type"`
			Text struct {
				Text string `json:"text"`
			} `json:"text"`
		} `json:"blocks"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	// The header block is plain_text, shown as written; the notification
	// text is mrkdwn, where & has to be escaped.
	if h := decoded.Blocks[0]; h.Type != "header" || h.Text.Text != "🔴 [A&B] Site is down" {
		t.Errorf("header = %+v, want 🔴 [A&B] Site is down", h)
	}
	if got.Slack.Text != "🔴 [A&amp;B] Site is down" {
		t.Errorf("Slack text = %q, want %q", got.Slack.Text, "🔴 [A&amp;B] Site is down")
	}
}

func TestProjectAlert_EscapesTheProjectNameInHTML(t *testing.T) {
	msg := buildUptimeDownAlert(db.UptimeMonitor{Name: "Site", Url: "https://example.com"}, "HTTP 500")
	got := projectAlert(msg, "p-1", "<Acme & Co>")
	for field, s := range map[string]string{"Telegram": got.Telegram, "EmailHTML": got.EmailHTML} {
		if !strings.Contains(s, "[&lt;Acme &amp; Co&gt;] <b>Site</b>") {
			t.Errorf("%s = %q, want the project name HTML-escaped", field, s)
		}
	}
	if got.EmailSubject != "[<Acme & Co>] DOWN: Site" {
		t.Errorf("EmailSubject = %q, want the plain-text name", got.EmailSubject)
	}
}

func TestProjectAlert_KeepsSMSToOneSegment(t *testing.T) {
	msg := buildUptimeDownAlert(db.UptimeMonitor{Name: "Site"}, strings.Repeat("x", 200))
	got := projectAlert(msg, "p-1", strings.Repeat("A", 100))
	if n := len([]rune(got.SMS)); n > smsSegmentLimit {
		t.Errorf("SMS is %d characters, want at most %d", n, smsSegmentLimit)
	}
}

func TestDispatchAlert_Project(t *testing.T) {
	pool := testPool(t)
	queries := db.New(pool)
	n := Notifiers{
		Queries: queries, Telegram: telegram.NewClient(""), Mailer: email.NewSender(""),
		Webhook: webhook.NewClientWithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
		Slack:   slack.NewClientWithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
		SMS:     twilio.NewClient("", "", "", "", ""), Logger: testLogger(),
	}

	// webhookBody dispatches msg for monitor to a webhook channel and
	// returns the JSON body it received.
	webhookBody := func(t *testing.T, org db.Org, monitorID uuid.UUID, dispatchOrg uuid.UUID) map[string]any {
		t.Helper()
		var body []byte
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()
		channel := testNotificationChannel(t, queries, org.ID, db.NotificationChannelTypeWebhook, map[string]string{"url": srv.URL, "secret": "shh"})
		attachNotificationChannel(t, queries, channel.ID, "cron", monitorID)

		msg := CronRecoveryAlert("Site", "5m")
		DispatchAlert(context.Background(), n, dispatchOrg, MonitorRef{Type: "cron", ID: monitorID}, msg)
		var got map[string]any
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("decode webhook body %q: %v", body, err)
		}
		return got
	}

	t.Run("a monitor in a project sends its project on the webhook event", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		m := testCronMonitor(t, queries, org.ID, "every 1h")
		var projectID uuid.UUID
		if err := pool.QueryRow(context.Background(), "INSERT INTO projects (org_id, name) VALUES ($1, 'Acme') RETURNING id", org.ID).Scan(&projectID); err != nil {
			t.Fatal(err)
		}
		mustExecWorker(t, pool, "UPDATE cron_monitors SET project_id = $1 WHERE id = $2", projectID, m.ID)

		got := webhookBody(t, org, m.ID, org.ID)
		if got["projectId"] != projectID.String() || got["projectName"] != "Acme" {
			t.Fatalf("want projectId %s and projectName Acme, got %v", projectID, got)
		}
	})

	t.Run("a monitor without a project sends the event without project fields", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		m := testCronMonitor(t, queries, org.ID, "every 1h")
		got := webhookBody(t, org, m.ID, org.ID)
		if _, ok := got["projectId"]; ok {
			t.Fatalf("want no projectId key, got %v", got)
		}
		if _, ok := got["projectName"]; ok {
			t.Fatalf("want no projectName key, got %v", got)
		}
	})

	t.Run("a monitor without a project gets the message exactly as built", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		m := testCronMonitor(t, queries, org.ID, "every 1h")
		for name, msg := range everyAlertVariant() {
			got := withMonitorProject(context.Background(), n, org.ID, MonitorRef{Type: "cron", ID: m.ID}, msg)
			if !reflect.DeepEqual(got, msg) {
				t.Errorf("%s: want the message unchanged, got %+v", name, got)
			}
		}
	})

	t.Run("a project is only looked up in the alert's own org", func(t *testing.T) {
		org := testOrg(t, queries, pool)
		other := testOrg(t, queries, pool)
		m := testCronMonitor(t, queries, org.ID, "every 1h")
		var projectID uuid.UUID
		if err := pool.QueryRow(context.Background(), "INSERT INTO projects (org_id, name) VALUES ($1, 'Acme') RETURNING id", org.ID).Scan(&projectID); err != nil {
			t.Fatal(err)
		}
		mustExecWorker(t, pool, "UPDATE cron_monitors SET project_id = $1 WHERE id = $2", projectID, m.ID)

		msg := CronRecoveryAlert("Site", "5m")
		got := withMonitorProject(context.Background(), n, other.ID, MonitorRef{Type: "cron", ID: m.ID}, msg)
		if !reflect.DeepEqual(got, msg) {
			t.Fatalf("want no project from another org's lookup, got %+v", got.Webhook)
		}
	})
}
