-- name: GetPlatformStats :one
-- Platform-wide totals for the landing page (GET /api/v1/stats). Deliberately
-- not filtered by org_id: it returns only counts summed across every org,
-- never a row, so nothing tenant-specific leaves it. The 30-day counts scan
-- the check tables, so the handler caches the result for an hour.
SELECT
    ((SELECT COUNT(*) FROM cron_monitors   WHERE status <> 'paused')
   + (SELECT COUNT(*) FROM uptime_monitors WHERE status <> 'paused')
   + (SELECT COUNT(*) FROM ssl_monitors    WHERE status <> 'paused')
   + (SELECT COUNT(*) FROM domain_monitors WHERE status <> 'paused')
   + (SELECT COUNT(*) FROM port_monitors   WHERE status <> 'paused')
   + (SELECT COUNT(*) FROM dns_monitors    WHERE status <> 'paused'))::bigint AS monitors,
    (SELECT COUNT(*) FROM cron_pings WHERE received_at >= NOW() - INTERVAL '30 days')::bigint AS pings,
    ((SELECT COUNT(*) FROM uptime_checks WHERE checked_at >= NOW() - INTERVAL '30 days')
   + (SELECT COUNT(*) FROM port_checks   WHERE checked_at >= NOW() - INTERVAL '30 days')
   + (SELECT COUNT(*) FROM dns_checks    WHERE checked_at >= NOW() - INTERVAL '30 days'))::bigint AS checks,
    ((SELECT COUNT(*) FROM cron_incidents   WHERE started_at >= NOW() - INTERVAL '30 days')
   + (SELECT COUNT(*) FROM uptime_incidents WHERE started_at >= NOW() - INTERVAL '30 days')
   + (SELECT COUNT(*) FROM port_incidents   WHERE started_at >= NOW() - INTERVAL '30 days')
   + (SELECT COUNT(*) FROM dns_incidents    WHERE started_at >= NOW() - INTERVAL '30 days'))::bigint AS incidents;
