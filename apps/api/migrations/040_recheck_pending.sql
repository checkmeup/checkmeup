-- +goose Up

-- ADR alerting: an SSL or domain check that errors right after a working one
-- is rechecked about 30 seconds later before the monitor shows Error. These
-- monitors keep no per-check history to tell a recheck from a first error,
-- so the pending recheck is a flag on the monitor.
ALTER TABLE ssl_monitors ADD COLUMN recheck_pending BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE domain_monitors ADD COLUMN recheck_pending BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE domain_monitors DROP COLUMN recheck_pending;
ALTER TABLE ssl_monitors DROP COLUMN recheck_pending;
