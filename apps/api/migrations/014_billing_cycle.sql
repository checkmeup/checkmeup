-- +goose Up

ALTER TABLE orgs ADD COLUMN billing_cycle TEXT NOT NULL DEFAULT 'monthly';

-- +goose Down

ALTER TABLE orgs DROP COLUMN billing_cycle;
