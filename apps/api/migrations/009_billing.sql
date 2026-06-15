-- +goose Up

ALTER TABLE orgs
    ADD COLUMN ls_customer_id     TEXT,
    ADD COLUMN ls_subscription_id TEXT,
    ADD COLUMN subscription_status TEXT NOT NULL DEFAULT 'free',
    ADD COLUMN plan_renews_at     TIMESTAMPTZ;

-- +goose Down

ALTER TABLE orgs
    DROP COLUMN IF EXISTS ls_customer_id,
    DROP COLUMN IF EXISTS ls_subscription_id,
    DROP COLUMN IF EXISTS subscription_status,
    DROP COLUMN IF EXISTS plan_renews_at;
