-- +goose Up

CREATE TYPE monitor_status AS ENUM ('waiting', 'up', 'down', 'paused');

ALTER TABLE orgs ADD COLUMN telegram_chat_id TEXT;

CREATE TABLE cron_monitors (
    id                UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id            UUID           NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name              TEXT           NOT NULL,
    schedule          TEXT           NOT NULL,
    grace_period_mins INTEGER        NOT NULL DEFAULT 5,
    ping_token        TEXT           UNIQUE NOT NULL,
    status            monitor_status NOT NULL DEFAULT 'waiting',
    alerts_enabled    BOOLEAN        NOT NULL DEFAULT TRUE,
    last_ping_at      TIMESTAMPTZ,
    next_ping_at      TIMESTAMPTZ,
    created_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE TABLE cron_pings (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    monitor_id  UUID        NOT NULL REFERENCES cron_monitors(id) ON DELETE CASCADE,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    source_ip   TEXT        NOT NULL DEFAULT ''
);

CREATE TABLE cron_incidents (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    monitor_id  UUID        NOT NULL REFERENCES cron_monitors(id) ON DELETE CASCADE,
    started_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE INDEX idx_cron_monitors_org_id      ON cron_monitors(org_id);
CREATE INDEX idx_cron_monitors_ping_token  ON cron_monitors(ping_token);
CREATE INDEX idx_cron_monitors_next_ping   ON cron_monitors(next_ping_at) WHERE status = 'up';
CREATE INDEX idx_cron_pings_monitor_id     ON cron_pings(monitor_id);
CREATE INDEX idx_cron_incidents_monitor_id ON cron_incidents(monitor_id);

-- +goose Down

DROP INDEX IF EXISTS idx_cron_incidents_monitor_id;
DROP INDEX IF EXISTS idx_cron_pings_monitor_id;
DROP INDEX IF EXISTS idx_cron_monitors_next_ping;
DROP INDEX IF EXISTS idx_cron_monitors_ping_token;
DROP INDEX IF EXISTS idx_cron_monitors_org_id;
DROP TABLE IF EXISTS cron_incidents;
DROP TABLE IF EXISTS cron_pings;
DROP TABLE IF EXISTS cron_monitors;
ALTER TABLE orgs DROP COLUMN IF EXISTS telegram_chat_id;
DROP TYPE IF EXISTS monitor_status;
