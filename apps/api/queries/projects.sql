-- ADR projects. Every query here is scoped by org_id.

-- name: CreateProject :one
INSERT INTO projects (org_id, name)
VALUES ($1, $2)
RETURNING *;

-- name: GetProject :one
SELECT * FROM projects WHERE id = $1 AND org_id = $2;

-- name: ListProjects :many
-- Capped at 1000, the most an org can have (ADR limits).
SELECT p.*, (
    (SELECT COUNT(*) FROM cron_monitors   m WHERE m.org_id = p.org_id AND m.project_id = p.id) +
    (SELECT COUNT(*) FROM uptime_monitors m WHERE m.org_id = p.org_id AND m.project_id = p.id) +
    (SELECT COUNT(*) FROM ssl_monitors    m WHERE m.org_id = p.org_id AND m.project_id = p.id) +
    (SELECT COUNT(*) FROM domain_monitors m WHERE m.org_id = p.org_id AND m.project_id = p.id) +
    (SELECT COUNT(*) FROM port_monitors   m WHERE m.org_id = p.org_id AND m.project_id = p.id) +
    (SELECT COUNT(*) FROM dns_monitors    m WHERE m.org_id = p.org_id AND m.project_id = p.id)
)::int AS monitor_count
FROM projects p
WHERE p.org_id = $1
ORDER BY lower(p.name)
LIMIT 1000;

-- name: CountProjectMonitors :one
SELECT (
    (SELECT COUNT(*) FROM cron_monitors   WHERE org_id = sqlc.arg('org_id')::uuid AND project_id = sqlc.arg('project_id')::uuid) +
    (SELECT COUNT(*) FROM uptime_monitors WHERE org_id = sqlc.arg('org_id')::uuid AND project_id = sqlc.arg('project_id')::uuid) +
    (SELECT COUNT(*) FROM ssl_monitors    WHERE org_id = sqlc.arg('org_id')::uuid AND project_id = sqlc.arg('project_id')::uuid) +
    (SELECT COUNT(*) FROM domain_monitors WHERE org_id = sqlc.arg('org_id')::uuid AND project_id = sqlc.arg('project_id')::uuid) +
    (SELECT COUNT(*) FROM port_monitors   WHERE org_id = sqlc.arg('org_id')::uuid AND project_id = sqlc.arg('project_id')::uuid) +
    (SELECT COUNT(*) FROM dns_monitors    WHERE org_id = sqlc.arg('org_id')::uuid AND project_id = sqlc.arg('project_id')::uuid)
)::int AS monitor_count;

-- name: CountProjects :one
SELECT COUNT(*) FROM projects WHERE org_id = $1;

-- name: RenameProject :one
UPDATE projects SET name = $3, updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: DeleteProject :execrows
-- The monitors' foreign keys set their project_id to NULL.
DELETE FROM projects WHERE id = $1 AND org_id = $2;


-- name: AssignCronMonitorsToProject :many
-- Puts exactly the listed monitors of this type into the project: listed ones
-- join it, others in it leave it. Returns every row touched, so the caller can
-- tell a listed ID that isn't one of the org's monitors.
UPDATE cron_monitors
SET project_id = CASE WHEN id = ANY(sqlc.arg('monitor_ids')::uuid[]) THEN sqlc.arg('project_id')::uuid END,
    updated_at = NOW()
WHERE org_id = sqlc.arg('org_id')
  AND (id = ANY(sqlc.arg('monitor_ids')::uuid[]) OR project_id = sqlc.arg('project_id')::uuid)
RETURNING id;


-- name: AssignUptimeMonitorsToProject :many
-- Puts exactly the listed monitors of this type into the project: listed ones
-- join it, others in it leave it. Returns every row touched, so the caller can
-- tell a listed ID that isn't one of the org's monitors.
UPDATE uptime_monitors
SET project_id = CASE WHEN id = ANY(sqlc.arg('monitor_ids')::uuid[]) THEN sqlc.arg('project_id')::uuid END,
    updated_at = NOW()
WHERE org_id = sqlc.arg('org_id')
  AND (id = ANY(sqlc.arg('monitor_ids')::uuid[]) OR project_id = sqlc.arg('project_id')::uuid)
RETURNING id;


