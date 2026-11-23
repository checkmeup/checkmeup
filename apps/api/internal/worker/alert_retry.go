package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/checkmeup/checkmeup/internal/db"
)

// An alert that didn't go out, for any reason (every channel and the Fallback
// email failed, the channels couldn't be read, the process stopped while
// sending), is sent again every 5 minutes while what it reports still holds
// (PRD alert-delivery): a down alert while its incident is open and its
// monitor down, an SSL or domain alert while the monitor isn't fine again.
// The incident, or the SSL or domain monitor, keeps the alert as first built,
// and the claim queries count each attempt in the database, so a restarted
// process, or a second one during a deploy, picks it up too. An alert whose
// only delivery was an SMS that then failed at the carrier, as its Fallback
// email did, is handed back here too (HandleSMSStatus): its incident's count
// goes back to 0, or the SSL or domain monitor gets its alert back.

// maxSMSAttempts is how many times one alert may try an SMS channel: each try
// uses a credit, even one Twilio refuses. The other channels and the Fallback
// email keep trying after that.
const maxSMSAttempts = 3

// alertRetry is one claimed down alert, whatever its monitor type. Its fields
// match the sqlc rows of the Claim*AlertRetries queries, so each converts to it.
type alertRetry struct {
	ID            uuid.UUID // the incident's, or the monitor's for SSL and domain
	MonitorID     uuid.UUID
	OrgID         uuid.UUID
	Alert         []byte
	AlertAttempts int32 // this attempt counted
}

// retrySource claims one monitor type's due retries and records a delivered
// one: on its incident, or for SSL and domain monitors, which keep no
// incident row, on the monitor. undo hands a delivered alert back to the
// claim, when the SMS that was its only delivery and its Fallback email both
// failed; keepsAlert is set for SSL and domain, whose monitor clears the
// alert once it is delivered, so the SMS keeps it to put back.
type retrySource struct {
	monitorType string
	claim       func(ctx context.Context) ([]alertRetry, error)
	delivered   func(ctx context.Context, id uuid.UUID) error
	undo        func(ctx context.Context, h smsHandBack) (int64, error)
	keepsAlert  bool
}

// smsHandBack is an alert to hand back to its retry, as its SMS keeps it.
type smsHandBack struct {
	OrgID       uuid.UUID
	MonitorType string
	SourceID    uuid.UUID // the incident's, or the monitor's for SSL and domain
	Alert       []byte    // SSL and domain only
	SentAt      pgtype.Timestamptz
}

func retrySources(n Notifiers) []retrySource {
	q := n.Queries
	return []retrySource{
		{
			monitorType: "uptime",
			claim: func(ctx context.Context) ([]alertRetry, error) {
				rows, err := q.ClaimUptimeAlertRetries(ctx)
				return asRetries(rows, func(r db.ClaimUptimeAlertRetriesRow) alertRetry { return alertRetry(r) }), err
			},
			delivered: func(ctx context.Context, id uuid.UUID) error {
				_, err := q.IncrementUptimeIncidentAlertCount(ctx, id)
				return err
			},
			undo: func(ctx context.Context, h smsHandBack) (int64, error) {
				return q.UndoUptimeAlertDelivery(ctx, db.UndoUptimeAlertDeliveryParams{ID: h.SourceID, OrgID: h.OrgID, SmsCreatedAt: h.SentAt})
			},
		},
		{
			monitorType: "port",
			claim: func(ctx context.Context) ([]alertRetry, error) {
				rows, err := q.ClaimPortAlertRetries(ctx)
				return asRetries(rows, func(r db.ClaimPortAlertRetriesRow) alertRetry { return alertRetry(r) }), err
			},
			delivered: func(ctx context.Context, id uuid.UUID) error {
				_, err := q.IncrementPortIncidentAlertCount(ctx, id)
				return err
			},
			undo: func(ctx context.Context, h smsHandBack) (int64, error) {
				return q.UndoPortAlertDelivery(ctx, db.UndoPortAlertDeliveryParams{ID: h.SourceID, OrgID: h.OrgID, SmsCreatedAt: h.SentAt})
			},
		},
		{
			monitorType: "dns",
			claim: func(ctx context.Context) ([]alertRetry, error) {
				rows, err := q.ClaimDNSAlertRetries(ctx)
				return asRetries(rows, func(r db.ClaimDNSAlertRetriesRow) alertRetry { return alertRetry(r) }), err
			},
			delivered: func(ctx context.Context, id uuid.UUID) error {
				_, err := q.IncrementDNSIncidentAlertCount(ctx, id)
				return err
			},
			undo: func(ctx context.Context, h smsHandBack) (int64, error) {
				return q.UndoDNSAlertDelivery(ctx, db.UndoDNSAlertDeliveryParams{ID: h.SourceID, OrgID: h.OrgID, SmsCreatedAt: h.SentAt})
			},
		},
		{
			monitorType: "cron",
			claim: func(ctx context.Context) ([]alertRetry, error) {
				rows, err := q.ClaimCronAlertRetries(ctx)
				return asRetries(rows, func(r db.ClaimCronAlertRetriesRow) alertRetry { return alertRetry(r) }), err
			},
			delivered: func(ctx context.Context, id uuid.UUID) error {
				_, err := q.IncrementCronIncidentAlertCount(ctx, id)
				return err
			},
			undo: func(ctx context.Context, h smsHandBack) (int64, error) {
				return q.UndoCronAlertDelivery(ctx, db.UndoCronAlertDeliveryParams{ID: h.SourceID, OrgID: h.OrgID, SmsCreatedAt: h.SentAt})
			},
		},
		{
			monitorType: "ssl",
			claim: func(ctx context.Context) ([]alertRetry, error) {
				rows, err := q.ClaimSSLAlertRetries(ctx)
				return asRetries(rows, func(r db.ClaimSSLAlertRetriesRow) alertRetry { return alertRetry(r) }), err
			},
			delivered: q.ClearSSLPendingAlert,
			undo: func(ctx context.Context, h smsHandBack) (int64, error) {
				return q.RestoreSSLPendingAlert(ctx, db.RestoreSSLPendingAlertParams{Alert: h.Alert, ID: h.SourceID, OrgID: h.OrgID, SmsCreatedAt: h.SentAt})
			},
			keepsAlert: true,
		},
		{
			monitorType: "domain",
			claim: func(ctx context.Context) ([]alertRetry, error) {
				rows, err := q.ClaimDomainAlertRetries(ctx)
				return asRetries(rows, func(r db.ClaimDomainAlertRetriesRow) alertRetry { return alertRetry(r) }), err
			},
			delivered: q.ClearDomainPendingAlert,
			undo: func(ctx context.Context, h smsHandBack) (int64, error) {
				return q.RestoreDomainPendingAlert(ctx, db.RestoreDomainPendingAlertParams{Alert: h.Alert, ID: h.SourceID, OrgID: h.OrgID, SmsCreatedAt: h.SentAt})
			},
			keepsAlert: true,
		},
	}
}

