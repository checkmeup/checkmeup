-- ─── status_page_incidents ───────────────────────────────────────────────────

-- name: CreateStatusPageIncident :one
INSERT INTO status_page_incidents (org_id, title, severity)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetStatusPageIncident :one
SELECT * FROM status_page_incidents WHERE id = $1 AND org_id = $2;

-- name: ListStatusPageIncidents :many
SELECT spi.*, COUNT(spim.id) AS monitor_count
FROM status_page_incidents spi
LEFT JOIN status_page_incident_monitors spim ON spim.incident_id = spi.id
WHERE spi.org_id = $1
GROUP BY spi.id
ORDER BY spi.created_at DESC
LIMIT 200;

-- name: UpdateStatusPageIncidentTitle :one
UPDATE status_page_incidents
SET title = $3, updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: UpdateStatusPageIncidentStatus :one
UPDATE status_page_incidents
SET status = $3,
    resolved_at = CASE WHEN $3 = 'resolved'::incident_status THEN NOW() ELSE resolved_at END,
    updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: DeleteStatusPageIncident :exec
DELETE FROM status_page_incidents WHERE id = $1 AND org_id = $2;

-- name: DeleteOldStatusPageIncidents :exec
-- Only resolved incidents age out — an incident still active (however
-- old) must stay visible on the status page, never silently vanish out
-- from under it. Same 90-day window as uptime_checks/port_checks, applied
-- uniformly across every plan (see ADR-015).
DELETE FROM status_page_incidents WHERE status = 'resolved' AND resolved_at < NOW() - INTERVAL '90 days';

-- ─── status_page_incident_monitors ───────────────────────────────────────────

-- name: InsertStatusPageIncidentMonitor :one
INSERT INTO status_page_incident_monitors (incident_id, monitor_type, monitor_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListStatusPageIncidentMonitors :many
SELECT * FROM status_page_incident_monitors WHERE incident_id = $1;

-- ─── status_page_incident_updates ────────────────────────────────────────────

-- name: InsertStatusPageIncidentUpdate :one
INSERT INTO status_page_incident_updates (incident_id, message, status)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListStatusPageIncidentUpdates :many
SELECT * FROM status_page_incident_updates WHERE incident_id = $1 ORDER BY created_at DESC;

-- name: GetLatestStatusPageIncidentUpdate :one
SELECT * FROM status_page_incident_updates WHERE incident_id = $1 ORDER BY created_at DESC LIMIT 1;

-- name: UpdateStatusPageIncidentUpdateMessage :one
UPDATE status_page_incident_updates
SET message = $3
WHERE id = $1 AND incident_id = $2
RETURNING *;

-- ─── public status page lookups (scoped by page_id) ──────────────────────────

-- name: ListActiveStatusPageIncidentsForPage :many
SELECT spi.id, spi.title, spi.severity, spi.status, spi.created_at,
       string_agg(DISTINCT spm.display_name, ', ') AS affected
FROM status_page_incidents spi
JOIN status_page_incident_monitors spim ON spim.incident_id = spi.id
JOIN status_page_monitors spm ON spm.monitor_type = spim.monitor_type AND spm.monitor_id = spim.monitor_id
WHERE spm.page_id = $1 AND spi.status != 'resolved'
GROUP BY spi.id, spi.title, spi.severity, spi.status, spi.created_at
ORDER BY spi.created_at DESC
LIMIT 200;

-- name: ListResolvedStatusPageIncidentsForPage :many
SELECT DISTINCT spi.id, spi.title, spi.severity, spi.status, spi.created_at, spi.resolved_at
FROM status_page_incidents spi
JOIN status_page_incident_monitors spim ON spim.incident_id = spi.id
JOIN status_page_monitors spm ON spm.monitor_type = spim.monitor_type AND spm.monitor_id = spim.monitor_id
WHERE spm.page_id = $1 AND spi.status = 'resolved'
ORDER BY spi.resolved_at DESC
LIMIT $2 OFFSET $3;

-- name: CountResolvedStatusPageIncidentsForPage :one
SELECT COUNT(DISTINCT spi.id)
FROM status_page_incidents spi
JOIN status_page_incident_monitors spim ON spim.incident_id = spi.id
JOIN status_page_monitors spm ON spm.monitor_type = spim.monitor_type AND spm.monitor_id = spim.monitor_id
WHERE spm.page_id = $1 AND spi.status = 'resolved';
