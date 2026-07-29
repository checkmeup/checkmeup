-- +goose Up

-- ADR-038 — per-page layout choice between the original single-column
-- design and a wider monitor-grid + incident-sidebar layout.
ALTER TABLE status_pages ADD COLUMN layout TEXT NOT NULL DEFAULT 'classic';
ALTER TABLE status_pages ADD CONSTRAINT status_pages_layout_check CHECK (layout IN ('classic', 'grid'));

-- +goose Down
ALTER TABLE status_pages DROP CONSTRAINT status_pages_layout_check;
ALTER TABLE status_pages DROP COLUMN layout;
