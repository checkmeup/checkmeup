-- +goose Up

-- Per-monitor cap on how many alerts to send per incident.
-- 0 = always alert. Default 3 (ADR-016).
ALTER TABLE cron_monitors
    ADD COLUMN max_alerts_per_incident INTEGER NOT NULL DEFAULT 3;

-- Track how many alerts have been sent for this incident.
ALTER TABLE cron_incidents
    ADD COLUMN alert_count INTEGER NOT NULL DEFAULT 0;

-- +goose Down

ALTER TABLE cron_incidents DROP COLUMN alert_count;
ALTER TABLE cron_monitors  DROP COLUMN max_alerts_per_incident;
