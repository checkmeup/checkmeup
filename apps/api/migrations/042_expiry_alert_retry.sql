-- +goose Up

-- PRD alert-delivery: alert_attempts caps how many times an SMS channel is
-- tried for an SSL or domain alert sent again (043), since each try uses a
-- credit. alerted_expired marks a domain's expired alert as sent: the 7-day
-- warning shares alerted_7d with it, and a lookup error leaves the status
-- Error. Domains already expired have had theirs.
ALTER TABLE ssl_monitors ADD COLUMN alert_attempts INTEGER NOT NULL DEFAULT 0;
ALTER TABLE domain_monitors ADD COLUMN alert_attempts INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN alerted_expired BOOLEAN NOT NULL DEFAULT false;
UPDATE domain_monitors SET alerted_expired = true WHERE status = 'expired';

-- +goose Down
ALTER TABLE domain_monitors DROP COLUMN alerted_expired, DROP COLUMN alert_attempts;
ALTER TABLE ssl_monitors DROP COLUMN alert_attempts;
