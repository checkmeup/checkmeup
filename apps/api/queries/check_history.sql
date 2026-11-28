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
-- window is an incident. Every hour from the monitor's first ping on counts
-- as a passing check, not only the hours a ping landed in: a daily job pings
-- in one hour of 24, and the other 23 read as gaps, as if the job were
-- failing. Incident hours come from ListCronIncidentHours and win over these.
-- A paused monitor's hours end at its last ping. Pings are kept 30 days
-- (DeleteOldCronPings), longer than the window, so the earliest one left is
-- the first one whenever it falls inside it.
SELECT m.id AS monitor_id,
       h::timestamptz AS hour,
       1::int AS checks,
       0::int AS failures
FROM cron_monitors m
CROSS JOIN LATERAL (
    SELECT MIN(p.received_at) AS first_ping FROM cron_pings p WHERE p.monitor_id = m.id
) f
CROSS JOIN LATERAL generate_series(
    date_trunc('hour', GREATEST(f.first_ping, date_trunc('hour', NOW()) - make_interval(hours => sqlc.arg(hours)::int - 1))),
    date_trunc('hour', CASE WHEN m.status = 'paused' THEN m.last_ping_at ELSE NOW() END),
    INTERVAL '1 hour'
) AS h
WHERE m.org_id = $1
  AND f.first_ping IS NOT NULL;

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
