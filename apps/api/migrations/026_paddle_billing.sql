-- +goose Up

ALTER TABLE orgs
    DROP COLUMN IF EXISTS ls_customer_id,
    DROP COLUMN IF EXISTS ls_subscription_id,
    ADD COLUMN paddle_customer_id     TEXT,
    ADD COLUMN paddle_subscription_id TEXT;

-- +goose Down

ALTER TABLE orgs
    DROP COLUMN IF EXISTS paddle_customer_id,
    DROP COLUMN IF EXISTS paddle_subscription_id,
    ADD COLUMN ls_customer_id     TEXT,
    ADD COLUMN ls_subscription_id TEXT;
