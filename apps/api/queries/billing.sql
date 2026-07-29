-- name: GetOrgPlan :one
SELECT plan FROM orgs WHERE id = $1;

-- name: CountOrgMonitors :one
SELECT
    (SELECT COUNT(*) FROM cron_monitors WHERE cron_monitors.org_id = $1)::int +
    (SELECT COUNT(*) FROM uptime_monitors WHERE uptime_monitors.org_id = $1)::int +
    (SELECT COUNT(*) FROM ssl_monitors WHERE ssl_monitors.org_id = $1)::int +
    (SELECT COUNT(*) FROM domain_monitors WHERE domain_monitors.org_id = $1)::int +
    (SELECT COUNT(*) FROM port_monitors WHERE port_monitors.org_id = $1)::int +
    (SELECT COUNT(*) FROM dns_monitors WHERE dns_monitors.org_id = $1)::int AS total;

-- name: CountOrgStatusPages :one
SELECT COUNT(*)::int AS total FROM status_pages WHERE org_id = $1;

-- name: CountOrgNotificationChannels :one
SELECT COUNT(*)::int AS total FROM notification_channels WHERE org_id = $1;

-- name: CountActiveMonitorsForOrg :one
-- Same shape as CountOrgMonitors, but excludes paused monitors — used to
-- gate resuming a paused monitor against the plan limit (a resume doesn't
-- create anything new, so the check is "how many active would this leave",
-- not the CountOrgMonitors total that create-blocking uses).
SELECT
    (SELECT COUNT(*) FROM cron_monitors WHERE cron_monitors.org_id = $1 AND cron_monitors.status != 'paused')::int +
    (SELECT COUNT(*) FROM uptime_monitors WHERE uptime_monitors.org_id = $1 AND uptime_monitors.status != 'paused')::int +
    (SELECT COUNT(*) FROM ssl_monitors WHERE ssl_monitors.org_id = $1 AND ssl_monitors.status != 'paused')::int +
    (SELECT COUNT(*) FROM domain_monitors WHERE domain_monitors.org_id = $1 AND domain_monitors.status != 'paused')::int +
    (SELECT COUNT(*) FROM port_monitors WHERE port_monitors.org_id = $1 AND port_monitors.status != 'paused')::int +
    (SELECT COUNT(*) FROM dns_monitors WHERE dns_monitors.org_id = $1 AND dns_monitors.status != 'paused')::int AS total;

-- name: ListActiveMonitorsForOrg :many
-- Every non-paused monitor across all 5 types, newest first — used by
-- billing.EnforceMonitorLimit (ADR-019) to decide which of an org's active
-- monitors to auto-pause after a downgrade, oldest-stays-active.
SELECT cron_monitors.id, 'cron'::text AS monitor_type, cron_monitors.created_at FROM cron_monitors WHERE cron_monitors.org_id = sqlc.arg(org_id) AND cron_monitors.status != 'paused'
UNION ALL
SELECT uptime_monitors.id, 'uptime'::text, uptime_monitors.created_at FROM uptime_monitors WHERE uptime_monitors.org_id = sqlc.arg(org_id) AND uptime_monitors.status != 'paused'
UNION ALL
SELECT ssl_monitors.id, 'ssl'::text, ssl_monitors.created_at FROM ssl_monitors WHERE ssl_monitors.org_id = sqlc.arg(org_id) AND ssl_monitors.status != 'paused'
UNION ALL
SELECT domain_monitors.id, 'domain'::text, domain_monitors.created_at FROM domain_monitors WHERE domain_monitors.org_id = sqlc.arg(org_id) AND domain_monitors.status != 'paused'
UNION ALL
SELECT port_monitors.id, 'port'::text, port_monitors.created_at FROM port_monitors WHERE port_monitors.org_id = sqlc.arg(org_id) AND port_monitors.status != 'paused'
UNION ALL
SELECT dns_monitors.id, 'dns'::text, dns_monitors.created_at FROM dns_monitors WHERE dns_monitors.org_id = sqlc.arg(org_id) AND dns_monitors.status != 'paused'
ORDER BY created_at DESC;

-- name: CountEnabledNotificationChannelsForOrg :one
SELECT COUNT(*)::int AS total FROM notification_channels WHERE org_id = $1 AND enabled = true;

-- name: SetNotificationChannelEnabled :one
-- Flips enabled without touching config — used by billing.EnforceNotificationChannelLimit
-- to auto-disable channels on downgrade, and simpler than routing through
-- UpdateNotificationChannel's full config-replace path when config isn't changing.
UPDATE notification_channels SET enabled = $3, updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING *;

-- name: GetOrgBillingInfo :one
SELECT
    o.plan,
    o.billing_cycle,
    o.paddle_customer_id,
    o.paddle_subscription_id,
    o.subscription_status,
    o.plan_renews_at,
    (
        (SELECT COUNT(*) FROM cron_monitors WHERE org_id = o.id)::int +
        (SELECT COUNT(*) FROM uptime_monitors WHERE org_id = o.id)::int +
        (SELECT COUNT(*) FROM ssl_monitors WHERE org_id = o.id)::int +
        (SELECT COUNT(*) FROM domain_monitors WHERE org_id = o.id)::int +
        (SELECT COUNT(*) FROM port_monitors WHERE org_id = o.id)::int +
        (SELECT COUNT(*) FROM dns_monitors WHERE org_id = o.id)::int
    ) AS monitor_count,
    (SELECT COUNT(*) FROM status_pages WHERE org_id = o.id)::int AS status_page_count,
    (SELECT COUNT(*) FROM notification_channels WHERE org_id = o.id)::int AS notification_channel_count,
    o.sms_credits_used_this_month,
    o.sms_credits_reset_at
FROM orgs o
WHERE o.id = $1;

-- name: ConsumeSMSCredit :one
-- Atomically applies the lazy monthly reset (ADR-032/US-1907) and consumes
-- credit_cost credits (1 in this pass — no destination weighting yet, see
-- ADR-032's "Implementation note"; a future per-destination cost band would
-- just pass a computed cost here instead of a hardcoded 1), in the same
-- round trip, only if doing so wouldn't exceed the caller-supplied plan
-- limit — the WHERE clause re-evaluates the same would-be-reset CASE as the
-- SET clause, so a stale pre-reset count can never wrongly block a send. No
-- rows are returned/updated when the org is out of credit; the caller
-- (worker.go's sendSMSAlert) treats pgx.ErrNoRows as "exhausted, skip this
-- send" rather than a hard error.
UPDATE orgs
SET
    sms_credits_used_this_month = CASE
        WHEN sms_credits_reset_at <= CURRENT_DATE THEN sqlc.arg(credit_cost)::int
        ELSE sms_credits_used_this_month + sqlc.arg(credit_cost)::int
    END,
    sms_credits_reset_at = CASE
        WHEN sms_credits_reset_at <= CURRENT_DATE THEN (date_trunc('month', NOW()) + interval '1 month')::date
        ELSE sms_credits_reset_at
    END,
    updated_at = NOW()
WHERE id = $1
    AND (CASE WHEN sms_credits_reset_at <= CURRENT_DATE THEN 0 ELSE sms_credits_used_this_month END) + sqlc.arg(credit_cost)::int <= sqlc.arg(credit_limit)::int
RETURNING sms_credits_used_this_month;

-- name: UpdateOrgPlan :exec
UPDATE orgs
SET
    plan                    = $2,
    billing_cycle           = $3,
    paddle_customer_id      = $4,
    paddle_subscription_id  = $5,
    subscription_status     = $6,
    plan_renews_at      = $7,
    updated_at          = NOW()
WHERE id = $1;
