-- +goose Up

-- Only the two types with working delivery today get an enum value now —
-- webhook/Slack/Teams/etc. each add their own value via their own migration
-- when that epic actually ships (EP-28, ADR-023), not reserved upfront.
CREATE TYPE notification_channel_type AS ENUM ('telegram', 'email');

CREATE TABLE notification_channels (
    id          UUID                       PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID                       NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    type        notification_channel_type  NOT NULL,
    name        TEXT                       NOT NULL,
    config      JSONB                      NOT NULL, -- {"chatId": "..."} / {"email": "..."}
    enabled     BOOLEAN                    NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ                NOT NULL DEFAULT NOW()
);

-- monitor_type is 'cron', 'uptime', or 'ssl'; no FK constraint, same pattern as
-- maintenance_window_monitors / status_page_monitors
CREATE TABLE monitor_notification_channels (
    id            UUID NOT NULL DEFAULT gen_random_uuid(),
    channel_id    UUID NOT NULL REFERENCES notification_channels(id) ON DELETE CASCADE,
    monitor_type  TEXT NOT NULL,
    monitor_id    UUID NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (channel_id, monitor_type, monitor_id)
);

CREATE INDEX idx_notification_channels_org      ON notification_channels(org_id);
CREATE INDEX idx_monitor_notification_channels  ON monitor_notification_channels(monitor_type, monitor_id);

-- ─── Backfill (US-2804) ────────────────────────────────────────────────────
-- Migrate existing org-level telegram_chat_id / alert_email into real
-- channel rows, then attach those channels to every monitor that currently
-- has alerts_enabled = true, so dispatch behavior is unchanged at cutover.

INSERT INTO notification_channels (org_id, type, name, config, enabled)
SELECT id, 'telegram', 'Telegram', jsonb_build_object('chatId', telegram_chat_id), true
FROM orgs
WHERE telegram_chat_id IS NOT NULL AND telegram_chat_id <> '';

INSERT INTO notification_channels (org_id, type, name, config, enabled)
SELECT id, 'email', 'Email', jsonb_build_object('email', alert_email), true
FROM orgs
WHERE alert_email IS NOT NULL AND alert_email <> '' AND email_alerts_enabled;

INSERT INTO monitor_notification_channels (channel_id, monitor_type, monitor_id)
SELECT nc.id, 'cron', cm.id
FROM notification_channels nc
JOIN cron_monitors cm ON cm.org_id = nc.org_id
WHERE cm.alerts_enabled;

INSERT INTO monitor_notification_channels (channel_id, monitor_type, monitor_id)
SELECT nc.id, 'uptime', um.id
FROM notification_channels nc
JOIN uptime_monitors um ON um.org_id = nc.org_id
WHERE um.alerts_enabled;

INSERT INTO monitor_notification_channels (channel_id, monitor_type, monitor_id)
SELECT nc.id, 'ssl', sm.id
FROM notification_channels nc
JOIN ssl_monitors sm ON sm.org_id = nc.org_id
WHERE sm.alerts_enabled;

-- Note: orgs.telegram_chat_id / alert_email / email_alerts_enabled are
-- intentionally NOT dropped here — kept until the channel-based dispatch
-- path (worker.go) has run in production, per ADR-023's migration path.

-- +goose Down

DROP TABLE monitor_notification_channels;
DROP TABLE notification_channels;
DROP TYPE notification_channel_type;
