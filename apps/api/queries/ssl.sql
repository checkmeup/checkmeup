-- name: CreateSSLMonitor :one
INSERT INTO ssl_monitors (org_id, name, hostname)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSSLMonitor :one
SELECT * FROM ssl_monitors WHERE id = $1 AND org_id = $2;

-- name: ListSSLMonitors :many
SELECT * FROM ssl_monitors WHERE org_id = $1 ORDER BY created_at DESC;

-- name: UpdateSSLMonitor :one
UPDATE ssl_monitors
SET name = $3, hostname = $4, alerts_enabled = $5, updated_at = NOW()
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
SELECT * FROM ssl_monitors WHERE next_check_at <= NOW() AND status != 'paused';

-- name: UpdateSSLMonitorCheck :one
UPDATE ssl_monitors
SET status          = $2,
    expires_at      = $3,
    issuer          = $4,
    error_msg       = $5,
    alerted_30d     = $6,
    alerted_14d     = $7,
    alerted_7d      = $8,
    last_checked_at = NOW(),
    next_check_at   = NOW() + INTERVAL '24 hours',
    updated_at      = NOW()
WHERE id = $1
RETURNING *;
