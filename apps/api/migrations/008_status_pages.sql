-- +goose Up

CREATE TABLE status_pages (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    slug        TEXT        NOT NULL UNIQUE,
    title       TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    logo_url    TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- monitor_type is 'cron', 'uptime', or 'ssl'; no FK constraint (polymorphic)
CREATE TABLE status_page_monitors (
    id            UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id       UUID    NOT NULL REFERENCES status_pages(id) ON DELETE CASCADE,
    monitor_type  TEXT    NOT NULL,
    monitor_id    UUID    NOT NULL,
    display_name  TEXT    NOT NULL,
    display_order INTEGER NOT NULL DEFAULT 0,
    UNIQUE (page_id, monitor_type, monitor_id)
);

CREATE INDEX idx_status_pages_org_id   ON status_pages(org_id);
CREATE INDEX idx_status_pages_slug     ON status_pages(slug);
CREATE INDEX idx_status_page_monitors  ON status_page_monitors(page_id, display_order);

-- +goose Down

DROP INDEX IF EXISTS idx_status_page_monitors;
DROP INDEX IF EXISTS idx_status_pages_slug;
DROP INDEX IF EXISTS idx_status_pages_org_id;
DROP TABLE IF EXISTS status_page_monitors;
DROP TABLE IF EXISTS status_pages;
