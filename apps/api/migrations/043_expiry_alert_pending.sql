-- +goose Up

-- PRD alert-delivery: an SSL or domain alert that didn't go out is sent again
-- every 5 minutes while its monitor still has the problem. As for incidents
-- (041), the monitor keeps the alert as built and the time of its last
-- attempt, so a restarted process, or a second one during a deploy, picks it
-- up without sending it twice.
ALTER TABLE ssl_monitors ADD COLUMN alert JSONB, ADD COLUMN alert_attempted_at TIMESTAMPTZ;
ALTER TABLE domain_monitors ADD COLUMN alert JSONB, ADD COLUMN alert_attempted_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE domain_monitors DROP COLUMN alert_attempted_at, DROP COLUMN alert;
ALTER TABLE ssl_monitors DROP COLUMN alert_attempted_at, DROP COLUMN alert;
