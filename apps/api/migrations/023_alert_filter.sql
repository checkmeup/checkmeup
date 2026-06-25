-- +goose Up

-- Per-monitor filter: suppress alerts for the first N consecutive failures.
-- 0 = alert immediately (default). A successful check resets the counter.

ALTER TABLE cron_monitors
    ADD COLUMN alert_after_n_failures INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN consecutive_failures    INTEGER NOT NULL DEFAULT 0;

ALTER TABLE uptime_monitors
    ADD COLUMN alert_after_n_failures INTEGER NOT NULL DEFAULT 0;

ALTER TABLE ssl_monitors
    ADD COLUMN alert_after_n_failures INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN consecutive_failures    INTEGER NOT NULL DEFAULT 0;

ALTER TABLE domain_monitors
    ADD COLUMN alert_after_n_failures INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN consecutive_failures    INTEGER NOT NULL DEFAULT 0;

-- +goose Down

ALTER TABLE domain_monitors
    DROP COLUMN IF EXISTS consecutive_failures,
    DROP COLUMN IF EXISTS alert_after_n_failures;

ALTER TABLE ssl_monitors
    DROP COLUMN IF EXISTS consecutive_failures,
    DROP COLUMN IF EXISTS alert_after_n_failures;

ALTER TABLE uptime_monitors
    DROP COLUMN IF EXISTS alert_after_n_failures;

ALTER TABLE cron_monitors
    DROP COLUMN IF EXISTS consecutive_failures,
    DROP COLUMN IF EXISTS alert_after_n_failures;
