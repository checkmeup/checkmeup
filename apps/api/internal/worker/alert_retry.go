package worker

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

// A down alert that didn't go out, for any reason (every channel and the
// Fallback email failed, the channels couldn't be read, the process stopped
// while sending), is sent again every 5 minutes while its incident is open
// and its monitor down (PRD alert-delivery). The incident keeps the alert as
// first built, and the claim queries count each attempt in the database, so
// a restarted process, or a second one during a deploy, picks it up too.

// maxSMSAttempts is how many times one alert may try an SMS channel: each try
// uses a credit, even one Twilio refuses. The other channels and the Fallback
// email keep trying after that.
const maxSMSAttempts = 3

// alertRetry is one claimed down alert, whatever its monitor type. Its fields
// match the sqlc rows of the Claim*AlertRetries queries, so each converts to it.
type alertRetry struct {
	ID            uuid.UUID // the incident's
	MonitorID     uuid.UUID
	OrgID         uuid.UUID
	Alert         []byte
	AlertAttempts int32 // this attempt counted
}

// retrySource claims one monitor type's due retries and counts a delivered one.
type retrySource struct {
	monitorType string
	claim       func(ctx context.Context) ([]alertRetry, error)
	countAlert  func(ctx context.Context, incidentID uuid.UUID) error
}

func retrySources(n Notifiers) []retrySource {
	q := n.Queries
	return []retrySource{
		{"uptime", func(ctx context.Context) ([]alertRetry, error) {
			rows, err := q.ClaimUptimeAlertRetries(ctx)
			out := make([]alertRetry, len(rows))
			for i, r := range rows {
				out[i] = alertRetry(r)
			}
			return out, err
		}, func(ctx context.Context, id uuid.UUID) error {
			_, err := q.IncrementUptimeIncidentAlertCount(ctx, id)
			return err
		}},
		{"port", func(ctx context.Context) ([]alertRetry, error) {
			rows, err := q.ClaimPortAlertRetries(ctx)
			out := make([]alertRetry, len(rows))
			for i, r := range rows {
				out[i] = alertRetry(r)
			}
			return out, err
		}, func(ctx context.Context, id uuid.UUID) error {
			_, err := q.IncrementPortIncidentAlertCount(ctx, id)
			return err
		}},
		{"dns", func(ctx context.Context) ([]alertRetry, error) {
			rows, err := q.ClaimDNSAlertRetries(ctx)
			out := make([]alertRetry, len(rows))
			for i, r := range rows {
				out[i] = alertRetry(r)
			}
			return out, err
		}, func(ctx context.Context, id uuid.UUID) error {
			_, err := q.IncrementDNSIncidentAlertCount(ctx, id)
			return err
		}},
		{"cron", func(ctx context.Context) ([]alertRetry, error) {
			rows, err := q.ClaimCronAlertRetries(ctx)
			out := make([]alertRetry, len(rows))
			for i, r := range rows {
				out[i] = alertRetry(r)
			}
			return out, err
		}, func(ctx context.Context, id uuid.UUID) error {
			_, err := q.IncrementCronIncidentAlertCount(ctx, id)
			return err
		}},
	}
}

// retryDownAlerts sends again every down alert due a retry. Each send runs in
// its own goroutine, so the tick doesn't wait on it.
func retryDownAlerts(ctx context.Context, n Notifiers) {
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
			ref := MonitorRef{Type: src.monitorType, ID: r.MonitorID}
			GoAlert(ctx, func(ctx context.Context) {
				if !DispatchAlert(ctx, n, r.OrgID, ref, msg) {
					return
				}
				if err := src.countAlert(ctx, r.ID); err != nil {
					n.Logger.Error("worker: increment alert count", "incident_id", r.ID, "err", err)
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
	if attempt > maxSMSAttempts {
		msg.SMS = ""
	}
	return msg, nil
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
