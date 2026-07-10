---
title: Notification Channel Architecture
type: knowledge
status: current
updated: 2026-07-10
tags: [architecture, alerts, telegram, sms, email, webhook, status-page, backend]
scope: apps/api/internal/{telegram,slack,webhook,twilio,email,deliver,httpsafe}, internal/handler/notification_channels.go, internal/worker (dispatch)
superseded_by:
---

# Notification channel architecture

**Investigated:** 2026-07-10
**Scope:** the five outbound alert channels (Telegram, email, webhook, Slack, SMS) and the shared plumbing added while extending them.

## Summary

Channels are a polymorphic table ([ADR-023](../decisions/023-notification-channels.md)), not one table per type: `notification_channels` (type, name, JSONB `config`) plus a join table `monitor_notification_channels` (channel × monitor, no FK on the monitor side — same no-FK-on-polymorphic-target pattern as `maintenance_window_monitors`/`status_page_monitors`). A monitor can have zero or more channels attached; `worker.DispatchAlert` (see [worker-architecture.md](worker-architecture.md)) is the single fan-out point, with an org-wide email fallback if nothing else is attached or every send fails.

## Findings

1. **Five channel packages, one shape each: `NewClient(...) *Client` + `Send(...) (statusCode int, err error)`.** `internal/telegram`, `internal/slack`, `internal/webhook`, `internal/twilio` all follow this signature (email is the exception — `internal/email.Sender` predates the channel-table model, ADR-012, and only returns `error`). `statusCode == 0` means no response was ever received (timeout/connection error/blocked dial) as opposed to a real non-2xx HTTP response, so callers can classify delivery failures without inspecting `err`'s type.

2. **Shared HTTP delivery mechanics live in `internal/deliver`** (added 2026-07-10): a `Timeout` constant (10 s, used by all four HTTP-based channels) and a `Do(client, req, chanName, errFn) (statusCode int, err error)` helper that POSTs and classifies the response — 2xx is success, anything else calls `errFn(resp)` so each channel keeps its own error wording (Twilio decodes a JSON error body; webhook/Slack just report the status code). `internal/telegram` uses the shared `Timeout` but not `Do` — the Telegram Bot API always returns HTTP 200 and reports failure via a JSON `ok` field instead of the status code, a different enough response shape that forcing it through the same classification would be a force-fit.

3. **Only Slack and the generic webhook channel dial a fully user-supplied URL**, and both go through `internal/httpsafe` for that reason: `httpsafe.Dialer` blocks loopback/private/link-local (incl. `169.254.169.254`)/unspecified/multicast dial targets (DNS-rebinding-safe — the check runs in `net.Dialer.Control`, after resolution, on the actual connect address), and `httpsafe.RefuseRedirects` stops a 3xx response from retargeting the request after the initial URL passed muster. Telegram/Twilio don't need this — their endpoint is always `api.telegram.org`/`api.twilio.com` (or a test-injected `baseURL`), never derived from channel config. (Uptime/port/SSL monitor checks needed the same guard for a different reason — see [worker-architecture.md](worker-architecture.md) finding 4 — and share the same `httpsafe` package.)

4. **One-attempt, no-retry delivery everywhere** (webhook: US-1404, Slack: US-1704, SMS: US-1904) — a deliberate choice to avoid retry storms against a possibly-already-struggling receiver, not an oversight. Delivery outcome (`success`/`failed` + a detail string) is recorded on the channel row (`UpdateNotificationChannelDelivery`) for webhook/Slack/SMS so the UI can show "Last delivery: failed, 500, 2 min ago" (`NotificationChannelsCard.vue`'s `deliverySummary`); Telegram/email don't record this (no `lastDeliveryStatus` surfaced for them in the UI).

5. **SMS is metered and consent-gated, the other four aren't.** Before every SMS send, `consumeSMSCredit` (`worker.go`) checks the org's monthly credit quota ([ADR-032](../decisions/032-sms-credit-quotas.md), 1 credit/send today, no per-destination cost band yet) and records a `failed`/`sms credit quota exhausted` delivery outcome if none remain — this still counts as "channel failed," so `DispatchAlert`'s fallback-email path is the safety net, same as any other SMS failure. Saving/testing an SMS channel from the frontend also requires a TCPA-style consent checkbox (`smsConsent` in `useNotificationChannelForm.ts`, backend-enforced too — [ADR-029](../decisions/029-sms-alerts-twilio.md)); editing the phone number away from the number consent was given for un-checks the box, since a changed number is a new recipient needing fresh consent.

6. **Config shape is a flat `Record<string, string>` inside JSONB**, keyed differently per type (`configKey` in `apps/web/src/lib/notificationChannelTypes.ts`): `chatId` (telegram), `email` (email), `url` (webhook and Slack — same key, different validation: webhook requires `https://`, Slack requires `https://hooks.slack.com/`), `phone_number` (sms, E.164-validated both client- and server-side). Webhook channels also carry a `secret` (HMAC-SHA256 signing key for `X-Checkmeup-Signature`, `internal/webhook.Sign`/`GenerateSecret`) generated on first save and rotatable via `regenerateWebhookSecret`.

7. **Frontend split**: `NotificationChannelsCard.vue` (list + toggle/remove) and `NotificationChannelForm.vue` (add/edit) share `useNotificationChannelForm.ts` (reactive form state + save/test/regenerate orchestration) and `lib/notificationChannelTypes.ts` (static per-type metadata — label/icon/config-key/value-label/placeholder — plus the pure `validateChannelSaveInput`/`buildChannelConfig` functions). The static/pure pieces were split into `lib/` on 2026-07-10 specifically so the list view doesn't have to instantiate the form's reactive state just to render an icon or label.

8. **Migration path from org-level fields**: `notification_channels`/`monitor_notification_channels` (migration `017_notification_channels.sql`) backfilled from the pre-existing `orgs.telegram_chat_id`/`orgs.alert_email`/`orgs.email_alerts_enabled` columns, attaching a channel to every monitor that had `alerts_enabled = true` at cutover so dispatch behavior didn't change. Those org-level columns are intentionally still present (not dropped by that migration) per ADR-023's migration path.

## Follow-ups

- No shared retry/backoff or circuit-breaker exists across channels — none is currently needed (one-attempt-by-design, see finding 4), but if that changes, `internal/deliver.Do` is the one place to add it for the four HTTP-based channels.
- Email (`internal/email.Sender`) predates the channel-table/`deliver` conventions and doesn't share either — worth revisiting if email delivery ever needs the same status-recording/testability treatment as the other four.
