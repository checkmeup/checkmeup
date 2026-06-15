-- +goose Up

CREATE TABLE uptime_monitors (
    id                      UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID           NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name                    TEXT           NOT NULL,
    url                     TEXT           NOT NULL,
    interval_mins           INTEGER        NOT NULL DEFAULT 10,
    status                  monitor_status NOT NULL DEFAULT 'waiting',
    alerts_enabled          BOOLEAN        NOT NULL DEFAULT TRUE,
    max_alerts_per_incident INTEGER        NOT NULL DEFAULT 3,
    consecutive_failures    INTEGER        NOT NULL DEFAULT 0,
    last_checked_at         TIMESTAMPTZ,
    next_check_at           TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    created_at              TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE TABLE uptime_checks (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    monitor_id       UUID        NOT NULL REFERENCES uptime_monitors(id) ON DELETE CASCADE,
    checked_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status_code      INTEGER,
    response_time_ms INTEGER     NOT NULL,
    is_up            BOOLEAN     NOT NULL
);

CREATE TABLE uptime_incidents (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    monitor_id  UUID        NOT NULL REFERENCES uptime_monitors(id) ON DELETE CASCADE,
    started_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    alert_count INTEGER     NOT NULL DEFAULT 0
);

CREATE INDEX idx_uptime_monitors_org_id      ON uptime_monitors(org_id);
CREATE INDEX idx_uptime_monitors_next_check  ON uptime_monitors(next_check_at) WHERE status != 'paused';
CREATE INDEX idx_uptime_checks_monitor       ON uptime_checks(monitor_id, checked_at DESC);
CREATE INDEX idx_uptime_incidents_monitor_id ON uptime_incidents(monitor_id);

-- +goose Down

DROP INDEX IF EXISTS idx_uptime_incidents_monitor_id;
DROP INDEX IF EXISTS idx_uptime_checks_monitor;
DROP INDEX IF EXISTS idx_uptime_monitors_next_check;
DROP INDEX IF EXISTS idx_uptime_monitors_org_id;
DROP TABLE IF EXISTS uptime_incidents;
DROP TABLE IF EXISTS uptime_checks;
DROP TABLE IF EXISTS uptime_monitors;
