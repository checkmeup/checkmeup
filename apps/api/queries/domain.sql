-- name: CreateDomainMonitor :one
INSERT INTO domain_monitors (org_id, name, domain)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetDomainMonitor :one
SELECT * FROM domain_monitors WHERE id = $1 AND org_id = $2;

-- name: ListDomainMonitors :many
SELECT * FROM domain_monitors WHERE org_id = $1 ORDER BY created_at DESC;

-- name: UpdateDomainMonitor :one
UPDATE domain_monitors
SET name = $3, domain = $4, alerts_enabled = $5, alert_after_n_failures = $6, max_alerts_per_incident = $7, updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: PauseDomainMonitor :one
UPDATE domain_monitors SET status = 'paused', updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: ResumeDomainMonitor :one
UPDATE domain_monitors SET status = 'waiting', next_check_at = NOW(), updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: DeleteDomainMonitor :exec
DELETE FROM domain_monitors WHERE id = $1 AND org_id = $2;

-- name: ListDueDomainMonitors :many
SELECT * FROM domain_monitors
WHERE next_check_at <= NOW() AND status != 'paused'
  AND NOT EXISTS (
    SELECT 1 FROM maintenance_window_monitors mwm
    JOIN maintenance_windows mw ON mw.id = mwm.window_id
    WHERE mwm.monitor_type = 'domain' AND mwm.monitor_id = domain_monitors.id
      AND mw.starts_at <= NOW() AND (mw.ends_at IS NULL OR mw.ends_at > NOW())
  );

-- name: UpdateDomainMonitorCheck :one
UPDATE domain_monitors
SET status               = $2,
    expires_at           = $3,
    registrar            = $4,
    error_msg            = $5,
    alerted_30d          = $6,
    alerted_14d          = $7,
    alerted_7d           = $8,
    consecutive_failures = $9,
    alert_count          = $10,
    last_checked_at      = NOW(),
    next_check_at        = NOW() + INTERVAL '24 hours',
    updated_at           = NOW()
WHERE id = $1
RETURNING *;
