-- name: CreateDNSMonitor :one
INSERT INTO dns_monitors (org_id, name, hostname, record_type, expected_value, interval_mins, max_alerts_per_incident, alert_after_n_failures)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetDNSMonitor :one
SELECT * FROM dns_monitors WHERE id = $1 AND org_id = $2;

-- name: ListDNSMonitors :many
SELECT * FROM dns_monitors WHERE org_id = $1 ORDER BY created_at DESC;

-- name: UpdateDNSMonitor :one
-- baseline_captured is unconditionally reset to false: an explicit edit
-- either pins a user-typed value (not a baseline) or clears it back to
-- NULL (baseline not yet re-captured) — either way the prior "auto-captured"
-- flag no longer applies. The next successful check re-captures it if the
-- value is NULL (RecordDNSCheckUp's COALESCE), same as at creation.
UPDATE dns_monitors
SET name = $3, hostname = $4, record_type = $5, expected_value = $6, baseline_captured = FALSE,
    interval_mins = $7, alerts_enabled = $8, max_alerts_per_incident = $9, alert_after_n_failures = $10,
    updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: PauseDNSMonitor :one
UPDATE dns_monitors SET status = 'paused', updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: ResumeDNSMonitor :one
UPDATE dns_monitors SET status = 'waiting', next_check_at = NOW(), updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: DeleteDNSMonitor :exec
DELETE FROM dns_monitors WHERE id = $1 AND org_id = $2;

-- name: ListDueDNSMonitors :many
SELECT * FROM dns_monitors
WHERE next_check_at <= NOW() AND status != 'paused'
  AND NOT EXISTS (
    SELECT 1 FROM maintenance_window_monitors mwm
    JOIN maintenance_windows mw ON mw.id = mwm.window_id
    WHERE mwm.monitor_type = 'dns' AND mwm.monitor_id = dns_monitors.id
      AND mw.starts_at <= NOW() AND (mw.ends_at IS NULL OR mw.ends_at > NOW())
  );

-- name: RecordDNSCheckUp :one
-- expected_value/baseline_captured only change the first time (COALESCE is
-- a no-op once expected_value is already set) — this is where a baseline
-- monitor's first successful lookup gets pinned as its comparison value.
UPDATE dns_monitors
SET status = 'up',
    consecutive_failures = 0,
    last_checked_at = NOW(),
    next_check_at = NOW() + (interval_mins * INTERVAL '1 minute'),
    last_resolved_value = $2,
    expected_value = COALESCE(expected_value, $2),
    baseline_captured = (expected_value IS NULL),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: RecordDNSCheckFailure :one
UPDATE dns_monitors
SET consecutive_failures = consecutive_failures + 1,
    last_checked_at = NOW(),
    next_check_at = NOW() + (interval_mins * INTERVAL '1 minute'),
    last_resolved_value = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: MarkDNSMonitorDown :exec
UPDATE dns_monitors SET status = 'down', updated_at = NOW() WHERE id = $1;

-- name: CreateDNSCheck :one
INSERT INTO dns_checks (monitor_id, response_time_ms, is_up, resolved_value, failure_reason)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListDNSChecks :many
SELECT * FROM dns_checks WHERE monitor_id = $1 ORDER BY checked_at DESC LIMIT $2 OFFSET $3;

-- name: ListDNSChecks24h :many
SELECT * FROM dns_checks
WHERE monitor_id = $1 AND checked_at >= NOW() - INTERVAL '24 hours'
ORDER BY checked_at ASC;

-- name: GetDNSStats :one
SELECT
    COUNT(*) FILTER (WHERE is_up AND checked_at >= NOW() - INTERVAL '24 hours')  AS up_24h,
    COUNT(*) FILTER (WHERE checked_at >= NOW() - INTERVAL '24 hours')             AS total_24h,
    COUNT(*) FILTER (WHERE is_up AND checked_at >= NOW() - INTERVAL '7 days')    AS up_7d,
    COUNT(*) FILTER (WHERE checked_at >= NOW() - INTERVAL '7 days')               AS total_7d,
    COUNT(*) FILTER (WHERE is_up AND checked_at >= NOW() - INTERVAL '30 days')   AS up_30d,
    COUNT(*) FILTER (WHERE checked_at >= NOW() - INTERVAL '30 days')              AS total_30d
FROM dns_checks
WHERE monitor_id = $1;

-- name: CreateDNSIncident :one
INSERT INTO dns_incidents (monitor_id) VALUES ($1) RETURNING *;

-- name: ResolveLatestDNSIncident :one
UPDATE dns_incidents
SET resolved_at = NOW()
WHERE id = (
    SELECT di.id FROM dns_incidents di
    WHERE di.monitor_id = $1 AND di.resolved_at IS NULL
    ORDER BY di.started_at DESC
    LIMIT 1
)
RETURNING *;

-- name: ListDNSIncidents :many
SELECT * FROM dns_incidents WHERE monitor_id = $1 ORDER BY started_at DESC LIMIT 200;

-- name: IncrementDNSIncidentAlertCount :one
UPDATE dns_incidents SET alert_count = alert_count + 1 WHERE id = $1 RETURNING *;

-- name: DeleteOldDNSChecks :exec
DELETE FROM dns_checks WHERE checked_at < NOW() - INTERVAL '90 days';

-- name: GetDNSDailyStatus90d :many
SELECT
    date_trunc('day', checked_at AT TIME ZONE 'UTC')::date AS day,
    COUNT(*) FILTER (WHERE NOT is_up)                       AS down_count
FROM dns_checks
WHERE monitor_id = $1 AND checked_at >= NOW() - INTERVAL '90 days'
GROUP BY 1;

-- name: GetDNSMonitorPublic :one
SELECT * FROM dns_monitors WHERE id = $1;
