-- +goose Up

CREATE TYPE port_expected_state AS ENUM ('open', 'closed');

CREATE TABLE port_monitors (
    id                      UUID                PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID                NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name                    TEXT                NOT NULL,
    host                    TEXT                NOT NULL,
    port                    INTEGER             NOT NULL,
    expected_state          port_expected_state NOT NULL DEFAULT 'open',
    interval_mins           INTEGER             NOT NULL DEFAULT 10,
    status                  monitor_status      NOT NULL DEFAULT 'waiting',
    alerts_enabled          BOOLEAN             NOT NULL DEFAULT TRUE,
    max_alerts_per_incident INTEGER             NOT NULL DEFAULT 3,
    alert_after_n_failures  INTEGER             NOT NULL DEFAULT 0,
    consecutive_failures    INTEGER             NOT NULL DEFAULT 0,
    last_checked_at         TIMESTAMPTZ,
    next_check_at           TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    created_at              TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ         NOT NULL DEFAULT NOW()
);

CREATE TABLE port_checks (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    monitor_id       UUID        NOT NULL REFERENCES port_monitors(id) ON DELETE CASCADE,
    checked_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    response_time_ms INTEGER     NOT NULL,
    is_up            BOOLEAN     NOT NULL,
    failure_reason   TEXT
);

CREATE TABLE port_incidents (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    monitor_id  UUID        NOT NULL REFERENCES port_monitors(id) ON DELETE CASCADE,
    started_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    alert_count INTEGER     NOT NULL DEFAULT 0
);

CREATE INDEX idx_port_monitors_org_id      ON port_monitors(org_id);
CREATE INDEX idx_port_monitors_next_check  ON port_monitors(next_check_at) WHERE status != 'paused';
CREATE INDEX idx_port_checks_monitor       ON port_checks(monitor_id, checked_at DESC);
CREATE INDEX idx_port_incidents_monitor_id ON port_incidents(monitor_id);

-- +goose Down

DROP INDEX IF EXISTS idx_port_incidents_monitor_id;
DROP INDEX IF EXISTS idx_port_checks_monitor;
DROP INDEX IF EXISTS idx_port_monitors_next_check;
DROP INDEX IF EXISTS idx_port_monitors_org_id;
DROP TABLE IF EXISTS port_incidents;
DROP TABLE IF EXISTS port_checks;
DROP TABLE IF EXISTS port_monitors;
DROP TYPE IF EXISTS port_expected_state;
