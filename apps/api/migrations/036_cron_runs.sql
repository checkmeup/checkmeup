-- +goose Up

-- ADR-039: start-of-run signal for zombie (EP-34) and overlap (EP-35)
-- detection. One row per run; a monitor that never calls the start-ping
-- endpoint never gets a row here, leaving today's single-ping behavior
-- unchanged. `overlap` is written by EP-35, not used yet.
CREATE TABLE cron_runs (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    monitor_id   UUID        NOT NULL REFERENCES cron_monitors(id) ON DELETE CASCADE,
    started_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    alerted_at   TIMESTAMPTZ,
    overlap      BOOLEAN     NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_cron_runs_monitor_open ON cron_runs(monitor_id) WHERE completed_at IS NULL;

-- Unset = zombie detection inactive for that monitor (US-3402).
ALTER TABLE cron_monitors ADD COLUMN max_duration_mins INTEGER;

-- Set only when a completion ping closes a matching open cron_runs row;
-- NULL means this ping had no preceding start ping (US-3404).
ALTER TABLE cron_pings ADD COLUMN run_started_at TIMESTAMPTZ;

-- +goose Down

ALTER TABLE cron_pings    DROP COLUMN run_started_at;
ALTER TABLE cron_monitors DROP COLUMN max_duration_mins;
DROP INDEX IF EXISTS idx_cron_runs_monitor_open;
DROP TABLE IF EXISTS cron_runs;
