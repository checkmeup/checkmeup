-- +goose Up

-- PRD alert-delivery: a down alert that didn't go out is sent again every
-- 5 minutes while its incident is open. alert holds the message as first
-- built, so a retry sends what the check found; alert_attempted_at, kept in
-- the database, stops this process or another one running beside it during a
-- deploy from sending it twice within 5 minutes; alert_attempts caps how many
-- times an SMS channel is tried, since each try uses a credit.
ALTER TABLE uptime_incidents ADD COLUMN alert JSONB, ADD COLUMN alert_attempted_at TIMESTAMPTZ, ADD COLUMN alert_attempts INTEGER NOT NULL DEFAULT 0;
ALTER TABLE port_incidents ADD COLUMN alert JSONB, ADD COLUMN alert_attempted_at TIMESTAMPTZ, ADD COLUMN alert_attempts INTEGER NOT NULL DEFAULT 0;
ALTER TABLE dns_incidents ADD COLUMN alert JSONB, ADD COLUMN alert_attempted_at TIMESTAMPTZ, ADD COLUMN alert_attempts INTEGER NOT NULL DEFAULT 0;
ALTER TABLE cron_incidents ADD COLUMN alert JSONB, ADD COLUMN alert_attempted_at TIMESTAMPTZ, ADD COLUMN alert_attempts INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE cron_incidents DROP COLUMN alert_attempts, DROP COLUMN alert_attempted_at, DROP COLUMN alert;
ALTER TABLE dns_incidents DROP COLUMN alert_attempts, DROP COLUMN alert_attempted_at, DROP COLUMN alert;
ALTER TABLE port_incidents DROP COLUMN alert_attempts, DROP COLUMN alert_attempted_at, DROP COLUMN alert;
ALTER TABLE uptime_incidents DROP COLUMN alert_attempts, DROP COLUMN alert_attempted_at, DROP COLUMN alert;
