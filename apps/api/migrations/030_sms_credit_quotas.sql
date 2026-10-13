-- +goose Up

-- ADR billing (simplified for this pass — flat 1 credit per SMS send, no
-- destination weighting yet; see .about/adr/billing.md's
-- "Simplification" note). Reset is lazy — checked and applied at send-time
-- (ConsumeSMSCredit) rather than via a scheduled job, consistent with
-- ADR workers's no-broker/no-extra-scheduler stance.
ALTER TABLE orgs ADD COLUMN sms_credits_used_this_month INT NOT NULL DEFAULT 0;
ALTER TABLE orgs ADD COLUMN sms_credits_reset_at DATE NOT NULL DEFAULT (date_trunc('month', now()) + interval '1 month')::date;

-- +goose Down
ALTER TABLE orgs DROP COLUMN sms_credits_reset_at;
ALTER TABLE orgs DROP COLUMN sms_credits_used_this_month;
