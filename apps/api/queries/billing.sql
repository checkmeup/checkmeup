-- name: GetOrgPlan :one
SELECT plan FROM orgs WHERE id = $1;

-- name: CountOrgMonitors :one
SELECT
    (SELECT COUNT(*) FROM cron_monitors WHERE cron_monitors.org_id = $1)::int +
    (SELECT COUNT(*) FROM uptime_monitors WHERE uptime_monitors.org_id = $1)::int +
    (SELECT COUNT(*) FROM ssl_monitors WHERE ssl_monitors.org_id = $1)::int AS total;

-- name: CountOrgStatusPages :one
SELECT COUNT(*)::int AS total FROM status_pages WHERE org_id = $1;

-- name: GetOrgBillingInfo :one
SELECT
    o.plan,
    o.ls_customer_id,
    o.ls_subscription_id,
    o.subscription_status,
    o.plan_renews_at,
    (
        (SELECT COUNT(*) FROM cron_monitors WHERE org_id = o.id)::int +
        (SELECT COUNT(*) FROM uptime_monitors WHERE org_id = o.id)::int +
        (SELECT COUNT(*) FROM ssl_monitors WHERE org_id = o.id)::int
    ) AS monitor_count,
    (SELECT COUNT(*) FROM status_pages WHERE org_id = o.id)::int AS status_page_count
FROM orgs o
WHERE o.id = $1;

-- name: UpdateOrgPlan :exec
UPDATE orgs
SET
    plan                = $2,
    ls_customer_id      = $3,
    ls_subscription_id  = $4,
    subscription_status = $5,
    plan_renews_at      = $6,
    updated_at          = NOW()
WHERE id = $1;
