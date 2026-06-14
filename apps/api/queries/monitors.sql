-- name: CreateCronMonitor :one
INSERT INTO cron_monitors (org_id, name, schedule, grace_period_mins, ping_token)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetCronMonitor :one
SELECT * FROM cron_monitors WHERE id = $1 AND org_id = $2;

-- name: GetCronMonitorByToken :one
SELECT * FROM cron_monitors WHERE ping_token = $1;

-- name: ListCronMonitors :many
SELECT * FROM cron_monitors WHERE org_id = $1 ORDER BY created_at DESC;

-- name: UpdateCronMonitor :one
UPDATE cron_monitors
SET name = $3, schedule = $4, grace_period_mins = $5, alerts_enabled = $6, updated_at = NOW()
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
SET status = 'up', last_ping_at = $3, next_ping_at = $4, updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: UpdateCronMonitorDown :exec
UPDATE cron_monitors
SET status = 'down', updated_at = NOW()
WHERE id = $1;

-- name: ListOverdueCronMonitors :many
SELECT * FROM cron_monitors
WHERE status = 'up' AND next_ping_at < NOW();

-- name: CreateCronPing :one
INSERT INTO cron_pings (monitor_id, received_at, source_ip)
VALUES ($1, NOW(), $2)
RETURNING *;

-- name: ListCronPings :many
SELECT * FROM cron_pings WHERE monitor_id = $1 ORDER BY received_at DESC LIMIT $2 OFFSET $3;

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
SELECT * FROM cron_incidents WHERE monitor_id = $1 ORDER BY started_at DESC;
