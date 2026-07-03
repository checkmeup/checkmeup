-- +goose Up

ALTER TABLE cron_pings ADD COLUMN metadata JSONB;

-- +goose Down

ALTER TABLE cron_pings DROP COLUMN IF EXISTS metadata;
