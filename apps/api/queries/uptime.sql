-- name: CreateUptimeMonitor :one
INSERT INTO uptime_monitors (org_id, name, url, interval_mins, max_alerts_per_incident, alert_after_n_failures, keyword, keyword_mode, keyword_case_sensitive, json_assertions, max_response_time_ms, http_method, accepted_status_codes, project_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, sqlc.narg('project_id'))
RETURNING *;

-- name: GetUptimeMonitor :one
SELECT * FROM uptime_monitors WHERE id = $1 AND org_id = $2;

-- name: ListUptimeMonitors :many
SELECT * FROM uptime_monitors WHERE org_id = $1 ORDER BY created_at DESC;

-- name: UpdateUptimeMonitor :one
-- next_check_at moves up to the last check plus the new interval when that's
-- sooner, so a shorter interval applies at once; LEAST never postpones a
-- pending first check or post-resume check (last_checked_at NULL is ignored).
UPDATE uptime_monitors
SET name = $3, url = $4, interval_mins = $5, alerts_enabled = $6, max_alerts_per_incident = $7,
    alert_after_n_failures = $8, keyword = $9, keyword_mode = $10, keyword_case_sensitive = $11,
    json_assertions = $12, max_response_time_ms = $13, http_method = $14, accepted_status_codes = $15,
    next_check_at = LEAST(next_check_at, last_checked_at + make_interval(mins => $5)),
    project_id = CASE WHEN sqlc.arg('set_project')::bool THEN sqlc.narg('project_id')::uuid ELSE project_id END,
    updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: PauseUptimeMonitor :one
-- Also resolves the monitor's open incident: pausing overwrites a down status,
-- so nothing would see the monitor as down again to resolve it. No recovery
-- alert goes out, since the monitor wasn't seen to recover.
WITH resolved AS (
    UPDATE uptime_incidents SET resolved_at = NOW()
    WHERE resolved_at IS NULL AND monitor_id = (
        SELECT m.id FROM uptime_monitors m WHERE m.id = sqlc.arg(id) AND m.org_id = sqlc.arg(org_id)
    )
)
UPDATE uptime_monitors SET status = 'paused', updated_at = NOW()
WHERE uptime_monitors.id = sqlc.arg(id) AND uptime_monitors.org_id = sqlc.arg(org_id)
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

-- name: ScheduleUptimeRecheck :exec
-- A first failed check is checked again on the next tick, about 30 seconds later,
-- before it counts as a failure (ADR alerting).
UPDATE uptime_monitors
SET last_checked_at = NOW(),
    next_check_at = NOW() + INTERVAL '30 seconds',
    updated_at = NOW()
WHERE id = $1;

-- name: LatestUptimeCheckFailed :one
SELECT COALESCE(
  (SELECT NOT is_up FROM uptime_checks WHERE monitor_id = $1 ORDER BY checked_at DESC LIMIT 1),
  false
)::boolean AS failed;

-- name: CreateUptimeCheck :one
INSERT INTO uptime_checks (monitor_id, status_code, response_time_ms, is_up, failure_reason)
VALUES ($1, $2, $3, $4, $5)
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

-- name: OpenUptimeIncident :one
-- Marks the monitor down and opens its incident in one statement, so neither
-- happens without the other: down with no incident skips every later check.
-- The incident keeps its down alert, attempted now when alerts are on, for
-- ClaimUptimeAlertRetries (PRD alert-delivery).
WITH down AS (
    UPDATE uptime_monitors SET status = 'down', updated_at = NOW()
    WHERE uptime_monitors.id = sqlc.arg(monitor_id) RETURNING id, alerts_enabled
)
INSERT INTO uptime_incidents (monitor_id, alert, alert_attempted_at, alert_attempts)
SELECT down.id, sqlc.arg(alert)::jsonb,
       CASE WHEN down.alerts_enabled THEN NOW() END,
       CASE WHEN down.alerts_enabled THEN 1 ELSE 0 END
FROM down
RETURNING *;

-- name: ClaimUptimeAlertRetries :many
-- Takes the open incidents whose down alert never went out and wasn't
-- attempted in the last 5 minutes, on a monitor still down, unmuted and
-- under no maintenance window, and counts the attempt (PRD alert-delivery).
-- One UPDATE, so of two processes running side by side during a deploy only
-- one takes an incident.
UPDATE uptime_incidents i
SET alert_attempted_at = NOW(), alert_attempts = i.alert_attempts + 1
FROM uptime_monitors m
WHERE m.id = i.monitor_id
  AND i.resolved_at IS NULL AND i.alert_count = 0 AND i.alert IS NOT NULL
  AND (i.alert_attempted_at IS NULL OR i.alert_attempted_at <= NOW() - INTERVAL '5 minutes')
  AND m.status = 'down' AND m.alerts_enabled
  AND NOT EXISTS (
    SELECT 1 FROM maintenance_window_monitors mwm
    JOIN maintenance_windows mw ON mw.id = mwm.window_id
    WHERE mwm.monitor_type = 'uptime' AND mwm.monitor_id = m.id
      AND mw.starts_at <= NOW() AND (mw.ends_at IS NULL OR mw.ends_at > NOW())
  )
RETURNING i.id, i.monitor_id, m.org_id, i.alert, i.alert_attempts;

-- name: ResolveLatestUptimeIncident :one
UPDATE uptime_incidents
SET resolved_at = NOW()
WHERE id = (
    SELECT ui.id FROM uptime_incidents ui
    WHERE ui.monitor_id = $1 AND ui.resolved_at IS NULL
    ORDER BY ui.started_at DESC
    LIMIT 1
)
RETURNING *;

-- name: ListUptimeIncidents :many
SELECT * FROM uptime_incidents WHERE monitor_id = $1 ORDER BY started_at DESC LIMIT 200;

-- name: IncrementUptimeIncidentAlertCount :one
UPDATE uptime_incidents SET alert_count = alert_count + 1 WHERE id = $1 RETURNING *;

-- name: DeleteOldUptimeChecks :exec
DELETE FROM uptime_checks WHERE checked_at < NOW() - INTERVAL '90 days';
