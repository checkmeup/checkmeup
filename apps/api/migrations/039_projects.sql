-- +goose Up

-- ADR projects: a project groups an org's monitors, typically one per client.
-- Names are unique per org regardless of case, so "Acme" and "acme" can't
-- both exist as two groups the user can't tell apart.
CREATE TABLE projects (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id     UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (id, org_id)
);

CREATE UNIQUE INDEX idx_projects_org_name ON projects(org_id, lower(name));

-- Each monitor is in at most one project. The foreign key covers org_id too,
-- so the database itself refuses a monitor in another org's project; deleting
-- a project nulls only project_id and leaves the monitor running, ungrouped.
ALTER TABLE cron_monitors   ADD COLUMN project_id UUID;
ALTER TABLE uptime_monitors ADD COLUMN project_id UUID;
ALTER TABLE ssl_monitors    ADD COLUMN project_id UUID;
ALTER TABLE domain_monitors ADD COLUMN project_id UUID;
ALTER TABLE port_monitors   ADD COLUMN project_id UUID;
ALTER TABLE dns_monitors    ADD COLUMN project_id UUID;

ALTER TABLE cron_monitors   ADD CONSTRAINT cron_monitors_project_fk   FOREIGN KEY (project_id, org_id) REFERENCES projects(id, org_id) ON DELETE SET NULL (project_id);
ALTER TABLE uptime_monitors ADD CONSTRAINT uptime_monitors_project_fk FOREIGN KEY (project_id, org_id) REFERENCES projects(id, org_id) ON DELETE SET NULL (project_id);
ALTER TABLE ssl_monitors    ADD CONSTRAINT ssl_monitors_project_fk    FOREIGN KEY (project_id, org_id) REFERENCES projects(id, org_id) ON DELETE SET NULL (project_id);
ALTER TABLE domain_monitors ADD CONSTRAINT domain_monitors_project_fk FOREIGN KEY (project_id, org_id) REFERENCES projects(id, org_id) ON DELETE SET NULL (project_id);
ALTER TABLE port_monitors   ADD CONSTRAINT port_monitors_project_fk   FOREIGN KEY (project_id, org_id) REFERENCES projects(id, org_id) ON DELETE SET NULL (project_id);
ALTER TABLE dns_monitors    ADD CONSTRAINT dns_monitors_project_fk    FOREIGN KEY (project_id, org_id) REFERENCES projects(id, org_id) ON DELETE SET NULL (project_id);

CREATE INDEX idx_cron_monitors_project   ON cron_monitors(org_id, project_id)   WHERE project_id IS NOT NULL;
CREATE INDEX idx_uptime_monitors_project ON uptime_monitors(org_id, project_id) WHERE project_id IS NOT NULL;
CREATE INDEX idx_ssl_monitors_project    ON ssl_monitors(org_id, project_id)    WHERE project_id IS NOT NULL;
CREATE INDEX idx_domain_monitors_project ON domain_monitors(org_id, project_id) WHERE project_id IS NOT NULL;
CREATE INDEX idx_port_monitors_project   ON port_monitors(org_id, project_id)   WHERE project_id IS NOT NULL;
CREATE INDEX idx_dns_monitors_project    ON dns_monitors(org_id, project_id)    WHERE project_id IS NOT NULL;

-- +goose Down

ALTER TABLE dns_monitors    DROP COLUMN project_id;
ALTER TABLE port_monitors   DROP COLUMN project_id;
ALTER TABLE domain_monitors DROP COLUMN project_id;
ALTER TABLE ssl_monitors    DROP COLUMN project_id;
ALTER TABLE uptime_monitors DROP COLUMN project_id;
ALTER TABLE cron_monitors   DROP COLUMN project_id;
DROP TABLE IF EXISTS projects;
