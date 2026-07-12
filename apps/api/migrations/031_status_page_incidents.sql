-- +goose Up

CREATE TYPE incident_severity AS ENUM ('minor', 'major', 'critical');
CREATE TYPE incident_status   AS ENUM ('investigating', 'identified', 'monitoring', 'resolved');

CREATE TABLE status_page_incidents (
    id          UUID              PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID              NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    title       TEXT              NOT NULL,
    severity    incident_severity NOT NULL,
    status      incident_status   NOT NULL DEFAULT 'investigating',
    created_at  TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ       NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

-- monitor_type is 'cron', 'uptime', 'ssl', 'domain', or 'port'; no FK constraint
-- (polymorphic, same pattern as maintenance_window_monitors)
CREATE TABLE status_page_incident_monitors (
    id           UUID NOT NULL DEFAULT gen_random_uuid(),
    incident_id  UUID NOT NULL REFERENCES status_page_incidents(id) ON DELETE CASCADE,
    monitor_type TEXT NOT NULL,
    monitor_id   UUID NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (incident_id, monitor_type, monitor_id)
);

-- append-only; the incident's own `status` column is a denormalized cache of
-- the latest update's status, kept in sync by the handler on every insert
CREATE TABLE status_page_incident_updates (
    id          UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id UUID            NOT NULL REFERENCES status_page_incidents(id) ON DELETE CASCADE,
    message     TEXT            NOT NULL,
    status      incident_status NOT NULL,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_status_page_incidents_org_id          ON status_page_incidents(org_id);
CREATE INDEX idx_status_page_incident_monitors         ON status_page_incident_monitors(monitor_type, monitor_id);
CREATE INDEX idx_status_page_incident_updates_incident ON status_page_incident_updates(incident_id);

-- +goose Down

DROP INDEX IF EXISTS idx_status_page_incident_updates_incident;
DROP INDEX IF EXISTS idx_status_page_incident_monitors;
DROP INDEX IF EXISTS idx_status_page_incidents_org_id;
DROP TABLE IF EXISTS status_page_incident_updates;
DROP TABLE IF EXISTS status_page_incident_monitors;
DROP TABLE IF EXISTS status_page_incidents;
DROP TYPE IF EXISTS incident_status;
DROP TYPE IF EXISTS incident_severity;
