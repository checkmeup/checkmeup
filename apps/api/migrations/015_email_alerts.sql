-- +goose Up

ALTER TABLE orgs ADD COLUMN alert_email TEXT;
ALTER TABLE orgs ADD COLUMN email_alerts_enabled BOOLEAN NOT NULL DEFAULT false;

-- +goose Down

ALTER TABLE orgs DROP COLUMN email_alerts_enabled;
ALTER TABLE orgs DROP COLUMN alert_email;
