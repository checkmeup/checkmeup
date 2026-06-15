-- +goose Up

-- Composite index for efficient paginated ping queries (by monitor, newest first)
-- and for the daily 30-day retention DELETE.
CREATE INDEX idx_cron_pings_monitor_received ON cron_pings(monitor_id, received_at DESC);

-- Drop the old single-column index — fully superseded by the composite one.
DROP INDEX idx_cron_pings_monitor_id;

-- +goose Down

CREATE INDEX idx_cron_pings_monitor_id ON cron_pings(monitor_id);
DROP INDEX idx_cron_pings_monitor_received;
