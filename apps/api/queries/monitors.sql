-- name: CreateCronMonitor :one
INSERT INTO cron_monitors (org_id, name, schedule, grace_period_mins, ping_token, max_alerts_per_incident, alert_after_n_failures, max_duration_mins, project_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, sqlc.narg('project_id'))
RETURNING *;

-- name: GetCronMonitor :one
SELECT * FROM cron_monitors WHERE id = $1 AND org_id = $2;

-- name: GetCronMonitorByToken :one
SELECT * FROM cron_monitors WHERE ping_token = $1;

-- name: ListCronMonitors :many
SELECT * FROM cron_monitors WHERE org_id = $1 ORDER BY created_at DESC;

-- name: UpdateCronMonitor :one
-- A changed schedule or grace period moves next_ping_at to $10, computed
-- from now with the new ones, so the job gets one full new window; other
-- edits leave it alone, and so does a monitor still waiting for its first ping.
UPDATE cron_monitors
SET name = $3, schedule = $4, grace_period_mins = $5, alerts_enabled = $6, max_alerts_per_incident = $7,
    alert_after_n_failures = $8, max_duration_mins = $9,
    next_ping_at = CASE
        WHEN next_ping_at IS NOT NULL AND (schedule <> $4 OR grace_period_mins <> $5) THEN $10
        ELSE next_ping_at
    END,
    project_id = CASE WHEN sqlc.arg('set_project')::bool THEN sqlc.narg('project_id')::uuid ELSE project_id END,
    updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: PauseCronMonitor :one
-- Also resolves the monitor's open incident: pausing overwrites a down status,
-- so nothing would see the monitor as down again to resolve it. No recovery
-- alert goes out, since the monitor wasn't seen to recover.
WITH resolved AS (
    UPDATE cron_incidents SET resolved_at = NOW()
    WHERE resolved_at IS NULL AND monitor_id = (
        SELECT m.id FROM cron_monitors m WHERE m.id = sqlc.arg(id) AND m.org_id = sqlc.arg(org_id)
    )
)
UPDATE cron_monitors SET status = 'paused', updated_at = NOW()
WHERE cron_monitors.id = sqlc.arg(id) AND cron_monitors.org_id = sqlc.arg(org_id)
RETURNING *;

-- name: ResumeCronMonitor :one
-- A monitor that has pinged before is checked again from its next run, so a
-- job that broke while paused alerts on schedule; one that never pinged keeps
-- waiting for its first ping, as a new monitor does (PRD cron-monitoring).
UPDATE cron_monitors
SET status = CASE WHEN last_ping_at IS NULL THEN 'waiting'::monitor_status ELSE 'up'::monitor_status END,
    next_ping_at = CASE WHEN last_ping_at IS NULL THEN next_ping_at ELSE sqlc.arg(next_ping_at) END,
    consecutive_failures = 0,
    updated_at = NOW()
WHERE id = sqlc.arg(id) AND org_id = sqlc.arg(org_id)
RETURNING *;

-- name: DeleteCronMonitor :exec
DELETE FROM cron_monitors WHERE id = $1 AND org_id = $2;

-- name: UpdateCronMonitorPing :one
-- Marks the monitor up and resolves its open incidents in one statement, as
-- RecordUptimeCheckUp does (PRD alert-delivery). was_down is read under the
-- row's lock, not from the ping's earlier read, so a monitor the worker marked
-- down in between still gets its recovery. incident_started_at is when the
-- earliest resolved incident started, NULL when none was open.
WITH up AS (
    UPDATE cron_monitors m
    SET status = 'up', last_ping_at = sqlc.arg(last_ping_at), next_ping_at = sqlc.arg(next_ping_at),
        consecutive_failures = 0, updated_at = NOW()
    FROM (SELECT c.id, c.status FROM cron_monitors c WHERE c.id = sqlc.arg(id) AND c.org_id = sqlc.arg(org_id) FOR UPDATE) prev
    WHERE m.id = prev.id
    RETURNING m.id, prev.status = 'down' AS was_down
), resolved AS (
    UPDATE cron_incidents SET resolved_at = NOW()
    WHERE resolved_at IS NULL AND monitor_id IN (SELECT id FROM up)
    RETURNING started_at
)
SELECT COALESCE((SELECT was_down FROM up), false)::boolean AS was_down,
       EXISTS (SELECT 1 FROM up) AS marked_up,
       (SELECT MIN(started_at) FROM resolved)::timestamptz AS incident_started_at;

-- name: CountCronMissedRun :one
-- Counts one missed run and expects the next at next_ping_at (the schedule's
-- next run plus the grace period, as a ping now would set it), so the alert
-- filter counts missed runs, not ticks (PRD alert-delivery). Only while still
-- up and overdue: a ping since the worker read the monitor leaves it as the
-- ping set it (no row).
UPDATE cron_monitors
SET consecutive_failures = consecutive_failures + 1, next_ping_at = sqlc.arg(next_ping_at), updated_at = NOW()
WHERE id = sqlc.arg(id) AND status = 'up' AND next_ping_at < NOW()
RETURNING *;

-- name: ListOverdueCronMonitors :many
SELECT * FROM cron_monitors
WHERE status = 'up' AND next_ping_at < NOW()
  AND NOT EXISTS (
    SELECT 1 FROM maintenance_window_monitors mwm
    JOIN maintenance_windows mw ON mw.id = mwm.window_id
    WHERE mwm.monitor_type = 'cron' AND mwm.monitor_id = cron_monitors.id
      AND mw.starts_at <= NOW() AND (mw.ends_at IS NULL OR mw.ends_at > NOW())
  );

