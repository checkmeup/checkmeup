-- +goose Up

-- EP-37 US-3701: request timeout is no longer a separate concept from
-- max_response_time_ms (EP-31) — it becomes the actual HTTP client
-- deadline, so it must be required with a default matching today's
-- hardcoded 10s timeout (ADR-014), not optional.
UPDATE uptime_monitors SET max_response_time_ms = 10000 WHERE max_response_time_ms IS NULL;
ALTER TABLE uptime_monitors
    ALTER COLUMN max_response_time_ms SET NOT NULL,
    ALTER COLUMN max_response_time_ms SET DEFAULT 10000;

CREATE TYPE http_method AS ENUM ('GET', 'HEAD', 'POST');

ALTER TABLE uptime_monitors
    ADD COLUMN http_method http_method NOT NULL DEFAULT 'GET',
    ADD COLUMN accepted_status_codes INTEGER[] NOT NULL DEFAULT '{200}';

-- +goose Down

ALTER TABLE uptime_monitors
    DROP COLUMN accepted_status_codes,
    DROP COLUMN http_method;

DROP TYPE http_method;

ALTER TABLE uptime_monitors
    ALTER COLUMN max_response_time_ms DROP NOT NULL,
    ALTER COLUMN max_response_time_ms DROP DEFAULT;
