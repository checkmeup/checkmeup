-- name: InsertSMSMessage :exec
-- monitor_type, alert_source_id and alert are set only for an SMS that was its
-- alert's only delivery, of an alert that is retried (PRD alert-delivery).
INSERT INTO sms_messages (sid, org_id, channel_id, fallback_subject, fallback_html, monitor_type, alert_source_id, alert)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: TakeSMSMessage :one
-- Deletes and returns the message, so of Twilio's several status callbacks
-- only the first final one acts on it. Looked up by Twilio's SID alone: the
-- callback carries no org context, and its URL token authenticates it. A
-- message kept for its undo (KeepSMSUndoPending) is the worker's, not a later
-- callback's.
DELETE FROM sms_messages WHERE sid = $1 AND NOT undo_pending
RETURNING org_id, channel_id, fallback_subject, fallback_html, monitor_type, alert_source_id, alert, created_at;

-- name: KeepSMSUndoPending :exec
-- Keeps a taken message whose delivery couldn't be undone yet, as when its
-- callback came before the alert was counted delivered, for the worker to
-- undo on its next tick (ListPendingSMSUndos). created_at stays the SMS's, so
-- the undo never takes back an attempt made after it.
INSERT INTO sms_messages (sid, org_id, channel_id, monitor_type, alert_source_id, alert, created_at, undo_pending)
VALUES ($1, $2, $3, $4, $5, $6, $7, true)
ON CONFLICT (sid) DO NOTHING;

-- name: ListPendingSMSUndos :many
-- Every org's messages kept for their undo: a worker read across tenants, as
-- the Claim*AlertRetries queries are.
SELECT sid, org_id, monitor_type, alert_source_id, alert, created_at
FROM sms_messages WHERE undo_pending ORDER BY created_at;

-- name: DeleteSMSMessage :exec
DELETE FROM sms_messages WHERE sid = $1 AND org_id = $2;

-- name: DeleteOldSMSMessages :exec
-- A message whose final status never arrived, or whose undo never applied.
DELETE FROM sms_messages WHERE created_at < NOW() - INTERVAL '7 days';
