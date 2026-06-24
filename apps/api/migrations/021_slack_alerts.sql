-- +goose Up

-- EP-17: slack becomes a real, deliverable notification_channel_type.
-- Slack Incoming Webhooks are plain HTTPS POSTs, so the config shape is
-- {"url": "https://hooks.slack.com/services/..."} — same key as webhook (EP-14)
-- but validated against the hooks.slack.com domain pattern.
ALTER TYPE notification_channel_type ADD VALUE 'slack';

-- +goose Down

-- Postgres can't remove an enum value, so the down migration can't fully
-- reverse the Up — matches goose's documented limitation for ADD VALUE.
