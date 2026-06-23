-- +goose Up

CREATE TYPE domain_monitor_status AS ENUM ('waiting', 'up', 'expiring_soon', 'expired', 'error', 'paused');

CREATE TABLE domain_monitors (
    id              UUID                  PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID                  NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name            TEXT                  NOT NULL,
    domain          TEXT                  NOT NULL,
    status          domain_monitor_status NOT NULL DEFAULT 'waiting',
    alerts_enabled  BOOLEAN               NOT NULL DEFAULT TRUE,
    expires_at      TIMESTAMPTZ,
    registrar       TEXT,
    error_msg       TEXT,
    alerted_30d     BOOLEAN               NOT NULL DEFAULT FALSE,
    alerted_14d     BOOLEAN               NOT NULL DEFAULT FALSE,
    alerted_7d      BOOLEAN               NOT NULL DEFAULT FALSE,
    last_checked_at TIMESTAMPTZ,
    next_check_at   TIMESTAMPTZ           NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ           NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ           NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_domain_monitors_org_id     ON domain_monitors(org_id);
CREATE INDEX idx_domain_monitors_next_check ON domain_monitors(next_check_at) WHERE status != 'paused';

-- +goose Down

DROP INDEX IF EXISTS idx_domain_monitors_next_check;
DROP INDEX IF EXISTS idx_domain_monitors_org_id;
DROP TABLE IF EXISTS domain_monitors;
DROP TYPE IF EXISTS domain_monitor_status;
