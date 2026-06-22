-- +goose Up

-- EP-14: webhook becomes a real, deliverable notification_channel_type.
ALTER TYPE notification_channel_type ADD VALUE 'webhook';

-- Delivery status (US-1404). Generic on the table rather than webhook-only —
-- Telegram/email already log failures via slog, but only webhook's epic asks
-- for this to be user-visible, so only sendToChannel's webhook branch writes
-- it for now.
ALTER TABLE notification_channels
    ADD COLUMN last_delivery_status TEXT,
    ADD COLUMN last_delivery_detail TEXT,
    ADD COLUMN last_delivery_at TIMESTAMPTZ;

-- +goose Down

ALTER TABLE notification_channels
    DROP COLUMN last_delivery_status,
    DROP COLUMN last_delivery_detail,
    DROP COLUMN last_delivery_at;

-- Postgres can't remove an enum value, so the down migration can't fully
-- reverse the Up — matches goose's documented limitation for ADD VALUE.
