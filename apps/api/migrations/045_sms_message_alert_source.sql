-- +goose Up

-- PRD alert-delivery, ADR alerting: an SMS that was its alert's only delivery
-- and fails at the carrier hands the alert back to its retry when its Fallback
-- email fails too. The message keeps what it was for: the monitor type and
-- the incident's ID (uptime, port, DNS, cron) or the monitor's (SSL, domain),
-- plus, for SSL and domain, the alert, which the monitor clears once it is
-- delivered. undo_pending marks a message whose callback came before the
-- alert was counted delivered, so the worker undoes it on its next tick.
-- Messages stored before this, and recovery and stuck-run ones, have no
-- source and stay one-shot.
ALTER TABLE sms_messages
    ADD COLUMN monitor_type    TEXT,
    ADD COLUMN alert_source_id UUID,
    ADD COLUMN alert           JSONB,
    ADD COLUMN undo_pending    BOOLEAN NOT NULL DEFAULT false;
CREATE INDEX sms_messages_undo_pending_idx ON sms_messages (created_at) WHERE undo_pending;

-- +goose Down
DROP INDEX sms_messages_undo_pending_idx;
ALTER TABLE sms_messages DROP COLUMN undo_pending, DROP COLUMN alert, DROP COLUMN alert_source_id, DROP COLUMN monitor_type;
