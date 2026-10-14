-- name: InsertSMSMessage :exec
INSERT INTO sms_messages (sid, org_id, channel_id, fallback_subject, fallback_html)
VALUES ($1, $2, $3, $4, $5);

-- name: TakeSMSMessage :one
-- Deletes and returns the message, so of Twilio's several status callbacks
-- only the first final one acts on it. Looked up by Twilio's SID alone: the
-- callback carries no org context, and its URL token authenticates it.
DELETE FROM sms_messages WHERE sid = $1
RETURNING org_id, channel_id, fallback_subject, fallback_html;

-- name: DeleteOldSMSMessages :exec
-- A message whose final status never arrived.
DELETE FROM sms_messages WHERE created_at < NOW() - INTERVAL '7 days';
