-- ─── status_pages ────────────────────────────────────────────────────────────

-- name: CreateStatusPage :one
INSERT INTO status_pages (org_id, slug, title, description, logo_url)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetStatusPage :one
SELECT * FROM status_pages WHERE id = $1 AND org_id = $2;

-- name: GetStatusPageBySlug :one
SELECT * FROM status_pages WHERE slug = $1;

-- name: ListStatusPages :many
SELECT * FROM status_pages WHERE org_id = $1 ORDER BY created_at DESC;

-- name: UpdateStatusPage :one
UPDATE status_pages
SET title = $3, description = $4, logo_url = $5, updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: DeleteStatusPage :exec
DELETE FROM status_pages WHERE id = $1 AND org_id = $2;

-- name: SlugAvailable :one
SELECT NOT EXISTS (SELECT 1 FROM status_pages WHERE slug = $1) AS available;

-- ─── status_page_monitors ────────────────────────────────────────────────────

-- name: ListStatusPageMonitors :many
SELECT * FROM status_page_monitors WHERE page_id = $1 ORDER BY display_order ASC;

-- name: DeleteStatusPageMonitors :exec
DELETE FROM status_page_monitors WHERE page_id = $1;

-- name: InsertStatusPageMonitor :one
INSERT INTO status_page_monitors (page_id, monitor_type, monitor_id, display_name, display_order)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- ─── 90-day bars for public page ─────────────────────────────────────────────

-- name: GetUptimeDailyStatus90d :many
SELECT
    date_trunc('day', checked_at AT TIME ZONE 'UTC')::date AS day,
    COUNT(*) FILTER (WHERE NOT is_up)                       AS down_count
FROM uptime_checks
WHERE monitor_id = $1 AND checked_at >= NOW() - INTERVAL '90 days'
GROUP BY 1
ORDER BY 1 ASC;

-- name: GetCronIncidentDays90d :many
SELECT DISTINCT date_trunc('day', started_at AT TIME ZONE 'UTC')::date AS day
FROM cron_incidents
WHERE monitor_id = $1 AND started_at >= NOW() - INTERVAL '90 days'
ORDER BY 1 ASC;

-- Public look-ups (no org_id check — page ownership already verified by slug)

-- name: GetCronMonitorPublic :one
SELECT id, org_id, name, schedule, grace_period_mins, ping_token, status, alerts_enabled,
       last_ping_at, next_ping_at, created_at, updated_at, max_alerts_per_incident
FROM cron_monitors WHERE id = $1;

-- name: GetUptimeMonitorPublic :one
SELECT * FROM uptime_monitors WHERE id = $1;

-- name: GetSSLMonitorPublic :one
SELECT * FROM ssl_monitors WHERE id = $1;

-- name: GetDomainMonitorPublic :one
SELECT * FROM domain_monitors WHERE id = $1;
