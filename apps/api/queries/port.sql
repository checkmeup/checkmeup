-- name: CreatePortMonitor :one
INSERT INTO port_monitors (org_id, name, host, port, expected_state, interval_mins, max_alerts_per_incident, alert_after_n_failures, project_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, sqlc.narg('project_id'))
RETURNING *;

-- name: GetPortMonitor :one
SELECT * FROM port_monitors WHERE id = $1 AND org_id = $2;

-- name: ListPortMonitors :many
SELECT * FROM port_monitors WHERE org_id = $1 ORDER BY created_at DESC;

-- name: UpdatePortMonitor :one
-- next_check_at: see UpdateUptimeMonitor.
UPDATE port_monitors
SET name = $3, host = $4, port = $5, expected_state = $6, interval_mins = $7, alerts_enabled = $8,
    max_alerts_per_incident = $9, alert_after_n_failures = $10,
    next_check_at = LEAST(next_check_at, last_checked_at + make_interval(mins => $7)),
    project_id = CASE WHEN sqlc.arg('set_project')::bool THEN sqlc.narg('project_id')::uuid ELSE project_id END,
    updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: PausePortMonitor :one
-- Also resolves the monitor's open incident: pausing overwrites a down status,
-- so nothing would see the monitor as down again to resolve it. No recovery
-- alert goes out, since the monitor wasn't seen to recover.
WITH resolved AS (
    UPDATE port_incidents SET resolved_at = NOW()
    WHERE resolved_at IS NULL AND monitor_id = (
        SELECT m.id FROM port_monitors m WHERE m.id = sqlc.arg(id) AND m.org_id = sqlc.arg(org_id)
    )
)
UPDATE port_monitors SET status = 'paused', updated_at = NOW()
WHERE port_monitors.id = sqlc.arg(id) AND port_monitors.org_id = sqlc.arg(org_id)
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
-- Marks the monitor up and resolves its open incidents in one statement, so
-- neither happens without the other (PRD alert-delivery): up with an open
-- incident would never resolve it. Returns when the earliest resolved one
-- started, NULL when none was open.
WITH up AS (
    UPDATE port_monitors
    SET status = 'up',
        consecutive_failures = 0,
        last_checked_at = NOW(),
        next_check_at = NOW() + (interval_mins * INTERVAL '1 minute'),
        updated_at = NOW()
    WHERE port_monitors.id = $1
    RETURNING port_monitors.id
), resolved AS (
    UPDATE port_incidents SET resolved_at = NOW()
    WHERE resolved_at IS NULL AND monitor_id IN (SELECT id FROM up)
    RETURNING started_at
)
SELECT (SELECT MIN(started_at) FROM resolved)::timestamptz AS incident_started_at;

-- name: RecordPortCheckFailure :one
UPDATE port_monitors
SET consecutive_failures = consecutive_failures + 1,
    last_checked_at = NOW(),
    next_check_at = NOW() + (interval_mins * INTERVAL '1 minute'),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SchedulePortRecheck :exec
-- A first failed check is checked again on the next tick, about 30 seconds later,
-- before it counts as a failure (ADR alerting).
UPDATE port_monitors
SET last_checked_at = NOW(),
    next_check_at = NOW() + INTERVAL '30 seconds',
    updated_at = NOW()
WHERE id = $1;

-- name: LatestPortCheckFailed :one
SELECT COALESCE(
  (SELECT NOT is_up FROM port_checks WHERE monitor_id = $1 ORDER BY checked_at DESC LIMIT 1),
  false
)::boolean AS failed;

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

-- name: OpenPortIncident :one
-- Marks the monitor down and opens its incident in one statement, so neither
-- happens without the other: down with no incident skips every later check.
-- Only from up or waiting: a monitor already down has its incident, and a
-- second check of it, as from a second process during a deploy, opens none
-- (no row). The incident keeps its down alert, attempted now when alerts are
-- on, for ClaimPortAlertRetries (PRD alert-delivery).
WITH down AS (
    UPDATE port_monitors SET status = 'down', updated_at = NOW()
    WHERE port_monitors.id = sqlc.arg(monitor_id) AND port_monitors.status IN ('up', 'waiting') RETURNING id, alerts_enabled
)
INSERT INTO port_incidents (monitor_id, alert, alert_attempted_at, alert_attempts)
SELECT down.id, sqlc.arg(alert)::jsonb,
       CASE WHEN down.alerts_enabled THEN NOW() END,
       CASE WHEN down.alerts_enabled THEN 1 ELSE 0 END
FROM down
RETURNING *;

-- name: ClaimPortAlertRetries :many
-- Takes the open incidents whose down alert never went out and wasn't
-- attempted in the last 5 minutes, on a monitor still down, unmuted and
-- under no maintenance window, and counts the attempt (PRD alert-delivery).
-- One UPDATE, so of two processes running side by side during a deploy only
-- one takes an incident.
UPDATE port_incidents i
SET alert_attempted_at = NOW(), alert_attempts = i.alert_attempts + 1
FROM port_monitors m
WHERE m.id = i.monitor_id
  AND i.resolved_at IS NULL AND i.alert_count = 0 AND i.alert IS NOT NULL
  AND (i.alert_attempted_at IS NULL OR i.alert_attempted_at <= NOW() - INTERVAL '5 minutes')
  AND m.status = 'down' AND m.alerts_enabled
  AND NOT EXISTS (
    SELECT 1 FROM maintenance_window_monitors mwm
    JOIN maintenance_windows mw ON mw.id = mwm.window_id
    WHERE mwm.monitor_type = 'port' AND mwm.monitor_id = m.id
      AND mw.starts_at <= NOW() AND (mw.ends_at IS NULL OR mw.ends_at > NOW())
  )
RETURNING i.id, i.monitor_id, m.org_id, i.alert, i.alert_attempts;

-- name: ListPortIncidents :many
SELECT * FROM port_incidents WHERE monitor_id = $1 ORDER BY started_at DESC LIMIT 200;

-- name: IncrementPortIncidentAlertCount :one
UPDATE port_incidents SET alert_count = alert_count + 1 WHERE id = $1 RETURNING *;

-- name: DeleteOldPortChecks :exec
DELETE FROM port_checks WHERE checked_at < NOW() - INTERVAL '90 days';
