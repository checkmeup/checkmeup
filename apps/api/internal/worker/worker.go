package worker

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/checkmeup/checkmeup/internal/db"
	"github.com/checkmeup/checkmeup/internal/telegram"
)

// Run starts the background worker loops. Returns when ctx is cancelled.
//   - Every 30 s: missed-ping detection + uptime HTTP checks
//   - Every 24 h: ping retention cleanup (ADR-015)
func Run(ctx context.Context, queries *db.Queries, tg *telegram.Client, logger *slog.Logger) {
	ticker := time.NewTicker(30 * time.Second)
	cleanupTicker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	defer cleanupTicker.Stop()
	logger.Info("worker started")
	for {
		select {
		case <-ctx.Done():
			logger.Info("worker stopped")
			return
		case <-ticker.C:
			checkOverdue(ctx, queries, tg, logger)
			checkUptimeMonitors(ctx, queries, tg, logger)
			checkSSLMonitors(ctx, queries, tg, logger)
		case <-cleanupTicker.C:
			pruneOldPings(ctx, queries, logger)
		}
	}
}

func pruneOldPings(ctx context.Context, queries *db.Queries, logger *slog.Logger) {
	if err := queries.DeleteOldCronPings(ctx); err != nil {
		logger.Error("worker: prune old pings", "err", err)
	} else {
		logger.Info("worker: pruned cron_pings older than 30 days")
	}
}

// ─── cron monitor checks ─────────────────────────────────────────────────────

func checkOverdue(ctx context.Context, queries *db.Queries, tg *telegram.Client, logger *slog.Logger) {
	monitors, err := queries.ListOverdueCronMonitors(ctx)
	if err != nil {
		logger.Error("worker: list overdue monitors", "err", err)
		return
	}

	for _, m := range monitors {
		if err := queries.UpdateCronMonitorDown(ctx, m.ID); err != nil {
			logger.Error("worker: mark down", "monitor_id", m.ID, "err", err)
			continue
		}

		inc, err := queries.CreateCronIncident(ctx, m.ID)
		if err != nil {
			logger.Error("worker: create incident", "monitor_id", m.ID, "err", err)
			continue
		}

		if !m.AlertsEnabled {
			continue
		}

		max := m.MaxAlertsPerIncident
		if max > 0 && inc.AlertCount >= max {
			continue
		}

		org, err := queries.GetOrgByID(ctx, m.OrgID)
		if err != nil || !org.TelegramChatID.Valid {
			continue
		}

		missedBy := time.Since(m.NextPingAt.Time).Round(time.Second)
		msg := fmt.Sprintf(
			"🔴 <b>%s</b> missed its ping\n\nSchedule: <code>%s</code>\nExpected at: %s\nMissed by: %s",
			m.Name,
			m.Schedule,
			m.NextPingAt.Time.Format("15:04:05 MST"),
			missedBy,
		)
		if err := tg.SendMessage(org.TelegramChatID.String, msg); err != nil {
			logger.Error("worker: send down alert", "monitor_id", m.ID, "err", err)
			continue
		}
		if _, err := queries.IncrementCronIncidentAlertCount(ctx, inc.ID); err != nil {
			logger.Error("worker: increment alert count", "incident_id", inc.ID, "err", err)
		}
	}
}

// ─── uptime monitor checks ───────────────────────────────────────────────────

func checkUptimeMonitors(ctx context.Context, queries *db.Queries, tg *telegram.Client, logger *slog.Logger) {
	monitors, err := queries.ListDueUptimeMonitors(ctx)
	if err != nil {
		logger.Error("uptime worker: list due monitors", "err", err)
		return
	}

	var wg sync.WaitGroup
	for _, m := range monitors {
		m := m
		wg.Add(1)
		go func() {
			defer wg.Done()
			checkOneUptimeMonitor(ctx, queries, tg, logger, m)
		}()
	}
	wg.Wait()
}

