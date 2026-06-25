-- +goose Up
ALTER TABLE ssl_monitors
    ADD COLUMN max_alerts_per_incident INTEGER NOT NULL DEFAULT 3,
    ADD COLUMN alert_count             INTEGER NOT NULL DEFAULT 0;
ALTER TABLE domain_monitors
    ADD COLUMN max_alerts_per_incident INTEGER NOT NULL DEFAULT 3,
    ADD COLUMN alert_count             INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE ssl_monitors
    DROP COLUMN max_alerts_per_incident,
    DROP COLUMN alert_count;
ALTER TABLE domain_monitors
    DROP COLUMN max_alerts_per_incident,
    DROP COLUMN alert_count;
