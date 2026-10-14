-- Hourly check results per monitor over the last N hours, for the check
-- strips on the dashboard and the /monitors list. Each row is one monitor and
-- one hour that had checks: how many ran and how many failed. Hours with no
-- checks have no row.

-- name: ListUptimeCheckHistory :many
SELECT c.monitor_id,
       date_trunc('hour', c.checked_at)::timestamptz AS hour,
       COUNT(*)::int AS checks,
       (COUNT(*) FILTER (WHERE NOT c.is_up))::int AS failures
FROM uptime_checks c
JOIN uptime_monitors m ON m.id = c.monitor_id
WHERE m.org_id = $1
  AND c.checked_at >= date_trunc('hour', NOW()) - make_interval(hours => sqlc.arg(hours)::int - 1)
GROUP BY c.monitor_id, hour;

-- name: ListPortCheckHistory :many
SELECT c.monitor_id,
       date_trunc('hour', c.checked_at)::timestamptz AS hour,
       COUNT(*)::int AS checks,
       (COUNT(*) FILTER (WHERE NOT c.is_up))::int AS failures
FROM port_checks c
JOIN port_monitors m ON m.id = c.monitor_id
WHERE m.org_id = $1
  AND c.checked_at >= date_trunc('hour', NOW()) - make_interval(hours => sqlc.arg(hours)::int - 1)
GROUP BY c.monitor_id, hour;

-- name: ListDNSCheckHistory :many
SELECT c.monitor_id,
       date_trunc('hour', c.checked_at)::timestamptz AS hour,
       COUNT(*)::int AS checks,
       (COUNT(*) FILTER (WHERE NOT c.is_up))::int AS failures
FROM dns_checks c
JOIN dns_monitors m ON m.id = c.monitor_id
WHERE m.org_id = $1
  AND c.checked_at >= date_trunc('hour', NOW()) - make_interval(hours => sqlc.arg(hours)::int - 1)
GROUP BY c.monitor_id, hour;

-- name: ListCronPingHistory :many
-- A cron job has no failed "check" row: a ping is a success, and a missed
-- window is an incident. Pings count as checks here; incident hours come
-- from ListCronIncidentHours.
SELECT p.monitor_id,
       date_trunc('hour', p.received_at)::timestamptz AS hour,
       COUNT(*)::int AS checks,
       0::int AS failures
FROM cron_pings p
JOIN cron_monitors m ON m.id = p.monitor_id
WHERE m.org_id = $1
  AND p.received_at >= date_trunc('hour', NOW()) - make_interval(hours => sqlc.arg(hours)::int - 1)
GROUP BY p.monitor_id, hour;

-- name: ListCronIncidentHours :many
-- Every hour a cron incident was open within the window, one row per
-- monitor and hour.
SELECT DISTINCT i.monitor_id, h::timestamptz AS hour
FROM cron_incidents i
JOIN cron_monitors m ON m.id = i.monitor_id
CROSS JOIN LATERAL generate_series(
    date_trunc('hour', GREATEST(i.started_at, date_trunc('hour', NOW()) - make_interval(hours => sqlc.arg(hours)::int - 1))),
    date_trunc('hour', COALESCE(i.resolved_at, NOW())),
    INTERVAL '1 hour'
) AS h
WHERE m.org_id = $1
  AND COALESCE(i.resolved_at, NOW()) >= date_trunc('hour', NOW()) - make_interval(hours => sqlc.arg(hours)::int - 1);
