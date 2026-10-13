# Notification channels

## Problem

Alerts went to one Telegram chat and one email per org, for every monitor; users need several destinations and to route each monitor to the right ones.

## Users

Every org with alerts; the foundation the per-channel PRDs build on (telegram-alerts, email-alerts, webhook-alerts, slack-alerts, sms-alerts).

## Success signal

A user routes a client's monitors to that client's Slack and their own Telegram without the two seeing each other's alerts.

## Scope

- Notification channels per org: add, edit, enable or disable, delete, test (ADR alerting); a per-plan limit on how many.
- Attaching channels to each monitor; new monitors get all enabled channels.
- The email fallback when a monitor has nowhere else to alert.
- One alert counts once against `max_alerts_per_incident`, however many channels carry it; recovery alerts are never capped.

## Non-goals

- Channel groups (ADR alerting).
- Escalation policies or on-call rotations.

## Acceptance criteria

- Settings lists each channel with its type, value, enabled state and last delivery result.
  → test: apps/web/src/components/NotificationChannelsCard.test.ts "renders channel rows with type, value, and delivery summary"
- A channel needs a name and its type's value; an unknown type is refused.
  → test: apps/api/internal/handler/notification_channels_test.go "TestCreateNotificationChannel/missing name"
  → test: apps/api/internal/handler/notification_channels_test.go "TestCreateNotificationChannel/unsupported type"
  → test: apps/web/src/lib/notificationChannelTypes.test.ts "requires a value"
- A new channel is enabled; it can be disabled and re-enabled from the list.
  → test: apps/api/internal/handler/notification_channels_test.go "TestCreateNotificationChannel/creates a telegram channel, enabled by default"
  → test: apps/web/src/components/NotificationChannelsCard.test.ts "toggles a channel's enabled state"
- Re-enabling a channel past the plan's limit is refused with an upgrade prompt.
  → test: apps/api/internal/handler/notification_channels_test.go "TestUpdateNotificationChannel/re-enabling a disabled channel is blocked once enabled channels are already at the plan limit (ADR billing)"
  → test: apps/web/src/components/NotificationChannelsCard.test.ts "shows an upgrade prompt instead of failing silently when re-enabling a channel is blocked by the plan limit"
- A monitor with channels attached but alerts turned off sends nothing.
  → test: apps/api/internal/worker/worker_test.go "TestCheckOverdue/an incident still opens when alerts are disabled, but no alert is sent"
- A monitor with no channel attached emails every user in the org; with no user, nothing is sent and nothing breaks.
  → test: apps/api/internal/worker/worker_test.go "TestDispatchAlert/no channel attached falls back to emailing every org user (ADR alerting)"
  → test: apps/api/internal/worker/worker_test.go "TestDispatchAlert/no channel attached and no org user falls back to nothing"
- After the cap is reached, further down alerts for the incident are suppressed.
  → test: apps/api/internal/worker/worker_test.go "TestBuildSSLCheckResult/the per-incident alert cap suppresses further alerts"
- Creating a channel past the plan's limit gets `402 plan_limit_reached`.
- Each monitor's form has a channel picker; a new monitor starts with all the org's enabled channels selected.
- One alert sent to several channels counts once against `max_alerts_per_incident`.
