-- +goose Up

-- ADR alerting: an SMS Twilio accepted can still fail at the carrier. Each
-- alert SMS is kept here by its Twilio SID until its status callback reports a
-- final status, so the failure lands on the channel and, when the SMS was the
-- alert's only delivery, the fallback email goes out with the stored text.
CREATE TABLE sms_messages (
    sid              TEXT        PRIMARY KEY,
    org_id           UUID        NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    channel_id       UUID        NOT NULL REFERENCES notification_channels(id) ON DELETE CASCADE,
    fallback_subject TEXT,
    fallback_html    TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX sms_messages_created_at_idx ON sms_messages (created_at);

-- +goose Down
DROP TABLE sms_messages;
