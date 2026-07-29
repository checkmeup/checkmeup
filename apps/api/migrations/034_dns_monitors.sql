-- +goose Up

-- EP-39: DNS record monitoring. expected_value is either user-pinned at
-- creation or left NULL and auto-captured from the first successful lookup
-- (baseline mode) — both modes then compare identically on every later
-- check. dns_checks.failure_reason is set only for a lookup error
-- (NXDOMAIN/SERVFAIL/timeout); it stays NULL for a value mismatch, so the
-- two failure kinds stay distinguishable (US-3902).
CREATE TYPE dns_record_type AS ENUM ('A', 'AAAA', 'CNAME', 'MX', 'TXT', 'NS');

CREATE TABLE dns_monitors (
    id                      UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id                  UUID             NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name                    TEXT             NOT NULL,
    hostname                TEXT             NOT NULL,
    record_type             dns_record_type  NOT NULL,
    expected_value          TEXT,
    baseline_captured       BOOLEAN          NOT NULL DEFAULT FALSE,
    last_resolved_value     TEXT,
    interval_mins           INTEGER          NOT NULL DEFAULT 10,
    status                  monitor_status   NOT NULL DEFAULT 'waiting',
    alerts_enabled          BOOLEAN          NOT NULL DEFAULT TRUE,
    max_alerts_per_incident INTEGER          NOT NULL DEFAULT 3,
    alert_after_n_failures  INTEGER          NOT NULL DEFAULT 0,
    consecutive_failures    INTEGER          NOT NULL DEFAULT 0,
    last_checked_at         TIMESTAMPTZ,
    next_check_at           TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    created_at              TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE TABLE dns_checks (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    monitor_id       UUID        NOT NULL REFERENCES dns_monitors(id) ON DELETE CASCADE,
    checked_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    response_time_ms INTEGER     NOT NULL,
    is_up            BOOLEAN     NOT NULL,
    resolved_value   TEXT,
    failure_reason   TEXT
);

CREATE TABLE dns_incidents (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    monitor_id  UUID        NOT NULL REFERENCES dns_monitors(id) ON DELETE CASCADE,
    started_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    alert_count INTEGER     NOT NULL DEFAULT 0
);

CREATE INDEX idx_dns_monitors_org_id      ON dns_monitors(org_id);
CREATE INDEX idx_dns_monitors_next_check  ON dns_monitors(next_check_at) WHERE status != 'paused';
CREATE INDEX idx_dns_checks_monitor       ON dns_checks(monitor_id, checked_at DESC);
CREATE INDEX idx_dns_incidents_monitor_id ON dns_incidents(monitor_id);

-- +goose Down

DROP INDEX IF EXISTS idx_dns_incidents_monitor_id;
DROP INDEX IF EXISTS idx_dns_checks_monitor;
DROP INDEX IF EXISTS idx_dns_monitors_next_check;
DROP INDEX IF EXISTS idx_dns_monitors_org_id;
DROP TABLE IF EXISTS dns_incidents;
DROP TABLE IF EXISTS dns_checks;
DROP TABLE IF EXISTS dns_monitors;
DROP TYPE IF EXISTS dns_record_type;
