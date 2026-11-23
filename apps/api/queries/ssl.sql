-- name: CreateSSLMonitor :one
INSERT INTO ssl_monitors (org_id, name, hostname, project_id)
VALUES ($1, $2, $3, sqlc.narg('project_id'))
RETURNING *;

-- name: GetSSLMonitor :one
SELECT * FROM ssl_monitors WHERE id = $1 AND org_id = $2;

-- name: ListSSLMonitors :many
SELECT * FROM ssl_monitors WHERE org_id = $1 ORDER BY created_at DESC;

-- name: UpdateSSLMonitor :one
UPDATE ssl_monitors
SET name = $3, hostname = $4, alerts_enabled = $5, alert_after_n_failures = $6, max_alerts_per_incident = $7,
    project_id = CASE WHEN sqlc.arg('set_project')::bool THEN sqlc.narg('project_id')::uuid ELSE project_id END,
    updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: PauseSSLMonitor :one
UPDATE ssl_monitors SET status = 'paused', updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: ResumeSSLMonitor :one
UPDATE ssl_monitors SET status = 'waiting', next_check_at = NOW(), updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: DeleteSSLMonitor :exec
DELETE FROM ssl_monitors WHERE id = $1 AND org_id = $2;

-- name: ListDueSSLMonitors :many
SELECT * FROM ssl_monitors
WHERE next_check_at <= NOW() AND status != 'paused'
  AND NOT EXISTS (
    SELECT 1 FROM maintenance_window_monitors mwm
    JOIN maintenance_windows mw ON mw.id = mwm.window_id
    WHERE mwm.monitor_type = 'ssl' AND mwm.monitor_id = ssl_monitors.id
      AND mw.starts_at <= NOW() AND (mw.ends_at IS NULL OR mw.ends_at > NOW())
  );

-- name: ScheduleSSLRecheck :exec
-- A check that errors right after a working one is checked again on the
-- next tick, about 30 seconds later, before the monitor shows Error (ADR alerting).
UPDATE ssl_monitors
SET recheck_pending = true,
    last_checked_at = NOW(),
    next_check_at   = NOW() + INTERVAL '30 seconds',
    updated_at      = NOW()
WHERE id = $1;

-- name: UpdateSSLMonitorCheck :one
-- An alert this check raised is kept, attempted now, until it is delivered
-- (ClearSSLPendingAlert) or the monitor is fine again; ClaimSSLAlertRetries
-- sends it again meanwhile (PRD alert-delivery). A newer alert replaces it.
UPDATE ssl_monitors
SET status               = sqlc.arg(status),
    expires_at           = sqlc.arg(expires_at),
    issuer               = sqlc.arg(issuer),
    error_msg            = sqlc.arg(error_msg),
    alerted_30d          = sqlc.arg(alerted_30d),
    alerted_14d          = sqlc.arg(alerted_14d),
    alerted_7d           = sqlc.arg(alerted_7d),
    consecutive_failures = sqlc.arg(consecutive_failures),
    alert_count          = sqlc.arg(alert_count),
    alert                = CASE WHEN sqlc.narg(alert)::jsonb IS NOT NULL THEN sqlc.narg(alert)::jsonb
                                WHEN sqlc.arg(clear_alert)::boolean THEN NULL ELSE alert END,
    alert_attempted_at   = CASE WHEN sqlc.narg(alert)::jsonb IS NOT NULL THEN NOW() ELSE alert_attempted_at END,
    alert_attempts       = CASE WHEN sqlc.narg(alert)::jsonb IS NOT NULL THEN 1 ELSE alert_attempts END,
    recheck_pending      = false,
    last_checked_at      = NOW(),
    next_check_at        = NOW() + INTERVAL '24 hours',
    updated_at           = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;


-- name: ClearSSLPendingAlert :exec
-- The monitor's pending alert was delivered. A newer one set meanwhile would
-- be cleared with it, but a newer one comes only with the next daily check.
UPDATE ssl_monitors SET alert = NULL WHERE id = $1;

-- name: ClaimSSLAlertRetries :many
-- Takes the monitors whose pending alert wasn't attempted in the last 5
-- minutes, still with the problem, unmuted and under no maintenance window,
-- and counts the attempt (PRD alert-delivery). One UPDATE, so of two
-- processes running side by side during a deploy only one takes a monitor.
UPDATE ssl_monitors m
SET alert_attempted_at = NOW(), alert_attempts = m.alert_attempts + 1
WHERE m.alert IS NOT NULL
  AND (m.alert_attempted_at IS NULL OR m.alert_attempted_at <= NOW() - INTERVAL '5 minutes')
  AND m.status NOT IN ('up', 'paused', 'waiting') AND m.alerts_enabled
  AND NOT EXISTS (
    SELECT 1 FROM maintenance_window_monitors mwm
    JOIN maintenance_windows mw ON mw.id = mwm.window_id
    WHERE mwm.monitor_type = 'ssl' AND mwm.monitor_id = m.id
      AND mw.starts_at <= NOW() AND (mw.ends_at IS NULL OR mw.ends_at > NOW())
  )
RETURNING m.id, m.id AS monitor_id, m.org_id, m.alert, m.alert_attempts;
