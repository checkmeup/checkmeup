-- ─── maintenance_windows ─────────────────────────────────────────────────────

-- name: CreateMaintenanceWindow :one
INSERT INTO maintenance_windows (org_id, title, message, starts_at, ends_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetMaintenanceWindow :one
SELECT * FROM maintenance_windows WHERE id = $1 AND org_id = $2;

-- name: ListMaintenanceWindows :many
SELECT mw.*, COUNT(mwm.id) AS monitor_count
FROM maintenance_windows mw
LEFT JOIN maintenance_window_monitors mwm ON mwm.window_id = mw.id
WHERE mw.org_id = $1
GROUP BY mw.id
ORDER BY mw.starts_at DESC;

-- name: UpdateMaintenanceWindow :one
UPDATE maintenance_windows
SET title = $3, message = $4, starts_at = $5, ends_at = $6, updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: EndMaintenanceWindowNow :one
UPDATE maintenance_windows
SET ends_at = NOW(), updated_at = NOW()
WHERE id = $1 AND org_id = $2 AND (ends_at IS NULL OR ends_at > NOW())
RETURNING *;

-- name: DeleteMaintenanceWindow :exec
DELETE FROM maintenance_windows WHERE id = $1 AND org_id = $2;

-- ─── maintenance_window_monitors ─────────────────────────────────────────────

-- name: ListMaintenanceWindowMonitors :many
SELECT * FROM maintenance_window_monitors WHERE window_id = $1;

-- name: DeleteMaintenanceWindowMonitors :exec
DELETE FROM maintenance_window_monitors WHERE window_id = $1;

-- name: InsertMaintenanceWindowMonitor :one
INSERT INTO maintenance_window_monitors (window_id, monitor_type, monitor_id)
VALUES ($1, $2, $3)
RETURNING *;

-- ─── public status page lookups ──────────────────────────────────────────────

-- name: GetActiveMaintenanceForOrg :many
SELECT mw.id AS window_id, mw.title, mw.message, mwm.monitor_type, mwm.monitor_id
FROM maintenance_windows mw
JOIN maintenance_window_monitors mwm ON mwm.window_id = mw.id
WHERE mw.org_id = $1
  AND mw.starts_at <= NOW() AND (mw.ends_at IS NULL OR mw.ends_at > NOW());