func checkOneUptimeMonitor(ctx context.Context, queries *db.Queries, tg *telegram.Client, logger *slog.Logger, m db.UptimeMonitor) {
	statusCode, responseTimeMs, isUp := performHTTPCheck(m.Url)

	var codeParam pgtype.Int4
	if statusCode > 0 {
		codeParam = pgtype.Int4{Int32: int32(statusCode), Valid: true}
	}

	if _, err := queries.CreateUptimeCheck(ctx, db.CreateUptimeCheckParams{
		MonitorID:      m.ID,
		StatusCode:     codeParam,
		ResponseTimeMs: int32(responseTimeMs),
		IsUp:           isUp,
	}); err != nil {
		logger.Error("uptime worker: create check", "monitor_id", m.ID, "err", err)
		return
	}

	prevStatus := m.Status

	if isUp {
		if _, err := queries.RecordUptimeCheckUp(ctx, m.ID); err != nil {
			logger.Error("uptime worker: record up", "monitor_id", m.ID, "err", err)
			return
		}
		if prevStatus == db.MonitorStatusDown {
			if err := queries.ResolveLatestUptimeIncident(ctx, m.ID); err != nil {
				logger.Error("uptime worker: resolve incident", "monitor_id", m.ID, "err", err)
			}
			if m.AlertsEnabled {
				org, err := queries.GetOrgByID(ctx, m.OrgID)
				if err == nil && org.TelegramChatID.Valid {
					msg := fmt.Sprintf("✅ <b>%s</b> is back up\n\nURL: <code>%s</code>", m.Name, m.Url)
					if err := tg.SendMessage(org.TelegramChatID.String, msg); err != nil {
						logger.Error("uptime worker: send recovery alert", "monitor_id", m.ID, "err", err)
					}
				}
			}
		}
		return
	}

	// Check failed
	updated, err := queries.RecordUptimeCheckFailure(ctx, m.ID)
	if err != nil {
		logger.Error("uptime worker: record failure", "monitor_id", m.ID, "err", err)
		return
	}

	if updated.ConsecutiveFailures >= 2 && prevStatus != db.MonitorStatusDown {
		if err := queries.MarkUptimeMonitorDown(ctx, m.ID); err != nil {
			logger.Error("uptime worker: mark down", "monitor_id", m.ID, "err", err)
			return
		}
		inc, err := queries.CreateUptimeIncident(ctx, m.ID)
		if err != nil {
			logger.Error("uptime worker: create incident", "monitor_id", m.ID, "err", err)
			return
		}
		if !m.AlertsEnabled {
			return
		}
		max := m.MaxAlertsPerIncident
		if max > 0 && inc.AlertCount >= max {
			return
		}
		org, err := queries.GetOrgByID(ctx, m.OrgID)
		if err != nil || !org.TelegramChatID.Valid {
			return
		}
		msg := fmt.Sprintf("🔴 <b>%s</b> is down\n\nURL: <code>%s</code>\nStatus: %s",
			m.Name, m.Url, httpStatusDesc(statusCode))
		if err := tg.SendMessage(org.TelegramChatID.String, msg); err != nil {
			logger.Error("uptime worker: send down alert", "monitor_id", m.ID, "err", err)
			return
		}
		if _, err := queries.IncrementUptimeIncidentAlertCount(ctx, inc.ID); err != nil {
			logger.Error("uptime worker: increment alert count", "incident_id", inc.ID, "err", err)
		}
	}
}

func performHTTPCheck(rawURL string) (statusCode int, responseTimeMs int64, isUp bool) {
	client := &http.Client{Timeout: 10 * time.Second}
	start := time.Now()
	resp, err := client.Get(rawURL)
	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		return 0, elapsed, false
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, elapsed, resp.StatusCode == http.StatusOK
}

func httpStatusDesc(code int) string {
	if code == 0 {
		return "timeout / connection error"
	}
	return fmt.Sprintf("HTTP %d", code)
}

// ─── SSL monitor checks ──────────────────────────────────────────────────────

func checkSSLMonitors(ctx context.Context, queries *db.Queries, tg *telegram.Client, logger *slog.Logger) {
	monitors, err := queries.ListDueSSLMonitors(ctx)
	if err != nil {
		logger.Error("ssl worker: list due monitors", "err", err)
		return
	}

	var wg sync.WaitGroup
	for _, m := range monitors {
		m := m
		wg.Add(1)
		go func() {
			defer wg.Done()
			checkOneSSLMonitor(ctx, queries, tg, logger, m)
		}()
	}
	wg.Wait()
}

