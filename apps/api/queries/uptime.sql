-- name: CreateUptimeMonitor :one
INSERT INTO uptime_monitors (org_id, name, url, interval_mins, max_alerts_per_incident)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUptimeMonitor :one
SELECT * FROM uptime_monitors WHERE id = $1 AND org_id = $2;

-- name: ListUptimeMonitors :many
SELECT * FROM uptime_monitors WHERE org_id = $1 ORDER BY created_at DESC;

-- name: UpdateUptimeMonitor :one
UPDATE uptime_monitors
SET name = $3, url = $4, interval_mins = $5, alerts_enabled = $6, max_alerts_per_incident = $7, updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: PauseUptimeMonitor :one
UPDATE uptime_monitors SET status = 'paused', updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: ResumeUptimeMonitor :one
UPDATE uptime_monitors SET status = 'waiting', next_check_at = NOW(), updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: DeleteUptimeMonitor :exec
DELETE FROM uptime_monitors WHERE id = $1 AND org_id = $2;

-- name: CountUptimeMonitors :one
SELECT COUNT(*) FROM uptime_monitors WHERE org_id = $1;

-- name: ListDueUptimeMonitors :many
SELECT * FROM uptime_monitors
WHERE next_check_at <= NOW() AND status != 'paused'
  AND NOT EXISTS (
    SELECT 1 FROM maintenance_window_monitors mwm
    JOIN maintenance_windows mw ON mw.id = mwm.window_id
    WHERE mwm.monitor_type = 'uptime' AND mwm.monitor_id = uptime_monitors.id
      AND mw.starts_at <= NOW() AND (mw.ends_at IS NULL OR mw.ends_at > NOW())
  );

-- name: RecordUptimeCheckUp :one
UPDATE uptime_monitors
SET status = 'up',
    consecutive_failures = 0,
    last_checked_at = NOW(),
    next_check_at = NOW() + (interval_mins * INTERVAL '1 minute'),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: RecordUptimeCheckFailure :one
UPDATE uptime_monitors
SET consecutive_failures = consecutive_failures + 1,
    last_checked_at = NOW(),
    next_check_at = NOW() + (interval_mins * INTERVAL '1 minute'),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: MarkUptimeMonitorDown :exec
UPDATE uptime_monitors SET status = 'down', updated_at = NOW() WHERE id = $1;

-- name: CreateUptimeCheck :one
INSERT INTO uptime_checks (monitor_id, status_code, response_time_ms, is_up)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListUptimeChecks :many
SELECT * FROM uptime_checks WHERE monitor_id = $1 ORDER BY checked_at DESC LIMIT $2 OFFSET $3;

-- name: ListUptimeChecks24h :many
SELECT * FROM uptime_checks
WHERE monitor_id = $1 AND checked_at >= NOW() - INTERVAL '24 hours'
ORDER BY checked_at ASC;

-- name: GetUptimeStats :one
SELECT
    COUNT(*) FILTER (WHERE is_up AND checked_at >= NOW() - INTERVAL '24 hours')  AS up_24h,
    COUNT(*) FILTER (WHERE checked_at >= NOW() - INTERVAL '24 hours')             AS total_24h,
    COUNT(*) FILTER (WHERE is_up AND checked_at >= NOW() - INTERVAL '7 days')    AS up_7d,
    COUNT(*) FILTER (WHERE checked_at >= NOW() - INTERVAL '7 days')               AS total_7d,
    COUNT(*) FILTER (WHERE is_up AND checked_at >= NOW() - INTERVAL '30 days')   AS up_30d,
    COUNT(*) FILTER (WHERE checked_at >= NOW() - INTERVAL '30 days')              AS total_30d
FROM uptime_checks
WHERE monitor_id = $1;

-- name: CreateUptimeIncident :one
INSERT INTO uptime_incidents (monitor_id) VALUES ($1) RETURNING *;

-- name: ResolveLatestUptimeIncident :exec
UPDATE uptime_incidents
SET resolved_at = NOW()
WHERE id = (
    SELECT ui.id FROM uptime_incidents ui
    WHERE ui.monitor_id = $1 AND ui.resolved_at IS NULL
    ORDER BY ui.started_at DESC
    LIMIT 1
);

-- name: ListUptimeIncidents :many
SELECT * FROM uptime_incidents WHERE monitor_id = $1 ORDER BY started_at DESC;

-- name: IncrementUptimeIncidentAlertCount :one
UPDATE uptime_incidents SET alert_count = alert_count + 1 WHERE id = $1 RETURNING *;
