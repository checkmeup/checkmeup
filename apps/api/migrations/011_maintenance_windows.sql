-- +goose Up

CREATE TABLE maintenance_windows (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    title       TEXT        NOT NULL,
    message     TEXT        NOT NULL DEFAULT '',
    starts_at   TIMESTAMPTZ NOT NULL,
    ends_at     TIMESTAMPTZ,            -- NULL = open-ended, ended manually
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- monitor_type is 'cron', 'uptime', or 'ssl'; no FK constraint (polymorphic, same pattern as status_page_monitors)
CREATE TABLE maintenance_window_monitors (
    id            UUID NOT NULL DEFAULT gen_random_uuid(),
    window_id     UUID NOT NULL REFERENCES maintenance_windows(id) ON DELETE CASCADE,
    monitor_type  TEXT NOT NULL,
    monitor_id    UUID NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (window_id, monitor_type, monitor_id)
);

CREATE INDEX idx_maintenance_windows_org     ON maintenance_windows(org_id);
CREATE INDEX idx_maintenance_windows_active  ON maintenance_windows(starts_at, ends_at);
CREATE INDEX idx_maintenance_window_monitors ON maintenance_window_monitors(monitor_type, monitor_id);

-- +goose Down

DROP INDEX IF EXISTS idx_maintenance_window_monitors;
DROP INDEX IF EXISTS idx_maintenance_windows_active;
DROP INDEX IF EXISTS idx_maintenance_windows_org;
DROP TABLE IF EXISTS maintenance_window_monitors;
DROP TABLE IF EXISTS maintenance_windows;
