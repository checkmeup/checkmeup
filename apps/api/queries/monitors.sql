-- name: CreateCronMonitor :one
INSERT INTO cron_monitors (org_id, name, schedule, grace_period_mins, ping_token, max_alerts_per_incident, alert_after_n_failures, max_duration_mins)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetCronMonitor :one
SELECT * FROM cron_monitors WHERE id = $1 AND org_id = $2;

-- name: GetCronMonitorByToken :one
SELECT * FROM cron_monitors WHERE ping_token = $1;

-- name: ListCronMonitors :many
SELECT * FROM cron_monitors WHERE org_id = $1 ORDER BY created_at DESC;

-- name: UpdateCronMonitor :one
UPDATE cron_monitors
SET name = $3, schedule = $4, grace_period_mins = $5, alerts_enabled = $6, max_alerts_per_incident = $7,
    alert_after_n_failures = $8, max_duration_mins = $9, updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: PauseCronMonitor :one
UPDATE cron_monitors SET status = 'paused', updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: ResumeCronMonitor :one
UPDATE cron_monitors SET status = 'waiting', updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: DeleteCronMonitor :exec
DELETE FROM cron_monitors WHERE id = $1 AND org_id = $2;

-- name: UpdateCronMonitorPing :one
UPDATE cron_monitors
SET status = 'up', last_ping_at = $3, next_ping_at = $4, consecutive_failures = 0, updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: UpdateCronMonitorDown :exec
UPDATE cron_monitors
SET status = 'down', updated_at = NOW()
WHERE id = $1;

-- name: IncrementCronConsecutiveFailures :one
UPDATE cron_monitors
SET consecutive_failures = consecutive_failures + 1, updated_at = NOW()
WHERE id = $1
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

-- name: CreateCronIncident :one
INSERT INTO cron_incidents (monitor_id)
VALUES ($1)
RETURNING *;

-- name: ResolveLatestCronIncident :one
UPDATE cron_incidents
SET resolved_at = NOW()
WHERE id = (
    SELECT ci.id FROM cron_incidents ci
    WHERE ci.monitor_id = $1 AND ci.resolved_at IS NULL
    ORDER BY ci.started_at DESC
    LIMIT 1
)
RETURNING *;

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
  AND cron_runs.started_at < NOW() - make_interval(mins => cron_monitors.max_duration_mins)
  AND NOT EXISTS (
    SELECT 1 FROM maintenance_window_monitors mwm
    JOIN maintenance_windows mw ON mw.id = mwm.window_id
    WHERE mwm.monitor_type = 'cron' AND mwm.monitor_id = cron_runs.monitor_id
      AND mw.starts_at <= NOW() AND (mw.ends_at IS NULL OR mw.ends_at > NOW())
  );

-- name: DeleteOldCronRuns :exec
DELETE FROM cron_runs WHERE started_at < NOW() - INTERVAL '30 days';
