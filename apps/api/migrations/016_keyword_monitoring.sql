-- +goose Up

CREATE TYPE keyword_mode AS ENUM ('contains', 'not_contains');

ALTER TABLE uptime_monitors ADD COLUMN keyword TEXT;
ALTER TABLE uptime_monitors ADD COLUMN keyword_mode keyword_mode NOT NULL DEFAULT 'contains';
ALTER TABLE uptime_monitors ADD COLUMN keyword_case_sensitive BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE uptime_checks ADD COLUMN failure_reason TEXT;

-- +goose Down

ALTER TABLE uptime_checks DROP COLUMN failure_reason;

ALTER TABLE uptime_monitors DROP COLUMN keyword_case_sensitive;
ALTER TABLE uptime_monitors DROP COLUMN keyword_mode;
ALTER TABLE uptime_monitors DROP COLUMN keyword;

DROP TYPE keyword_mode;
