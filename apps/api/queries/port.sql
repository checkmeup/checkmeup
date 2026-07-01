-- name: CreatePortMonitor :one
INSERT INTO port_monitors (org_id, name, host, port, expected_state, interval_mins, max_alerts_per_incident, alert_after_n_failures)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetPortMonitor :one
SELECT * FROM port_monitors WHERE id = $1 AND org_id = $2;

-- name: ListPortMonitors :many
SELECT * FROM port_monitors WHERE org_id = $1 ORDER BY created_at DESC;

-- name: UpdatePortMonitor :one
UPDATE port_monitors
SET name = $3, host = $4, port = $5, expected_state = $6, interval_mins = $7, alerts_enabled = $8,
    max_alerts_per_incident = $9, alert_after_n_failures = $10, updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: PausePortMonitor :one
UPDATE port_monitors SET status = 'paused', updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: ResumePortMonitor :one
UPDATE port_monitors SET status = 'waiting', next_check_at = NOW(), updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: DeletePortMonitor :exec
DELETE FROM port_monitors WHERE id = $1 AND org_id = $2;

-- name: ListDuePortMonitors :many
SELECT * FROM port_monitors
WHERE next_check_at <= NOW() AND status != 'paused'
  AND NOT EXISTS (
    SELECT 1 FROM maintenance_window_monitors mwm
    JOIN maintenance_windows mw ON mw.id = mwm.window_id
    WHERE mwm.monitor_type = 'port' AND mwm.monitor_id = port_monitors.id
      AND mw.starts_at <= NOW() AND (mw.ends_at IS NULL OR mw.ends_at > NOW())
  );

-- name: RecordPortCheckUp :one
UPDATE port_monitors
SET status = 'up',
    consecutive_failures = 0,
    last_checked_at = NOW(),
    next_check_at = NOW() + (interval_mins * INTERVAL '1 minute'),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: RecordPortCheckFailure :one
UPDATE port_monitors
SET consecutive_failures = consecutive_failures + 1,
    last_checked_at = NOW(),
    next_check_at = NOW() + (interval_mins * INTERVAL '1 minute'),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: MarkPortMonitorDown :exec
UPDATE port_monitors SET status = 'down', updated_at = NOW() WHERE id = $1;

-- name: CreatePortCheck :one
INSERT INTO port_checks (monitor_id, response_time_ms, is_up, failure_reason)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListPortChecks :many
SELECT * FROM port_checks WHERE monitor_id = $1 ORDER BY checked_at DESC LIMIT $2 OFFSET $3;

-- name: ListPortChecks24h :many
SELECT * FROM port_checks
WHERE monitor_id = $1 AND checked_at >= NOW() - INTERVAL '24 hours'
ORDER BY checked_at ASC;

-- name: GetPortStats :one
SELECT
    COUNT(*) FILTER (WHERE is_up AND checked_at >= NOW() - INTERVAL '24 hours')  AS up_24h,
    COUNT(*) FILTER (WHERE checked_at >= NOW() - INTERVAL '24 hours')             AS total_24h,
    COUNT(*) FILTER (WHERE is_up AND checked_at >= NOW() - INTERVAL '7 days')    AS up_7d,
    COUNT(*) FILTER (WHERE checked_at >= NOW() - INTERVAL '7 days')               AS total_7d,
    COUNT(*) FILTER (WHERE is_up AND checked_at >= NOW() - INTERVAL '30 days')   AS up_30d,
    COUNT(*) FILTER (WHERE checked_at >= NOW() - INTERVAL '30 days')              AS total_30d
FROM port_checks
WHERE monitor_id = $1;

-- name: CreatePortIncident :one
INSERT INTO port_incidents (monitor_id) VALUES ($1) RETURNING *;

-- name: ResolveLatestPortIncident :one
UPDATE port_incidents
SET resolved_at = NOW()
WHERE id = (
    SELECT pi.id FROM port_incidents pi
    WHERE pi.monitor_id = $1 AND pi.resolved_at IS NULL
    ORDER BY pi.started_at DESC
    LIMIT 1
)
RETURNING *;

-- name: ListPortIncidents :many
SELECT * FROM port_incidents WHERE monitor_id = $1 ORDER BY started_at DESC LIMIT 200;

-- name: IncrementPortIncidentAlertCount :one
UPDATE port_incidents SET alert_count = alert_count + 1 WHERE id = $1 RETURNING *;

-- name: DeleteOldPortChecks :exec
DELETE FROM port_checks WHERE checked_at < NOW() - INTERVAL '90 days';
