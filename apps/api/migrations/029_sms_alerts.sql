-- +goose Up

-- EP-19: sms becomes a real, deliverable notification_channel_type, sent via
-- Twilio (ADR-029). config shape: {"phone_number": "+972...", "consent_at": "..."}
-- — consent_at is set server-side only, once an explicit opt-in checkbox is
-- checked (TCPA-style regulatory requirement, see ADR-029).
ALTER TYPE notification_channel_type ADD VALUE 'sms';

-- +goose Down

-- Postgres can't remove an enum value, so the down migration can't fully
-- reverse the Up — matches goose's documented limitation for ADD VALUE
-- (same as 018/021).
