# SMS alerts

## Problem

When a monitor goes down at night, a text message wakes people up where a chat notification doesn't, and it needs no app.

## Users

Users on paid plans who want critical alerts by text. Not Hobby: SMS costs money per message (ADR billing).

## Success signal

A user adds a number with consent and gets the next down alert as one text, without the org's SMS spend ever going past its plan's quota.

## Scope

- SMS notification channels: E.164 numbers, several per org, each saved only with recorded consent (ADR alerting).
- Down and recovery texts kept within one 160-character segment.
- A test SMS before relying on it, rate limited separately from alerts.
- Monthly SMS credits per plan, reset on the 1st; running out skips SMS and falls back as for any failed channel.
- Remaining credits on the billing page.

## Non-goals

- Buying extra credits or paying per message.
- Weighting credits by destination (ADR billing).
- Two-way SMS or replying to acknowledge.

## Acceptance criteria

- A number not in E.164 format is refused.
  → test: apps/api/internal/handler/notification_channels_test.go "TestCreateNotificationChannel/sms phone_number must be E.164"
  → test: apps/web/src/components/NotificationChannelsCard.test.ts "rejects a phone number that is not E.164"
- A number can't be saved or tested until the consent box is checked; the consent time is stamped by the server, never taken from the client.
  → test: apps/web/src/components/NotificationChannelsCard.test.ts "disables save and test until the sms consent checkbox is checked"
  → test: apps/api/internal/handler/notification_channels_test.go "TestCreateNotificationChannel/creates an sms channel with server-stamped consent_at, stripping the client's consent flag"
  → test: apps/api/internal/handler/notification_channels_test.go "TestCreateNotificationChannel/a client-supplied consent_at is ignored and overwritten with a server-stamped value"
- Changing a channel's number needs fresh consent; keeping the number keeps the consent on file.
  → test: apps/api/internal/handler/notification_channels_test.go "TestUpdateNotificationChannel/changing an sms channel's number requires fresh consent"
  → test: apps/api/internal/handler/notification_channels_test.go "TestUpdateNotificationChannel/updating an sms channel with the same number carries consent forward without re-requiring it"
- On Hobby the SMS option is hidden and the API refuses SMS channels and test sends.
  → test: apps/web/src/components/NotificationChannelsCard.test.ts "hides the SMS option from the type picker on the Hobby plan"
  → test: apps/api/internal/handler/notification_channels_test.go "TestCreateNotificationChannel/sms blocked on the Hobby plan (interim guard ahead of ADR billing credit quotas)"
  → test: apps/api/internal/handler/notification_channels_test.go "TestTestNotificationChannel/sms test-send blocked on the Hobby plan (same guard as create)"
- Test SMS sends are limited to 10 a minute, 10 an hour and 20 a day per org, and don't use credits.
  → test: apps/api/internal/handler/notification_channels_test.go "TestTestNotificationChannel/sms test-send is rate limited to 10 per minute per org"
  → test: apps/api/internal/handler/notification_channels_test.go "TestTestNotificationChannel/sms test-send is capped at 20 per day per org"
  → test: apps/api/internal/handler/notification_channels_test.go "TestTestNotificationChannel/sms test-send rate limit is scoped per org, not global"
- An alert text longer than 160 characters is cut with an ellipsis, by character, not by byte.
  → test: apps/api/internal/worker/worker_test.go "TestTruncateSMS/message over 160 chars is truncated with an ellipsis"
  → test: apps/api/internal/worker/worker_test.go "TestTruncateSMS/multi-byte characters are cut by rune, not by byte"
- Each alert sent uses one credit; with no credits left the SMS is skipped and recorded as a delivery failure, without reaching Twilio.
  → test: apps/api/internal/worker/worker_test.go "TestDispatchAlert/sms send within the monthly quota consumes exactly one credit"
  → test: apps/api/internal/worker/worker_test.go "TestDispatchAlert/sms channel on the Hobby plan is skipped for quota exhaustion without ever reaching Twilio"
  → test: apps/api/internal/worker/worker_test.go "TestConsumeSMSCredit/blocks once the plan limit is reached, without going negative or over"
- Credits reset on the 1st of the month, and the billing page shows zero used once the reset date has passed.
  → test: apps/api/internal/worker/worker_test.go "TestConsumeSMSCredit/lazily resets once the reset date has passed"
  → test: apps/api/internal/handler/billing_test.go "TestGetBillingInfo/displays 0 used once the reset date has passed, even before any send applies it"
- When SMS was the only channel, or every channel failed, the alert goes to the org's users by email.
  → test: apps/api/internal/worker/worker_test.go "TestDispatchAlert/every attached channel failing falls back to org user email (generalized ADR billing fallback)"
- An SMS Twilio accepted but the carrier failed is recorded as failed once Twilio reports it, and when it was the alert's only delivery the org's users get the alert by email.
  → test: apps/api/internal/worker/worker_test.go "TestHandleSMSStatus/undelivered records the failure and emails the org when the sms was the only delivery"
  → test: apps/api/internal/worker/worker_test.go "TestDispatchAlert/an accepted sms keeps the fallback email only when it is the alert's only delivery (ADR alerting)"
  → test: apps/api/internal/handler/sms_status_test.go "TestHandleSMSStatus/a wrong token is a 404 and changes nothing"
- A Twilio rejection is recorded with Twilio's error message and shown in Settings; failures aren't retried.
  → test: apps/api/internal/twilio/twilio_test.go "TestSend/non-2xx response surfaces Twilio's error message"
  → test: apps/api/internal/worker/worker_test.go "TestDispatchAlert/sms channel attempts a send and records failure when Twilio isn't configured"
