# Slack alerts

## Problem

Teams that live in Slack want alerts in a channel, readable as messages rather than raw JSON.

## Users

Teams using Slack.

## Success signal

A user pastes an Incoming Webhook URL and the next alert appears in the channel as a readable message.

## Scope

- Slack notification channels from Incoming Webhook URLs, several per org.
- Down and recovery messages formatted for Slack.
- Delivery results shown in Settings, without retries.

## Non-goals

- A Slack app with OAuth, slash commands or acknowledging from Slack.

## Acceptance criteria

- Only a `hooks.slack.com` Incoming Webhook URL is accepted.
  → test: apps/web/src/lib/notificationChannelTypes.test.ts "requires slack URLs to point at hooks.slack.com"
  → test: apps/web/src/components/NotificationChannelsCard.test.ts "rejects a Slack URL that is not an Incoming Webhook URL"
- After creation the API returns the URL masked (scheme, host and its last 4 characters); saving an edit with the URL left empty keeps the stored one (ADR alerting).
  → test: apps/api/internal/handler/notification_channels_test.go "TestMaskChannelURL"
  → test: apps/api/internal/handler/notification_channels_test.go "TestUpdateNotificationChannel/an empty or masked URL keeps the stored one (ADR alerting)"
  → test: apps/web/src/lib/notificationChannelTypes.test.ts "allows an empty URL when the channel already has a stored one"
- A down message shows the monitor's name, type, reason and time; a recovery message shows the downtime.
  → test: apps/api/internal/slack/slack_test.go "TestDownMessage_ContainsMonitorDetails"
  → test: apps/api/internal/slack/slack_test.go "TestRecoveryMessage_ContainsDowntime"
- A delivery's success or failure (such as a removed webhook's `404`) is recorded on the channel and shown in Settings.
  → test: apps/api/internal/worker/worker_test.go "TestDispatchAlert/slack channel delivers and records success on the channel row"
  → test: apps/api/internal/worker/worker_test.go "TestDispatchAlert/slack channel records failure on a non-2xx response"
- "Send test message" reports whether Slack accepted it.
  → test: apps/api/internal/handler/notification_channels_test.go "TestTestNotificationChannel/slack test message reports whether Slack accepted it"
