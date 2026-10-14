-- +goose Up

-- ADR alerting: alerts go to notification channels; 017 copied these
-- org-level fields into channels and kept them only as a rollback path.
-- Nothing has read or written them since 1be6226.
ALTER TABLE orgs DROP COLUMN email_alerts_enabled;
ALTER TABLE orgs DROP COLUMN alert_email;
ALTER TABLE orgs DROP COLUMN telegram_chat_id;

-- +goose Down
-- Restores the columns, not their data.
ALTER TABLE orgs ADD COLUMN telegram_chat_id TEXT;
ALTER TABLE orgs ADD COLUMN alert_email TEXT;
ALTER TABLE orgs ADD COLUMN email_alerts_enabled BOOLEAN NOT NULL DEFAULT false;
