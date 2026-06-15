-- +goose Up

CREATE TYPE ssl_monitor_status AS ENUM ('waiting', 'up', 'expiring_soon', 'expired', 'error', 'paused');

CREATE TABLE ssl_monitors (
    id              UUID               PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID               NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name            TEXT               NOT NULL,
    hostname        TEXT               NOT NULL,
    status          ssl_monitor_status NOT NULL DEFAULT 'waiting',
    alerts_enabled  BOOLEAN            NOT NULL DEFAULT TRUE,
    expires_at      TIMESTAMPTZ,
    issuer          TEXT,
    error_msg       TEXT,
    alerted_30d     BOOLEAN            NOT NULL DEFAULT FALSE,
    alerted_14d     BOOLEAN            NOT NULL DEFAULT FALSE,
    alerted_7d      BOOLEAN            NOT NULL DEFAULT FALSE,
    last_checked_at TIMESTAMPTZ,
    next_check_at   TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ        NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ssl_monitors_org_id     ON ssl_monitors(org_id);
CREATE INDEX idx_ssl_monitors_next_check ON ssl_monitors(next_check_at) WHERE status != 'paused';

-- +goose Down

DROP INDEX IF EXISTS idx_ssl_monitors_next_check;
DROP INDEX IF EXISTS idx_ssl_monitors_org_id;
DROP TABLE IF EXISTS ssl_monitors;
DROP TYPE IF EXISTS ssl_monitor_status;