-- name: CreateCronPing :one
INSERT INTO cron_pings (monitor_id, received_at, source_ip, metadata, run_started_at)
VALUES ($1, NOW(), $2, $3, $4)
RETURNING *;

-- name: ListCronPings :many
SELECT * FROM cron_pings WHERE monitor_id = $1 ORDER BY received_at DESC LIMIT $2 OFFSET $3;

-- name: GetLatestCronPing :one
SELECT * FROM cron_pings WHERE monitor_id = $1 ORDER BY received_at DESC LIMIT 1;

-- name: CountCronPings :one
SELECT COUNT(*) FROM cron_pings WHERE monitor_id = $1;

-- name: OpenCronIncident :one
-- Marks the monitor down and opens its incident in one statement, so neither
-- happens without the other: down with no incident skips every later check.
-- Only while still up and overdue: a monitor already down has its incident,
-- and one that pinged after the worker read it as overdue isn't down (no row).
-- The incident keeps its down alert, attempted now when alerts are on, for
-- ClaimCronAlertRetries (PRD alert-delivery).
WITH down AS (
    UPDATE cron_monitors SET status = 'down', consecutive_failures = consecutive_failures + 1, updated_at = NOW()
    WHERE cron_monitors.id = sqlc.arg(monitor_id) AND cron_monitors.status = 'up' AND cron_monitors.next_ping_at < NOW()
    RETURNING id, alerts_enabled
)
INSERT INTO cron_incidents (monitor_id, alert, alert_attempted_at, alert_attempts)
SELECT down.id, sqlc.arg(alert)::jsonb,
       CASE WHEN down.alerts_enabled THEN NOW() END,
       CASE WHEN down.alerts_enabled THEN 1 ELSE 0 END
FROM down
RETURNING *;

-- name: ClaimCronAlertRetries :many
-- Takes the open incidents whose down alert never went out and wasn't
-- attempted in the last 5 minutes, on a monitor still down, unmuted and
-- under no maintenance window, and counts the attempt (PRD alert-delivery).
-- One UPDATE, so of two processes running side by side during a deploy only
-- one takes an incident.
UPDATE cron_incidents i
SET alert_attempted_at = NOW(), alert_attempts = i.alert_attempts + 1
FROM cron_monitors m
WHERE m.id = i.monitor_id
  AND i.resolved_at IS NULL AND i.alert_count = 0 AND i.alert IS NOT NULL
  AND (i.alert_attempted_at IS NULL OR i.alert_attempted_at <= NOW() - INTERVAL '5 minutes')
  AND m.status = 'down' AND m.alerts_enabled
  AND NOT EXISTS (
    SELECT 1 FROM maintenance_window_monitors mwm
    JOIN maintenance_windows mw ON mw.id = mwm.window_id
    WHERE mwm.monitor_type = 'cron' AND mwm.monitor_id = m.id
      AND mw.starts_at <= NOW() AND (mw.ends_at IS NULL OR mw.ends_at > NOW())
  )
RETURNING i.id, i.monitor_id, m.org_id, i.alert, i.alert_attempts;

-- name: ListCronIncidents :many
SELECT * FROM cron_incidents WHERE monitor_id = $1 ORDER BY started_at DESC LIMIT 200;

-- name: IncrementCronIncidentAlertCount :one
UPDATE cron_incidents SET alert_count = alert_count + 1 WHERE id = $1 RETURNING *;

-- name: DeleteOldCronPings :exec
DELETE FROM cron_pings WHERE received_at < NOW() - INTERVAL '30 days';

-- name: CreateCronRun :one
INSERT INTO cron_runs (monitor_id)
VALUES ($1)
RETURNING *;

-- name: GetOpenCronRun :one
SELECT * FROM cron_runs WHERE monitor_id = $1 AND completed_at IS NULL ORDER BY started_at DESC LIMIT 1;

-- name: CompleteCronRun :one
UPDATE cron_runs
SET completed_at = NOW()
WHERE id = (
    SELECT cr.id FROM cron_runs cr
    WHERE cr.monitor_id = $1 AND cr.completed_at IS NULL
    ORDER BY cr.started_at DESC
    LIMIT 1
)
RETURNING *;

-- name: MarkCronRunAlerted :one
UPDATE cron_runs SET alerted_at = NOW() WHERE id = $1 RETURNING *;

-- name: ListStuckCronRuns :many
SELECT cron_runs.*, cron_monitors.name AS monitor_name, cron_monitors.org_id AS monitor_org_id, cron_monitors.alerts_enabled AS monitor_alerts_enabled, cron_monitors.max_duration_mins AS monitor_max_duration_mins
FROM cron_runs
JOIN cron_monitors ON cron_monitors.id = cron_runs.monitor_id
WHERE cron_runs.completed_at IS NULL
  AND cron_runs.alerted_at IS NULL
  AND cron_monitors.max_duration_mins IS NOT NULL
  -- A paused monitor's run alerts once it is resumed, if still open (PRD alert-delivery).
  AND cron_monitors.status <> 'paused'
  AND cron_runs.started_at < NOW() - make_interval(mins => cron_monitors.max_duration_mins)
  AND NOT EXISTS (
    SELECT 1 FROM maintenance_window_monitors mwm
    JOIN maintenance_windows mw ON mw.id = mwm.window_id
    WHERE mwm.monitor_type = 'cron' AND mwm.monitor_id = cron_runs.monitor_id
      AND mw.starts_at <= NOW() AND (mw.ends_at IS NULL OR mw.ends_at > NOW())
  );

-- name: DeleteOldCronRuns :exec
DELETE FROM cron_runs WHERE started_at < NOW() - INTERVAL '30 days';
