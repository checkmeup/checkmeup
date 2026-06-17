-- +goose Up

ALTER TABLE users ADD COLUMN terms_version     TEXT;
ALTER TABLE users ADD COLUMN terms_accepted_at TIMESTAMPTZ;

-- +goose Down

ALTER TABLE users DROP COLUMN terms_accepted_at;
ALTER TABLE users DROP COLUMN terms_version;
