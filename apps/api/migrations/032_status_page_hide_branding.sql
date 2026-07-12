-- +goose Up

-- ADR-035 — per-page toggle to remove the "Powered by Checkmeup" footer and
-- FAQ/Terms/Privacy links; gated to paid plans at the handler layer, not here.
ALTER TABLE status_pages ADD COLUMN hide_branding BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE status_pages DROP COLUMN hide_branding;