-- name: AssignSSLMonitorsToProject :many
-- Puts exactly the listed monitors of this type into the project: listed ones
-- join it, others in it leave it. Returns every row touched, so the caller can
-- tell a listed ID that isn't one of the org's monitors.
UPDATE ssl_monitors
SET project_id = CASE WHEN id = ANY(sqlc.arg('monitor_ids')::uuid[]) THEN sqlc.arg('project_id')::uuid END,
    updated_at = NOW()
WHERE org_id = sqlc.arg('org_id')
  AND (id = ANY(sqlc.arg('monitor_ids')::uuid[]) OR project_id = sqlc.arg('project_id')::uuid)
RETURNING id;


-- name: AssignDomainMonitorsToProject :many
-- Puts exactly the listed monitors of this type into the project: listed ones
-- join it, others in it leave it. Returns every row touched, so the caller can
-- tell a listed ID that isn't one of the org's monitors.
UPDATE domain_monitors
SET project_id = CASE WHEN id = ANY(sqlc.arg('monitor_ids')::uuid[]) THEN sqlc.arg('project_id')::uuid END,
    updated_at = NOW()
WHERE org_id = sqlc.arg('org_id')
  AND (id = ANY(sqlc.arg('monitor_ids')::uuid[]) OR project_id = sqlc.arg('project_id')::uuid)
RETURNING id;


-- name: AssignPortMonitorsToProject :many
-- Puts exactly the listed monitors of this type into the project: listed ones
-- join it, others in it leave it. Returns every row touched, so the caller can
-- tell a listed ID that isn't one of the org's monitors.
UPDATE port_monitors
SET project_id = CASE WHEN id = ANY(sqlc.arg('monitor_ids')::uuid[]) THEN sqlc.arg('project_id')::uuid END,
    updated_at = NOW()
WHERE org_id = sqlc.arg('org_id')
  AND (id = ANY(sqlc.arg('monitor_ids')::uuid[]) OR project_id = sqlc.arg('project_id')::uuid)
RETURNING id;


-- name: AssignDNSMonitorsToProject :many
-- Puts exactly the listed monitors of this type into the project: listed ones
-- join it, others in it leave it. Returns every row touched, so the caller can
-- tell a listed ID that isn't one of the org's monitors.
UPDATE dns_monitors
SET project_id = CASE WHEN id = ANY(sqlc.arg('monitor_ids')::uuid[]) THEN sqlc.arg('project_id')::uuid END,
    updated_at = NOW()
WHERE org_id = sqlc.arg('org_id')
  AND (id = ANY(sqlc.arg('monitor_ids')::uuid[]) OR project_id = sqlc.arg('project_id')::uuid)
RETURNING id;

-- name: GetMonitorProject :one
-- The project of one monitor of any type, for an alert's [Project] prefix
-- (ADR alerting); no row when the monitor has no project.
SELECT p.id, p.name FROM projects p
WHERE p.org_id = sqlc.arg('org_id') AND p.id = (
    CASE sqlc.arg('monitor_type')::text
        WHEN 'cron'   THEN (SELECT m.project_id FROM cron_monitors m   WHERE m.id = sqlc.arg('monitor_id') AND m.org_id = sqlc.arg('org_id'))
        WHEN 'uptime' THEN (SELECT m.project_id FROM uptime_monitors m WHERE m.id = sqlc.arg('monitor_id') AND m.org_id = sqlc.arg('org_id'))
        WHEN 'ssl'    THEN (SELECT m.project_id FROM ssl_monitors m    WHERE m.id = sqlc.arg('monitor_id') AND m.org_id = sqlc.arg('org_id'))
        WHEN 'domain' THEN (SELECT m.project_id FROM domain_monitors m WHERE m.id = sqlc.arg('monitor_id') AND m.org_id = sqlc.arg('org_id'))
        WHEN 'port'   THEN (SELECT m.project_id FROM port_monitors m   WHERE m.id = sqlc.arg('monitor_id') AND m.org_id = sqlc.arg('org_id'))
        WHEN 'dns'    THEN (SELECT m.project_id FROM dns_monitors m    WHERE m.id = sqlc.arg('monitor_id') AND m.org_id = sqlc.arg('org_id'))
    END
);