// undoDelivery hands h's alert back to its retry and reports whether it did:
// not while the incident is resolved, the SSL or domain monitor fine again or
// with a newer alert, or once the alert was attempted again after the SMS.
func undoDelivery(ctx context.Context, n Notifiers, h smsHandBack) (bool, error) {
	for _, src := range retrySources(n) {
		if src.monitorType != h.MonitorType {
			continue
		}
		rows, err := src.undo(ctx, h)
		return rows > 0, err
	}
	return false, fmt.Errorf("no alert retry for monitor type %q", h.MonitorType)
}

// undoPendingDeliveries undoes the deliveries HandleSMSStatus couldn't yet,
// as when Twilio's callback came before the alert was counted delivered. A
// message whose undo never applies waits until it is pruned after 7 days.
func undoPendingDeliveries(ctx context.Context, n Notifiers) {
	pending, err := n.Queries.ListPendingSMSUndos(ctx)
	if err != nil {
		n.Logger.Error("worker: list pending sms undos", "err", err)
		return
	}
	for _, p := range pending {
		undone, err := undoDelivery(ctx, n, smsHandBack{
			OrgID: p.OrgID, MonitorType: p.MonitorType.String, SourceID: p.AlertSourceID.Bytes, Alert: p.Alert, SentAt: p.CreatedAt,
		})
		if err != nil {
			n.Logger.Error("worker: hand alert back to its retry", "sid", p.Sid, "err", err)
			continue
		}
		if !undone {
			continue
		}
		if err := n.Queries.DeleteSMSMessage(ctx, db.DeleteSMSMessageParams{Sid: p.Sid, OrgID: p.OrgID}); err != nil {
			n.Logger.Error("worker: delete sms message", "sid", p.Sid, "err", err)
		}
	}
}

// asRetries converts one monitor type's claimed rows.
func asRetries[R any](rows []R, conv func(R) alertRetry) []alertRetry {
	out := make([]alertRetry, len(rows))
	for i, r := range rows {
		out[i] = conv(r)
	}
	return out
}

// retryAlerts sends again every alert due a retry. Each send runs in
// its own goroutine, so the tick doesn't wait on it.
func retryAlerts(ctx context.Context, n Notifiers) {
	undoPendingDeliveries(ctx, n)
	for _, src := range retrySources(n) {
		retries, err := src.claim(ctx)
		if err != nil {
			n.Logger.Error("worker: claim alert retries", "monitor_type", src.monitorType, "err", err)
			continue
		}
		for _, r := range retries {
			msg, err := retryMessage(r.Alert, r.AlertAttempts)
			if err != nil {
				n.Logger.Error("worker: decode alert to retry", "incident_id", r.ID, "err", err)
				continue
			}
			ref := MonitorRef{Type: src.monitorType, ID: r.MonitorID, Retry: &AlertSource{ID: r.ID}}
			if src.keepsAlert {
				ref.Retry.Alert = r.Alert
			}
			GoAlert(ctx, func(ctx context.Context) {
				if !DispatchAlert(ctx, n, r.OrgID, ref, msg) {
					return
				}
				if err := src.delivered(ctx, r.ID); err != nil {
					n.Logger.Error("worker: record delivered alert", "monitor_type", src.monitorType, "id", r.ID, "err", err)
				}
			})
		}
	}
}

// retryMessage decodes a stored alert for its attempt-th try, without the SMS
// once that is past maxSMSAttempts.
func retryMessage(alert []byte, attempt int32) (AlertMessage, error) {
	var msg AlertMessage
	if err := json.Unmarshal(alert, &msg); err != nil {
		return AlertMessage{}, err
	}
	return withSMSCap(msg, attempt), nil
}

// withSMSCap leaves out msg's SMS on an attempt past maxSMSAttempts.
func withSMSCap(msg AlertMessage, attempt int32) AlertMessage {
	if attempt > maxSMSAttempts {
		msg.SMS = ""
	}
	return msg
}

// alertJSON encodes msg for its incident to keep. An encoding error, which no
// AlertMessage should hit, leaves the incident without one: its first attempt
// still goes out, only its retries are lost.
func alertJSON(n Notifiers, monitorID uuid.UUID, msg AlertMessage) []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		n.Logger.Error("worker: encode alert", "monitor_id", monitorID, "err", err)
		return nil
	}
	return b
}
