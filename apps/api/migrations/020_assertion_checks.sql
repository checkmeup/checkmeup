-- +goose Up

ALTER TABLE uptime_monitors
    ADD COLUMN json_assertions    JSONB NOT NULL DEFAULT '[]',
    ADD COLUMN max_response_time_ms INT;

-- +goose Down

ALTER TABLE uptime_monitors
    DROP COLUMN max_response_time_ms,
    DROP COLUMN json_assertions;
