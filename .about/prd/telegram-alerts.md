# Telegram alerts

## Problem

An alert is only useful where the user will see it at once; many developers live in Telegram rather than their inbox.

## Users

Anyone who wants down and recovery alerts in a Telegram chat.

## Success signal

A user connects a chat without leaving the settings page and gets the next down alert there within one check cycle.

## Scope

- Connecting a chat as a notification channel (PRD notification-channels): the bot answers `/start` with the chat ID to paste.
- A test message before relying on it.
- Down and recovery alerts for every monitor type.
- Muting alerts per monitor.

## Non-goals

- Acknowledging or silencing an incident from inside Telegram.

## Acceptance criteria

- Sending `/start` to the bot gets a reply with the chat ID.
  → test: apps/api/internal/telegram/telegram_test.go "TestHandleUpdate_RepliesToStartWithChatID"
- Adding a Telegram channel without a chat ID is refused.
  → test: apps/api/internal/handler/notification_channels_test.go "TestCreateNotificationChannel/telegram missing chatId"
- "Send test message" reports a failure (wrong chat ID, bot not started, bot not configured) instead of pretending it worked.
  → test: apps/api/internal/handler/notification_channels_test.go "TestTestNotificationChannel/telegram with no bot token configured surfaces as a 502"
- A down alert names the monitor, its type, the reason and the time; a recovery alert names the monitor and how long it was down.
  → test: apps/api/internal/telegram/telegram_test.go "TestSendMessage_SendsChatIDAndText"
- A failed Telegram delivery doesn't count as delivered, so the monitor's other channels, or the email fallback, still carry the alert.
  → test: apps/api/internal/worker/worker_test.go "TestDispatchAlert/telegram channel attached but unreachable in this environment does not count as delivered"
- Updates to the bot webhook without the right secret header are ignored, and the webhook always answers `200` so Telegram never retries.
  → test: apps/api/internal/handler/settings_test.go "TestHandleTelegramWebhook/wrong secret header returns 200"
  → test: apps/api/internal/handler/settings_test.go "TestHandleTelegramWebhook/malformed JSON body still returns 200 (Telegram must never see a retry-worthy status)"
- The webhook secret is derived from the bot token and registered with Telegram at startup.
  → test: apps/api/internal/config/config_test.go "TestLoad/derives the Telegram webhook secret from the bot token"
  → test: apps/api/internal/telegram/telegram_test.go "TestSetWebhook_SendsURLAndSecret"
- A monitor with alerts turned off still records its status, incidents and history, but sends nothing.
  → test: apps/api/internal/worker/worker_test.go "TestCheckOverdue/an incident still opens when alerts are disabled, but no alert is sent"
- A muted monitor is marked as muted in its monitor list.
