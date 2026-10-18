# Security hardening of public endpoints

## Problem

The ping endpoint, the Telegram webhook and every request body are reachable by anyone; without limits a single client can flood the database, burn quotas or spoof Telegram updates.

## Users

The platform itself; no user-visible feature.

## Success signal

No public endpoint can be flooded or spoofed by one client without hitting a limit or a signature check.

## Scope

- Per-token rate limit on pings, per-IP limits on the Telegram webhook and test sends.
- A global request-body size cap.
- A secret on the Telegram webhook.

## Non-goals

- Protection against distributed volumetric attacks (see the Inbox, ADR infrastructure).

## Acceptance criteria

- Pings are rate limited per token, not per IP, so many jobs behind one NAT don't share a limit.
  → test: apps/api/internal/server/server_test.go "TestPingTokenKey"
- Telegram updates without the right secret header are ignored (ADR alerting, PRD telegram-alerts).
  → test: apps/api/internal/handler/settings_test.go "TestHandleTelegramWebhook/wrong secret header returns 200"
- A 61st ping in a minute on one token gets `429`.
  → test: apps/api/internal/server/server_test.go "TestRoutes_PingsAreRateLimitedPerToken/a 61st ping in a minute on one token gets 429"
- A request body over 64 KB is refused with `413` before any handler reads it.
- More than 60 Telegram webhook calls a minute from one IP get `429`.
  → test: apps/api/internal/server/server_test.go "TestRoutes_TelegramWebhookIsRateLimitedPerIP/more than 60 calls a minute from one IP get 429"
