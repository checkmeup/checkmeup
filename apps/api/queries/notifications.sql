-- ─── notification_channels ───────────────────────────────────────────────────

-- name: CreateNotificationChannel :one
INSERT INTO notification_channels (org_id, type, name, config)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListNotificationChannels :many
SELECT * FROM notification_channels WHERE org_id = $1 ORDER BY created_at DESC;

-- name: ListEnabledNotificationChannels :many
SELECT * FROM notification_channels WHERE org_id = $1 AND enabled = true ORDER BY created_at;

-- name: GetNotificationChannel :one
SELECT * FROM notification_channels WHERE id = $1 AND org_id = $2;

-- name: UpdateNotificationChannel :one
UPDATE notification_channels
SET name = $3, config = $4, enabled = $5, updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: DeleteNotificationChannel :exec
DELETE FROM notification_channels WHERE id = $1 AND org_id = $2;

-- name: UpdateNotificationChannelDelivery :exec
-- Records the outcome of the most recent send attempt (US-1404). No org_id
-- filter — only called from the worker, which already scoped the channel via
-- ListMonitorNotificationChannels, not from a user-facing handler.
UPDATE notification_channels
SET last_delivery_status = $2, last_delivery_detail = $3, last_delivery_at = NOW()
WHERE id = $1;

-- name: UpdateNotificationChannelConfig :one
-- Used for regenerating a webhook's signing secret in place, without
-- touching name/enabled (US-1403) — UpdateNotificationChannel always
-- requires both, which would force the regenerate endpoint to re-send them.
UPDATE notification_channels
SET config = $3, updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- ─── monitor_notification_channels ───────────────────────────────────────────

-- name: ListMonitorNotificationChannelIDs :many
SELECT channel_id FROM monitor_notification_channels WHERE monitor_type = $1 AND monitor_id = $2;

-- name: ListMonitorNotificationChannels :many
SELECT nc.* FROM notification_channels nc
JOIN monitor_notification_channels mnc ON mnc.channel_id = nc.id
WHERE mnc.monitor_type = $1 AND mnc.monitor_id = $2 AND nc.enabled = true;

-- name: DeleteMonitorNotificationChannels :exec
DELETE FROM monitor_notification_channels WHERE monitor_type = $1 AND monitor_id = $2;

-- name: InsertMonitorNotificationChannel :exec
INSERT INTO monitor_notification_channels (channel_id, monitor_type, monitor_id)
VALUES ($1, $2, $3);

-- ─── fallback (ADR-023) ───────────────────────────────────────────────────────

-- name: ListOrgUserEmails :many
SELECT email FROM users WHERE org_id = $1;