func checkOneSSLMonitor(ctx context.Context, queries *db.Queries, tg *telegram.Client, logger *slog.Logger, m db.SslMonitor) {
	expiresAt, issuer, daysLeft, err := performTLSCheck(m.Hostname)

	var newStatus db.SslMonitorStatus
	var expiresAtParam pgtype.Timestamptz
	var issuerParam pgtype.Text
	var errorMsgParam pgtype.Text
	alerted30d, alerted14d, alerted7d := m.Alerted30d, m.Alerted14d, m.Alerted7d

	if err != nil {
		newStatus = db.SslMonitorStatusError
		errorMsgParam = pgtype.Text{String: err.Error(), Valid: true}
	} else {
		expiresAtParam = pgtype.Timestamptz{Time: expiresAt, Valid: true}
		issuerParam = pgtype.Text{String: issuer, Valid: true}

		switch {
		case daysLeft < 0:
			newStatus = db.SslMonitorStatusExpired
		case daysLeft <= 30:
			newStatus = db.SslMonitorStatusExpiringSoon
		default:
			newStatus = db.SslMonitorStatusUp
			// Reset flags when cert is renewed
			alerted30d, alerted14d, alerted7d = false, false, false
		}
	}

	// Send threshold alerts (one per crossing)
	if m.AlertsEnabled && (newStatus == db.SslMonitorStatusExpiringSoon || newStatus == db.SslMonitorStatusExpired) {
		org, orgErr := queries.GetOrgByID(ctx, m.OrgID)
		if orgErr == nil && org.TelegramChatID.Valid {
			var alertMsg string
			switch {
			case daysLeft < 0 && !alerted7d:
				alertMsg = fmt.Sprintf("🔴 <b>%s</b> SSL certificate has <b>expired</b>\n\nHost: <code>%s</code>\nExpired: %s",
					m.Name, m.Hostname, expiresAt.Format("2006-01-02"))
				alerted7d = true
			case daysLeft <= 7 && !alerted7d:
				alertMsg = fmt.Sprintf("🔐 <b>%s</b> SSL certificate expires in <b>%d days</b>\n\nHost: <code>%s</code>\nExpires: %s",
					m.Name, daysLeft, m.Hostname, expiresAt.Format("2006-01-02"))
				alerted7d = true
			case daysLeft <= 14 && !alerted14d:
				alertMsg = fmt.Sprintf("🔐 <b>%s</b> SSL certificate expires in <b>%d days</b>\n\nHost: <code>%s</code>\nExpires: %s",
					m.Name, daysLeft, m.Hostname, expiresAt.Format("2006-01-02"))
				alerted14d = true
			case daysLeft <= 30 && !alerted30d:
				alertMsg = fmt.Sprintf("🔐 <b>%s</b> SSL certificate expires in <b>%d days</b>\n\nHost: <code>%s</code>\nExpires: %s",
					m.Name, daysLeft, m.Hostname, expiresAt.Format("2006-01-02"))
				alerted30d = true
			}
			if alertMsg != "" {
				if sendErr := tg.SendMessage(org.TelegramChatID.String, alertMsg); sendErr != nil {
					logger.Error("ssl worker: send alert", "monitor_id", m.ID, "err", sendErr)
				}
			}
		}
	}

	if _, err := queries.UpdateSSLMonitorCheck(ctx, db.UpdateSSLMonitorCheckParams{
		ID:         m.ID,
		Status:     newStatus,
		ExpiresAt:  expiresAtParam,
		Issuer:     issuerParam,
		ErrorMsg:   errorMsgParam,
		Alerted30d: alerted30d,
		Alerted14d: alerted14d,
		Alerted7d:  alerted7d,
	}); err != nil {
		logger.Error("ssl worker: update check", "monitor_id", m.ID, "err", err)
	}
}

func performTLSCheck(hostname string) (expiresAt time.Time, issuer string, daysLeft int, err error) {
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", hostname+":443", &tls.Config{
		ServerName: hostname,
	})
	if err != nil {
		return time.Time{}, "", 0, err
	}
	defer conn.Close() //nolint:errcheck // TLS close errors are not actionable

	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return time.Time{}, "", 0, fmt.Errorf("no peer certificates")
	}

	leaf := certs[0]
	expiresAt = leaf.NotAfter
	issuer = leaf.Issuer.CommonName
	daysLeft = int(time.Until(expiresAt).Hours() / 24)
	return expiresAt, issuer, daysLeft, nil
}
