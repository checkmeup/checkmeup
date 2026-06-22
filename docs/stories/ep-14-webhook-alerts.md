# EP-14: Webhook alerts

**Shipped 2026-06-22 (v1.7), 5/5 stories.**

A third, generic alert channel alongside Telegram ([EP-05](ep-05-telegram-alerts.md)) and email ([EP-13](ep-13-email-alerts.md)) — lets users wire monitor events into their own automation (Slack incoming webhooks, PagerDuty, custom scripts) without checkmeup building a first-class integration for each target. Distinct from the existing *incoming* Telegram webhook (`/telegram/webhook`, used to receive bot updates) — this is checkmeup acting as the sender.

Builds on the multi-channel model in [EP-28](ep-28-notification-channels.md) ([ADR-023](../decisions/023-notification-channels.md)), which landed first — each saved webhook is a `notification_channels` row, not a single org-level field. "Off at the org level" below reads as "channel disabled or not attached to that monitor."

---

### US-1401: Configure a webhook URL

**As a** user, **I want** to set a webhook URL **so that** monitor events can trigger my own automation.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] Add a webhook notification channel: URL must be `https://`, validated; multiple webhook channels can be added (EP-28)
- [x] "Send test webhook" button posts a sample payload and shows success/failure (status code or error) before saving
- [x] Webhook can be enabled independent of Telegram/email
- [x] A signing secret is generated automatically the first time the webhook is saved (used in US-1403)

---

### US-1402: Receive a down/recovery event via webhook

**As a** user, **I want** a POST request when a monitor goes down or recovers **so that** my own systems can react automatically.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] POST body (JSON): event type (`down` / `recovery`), monitor name, type (cron / uptime / SSL), reason (down only), downtime duration (recovery only), timestamp
- [x] Sent within one check cycle of the transition — same timing as Telegram ([US-0502](ep-05-telegram-alerts.md)/[US-0503](ep-05-telegram-alerts.md)) and email
- [x] Not sent if alerts are disabled for that monitor or webhook alerts are off at the org level
- [x] 10-second request timeout, matching the uptime check timeout convention ([ADR-014](../decisions/014-uptime-check-mechanics.md)) — a slow endpoint never blocks the worker

---

### US-1403: Verify webhook authenticity

**As a** user, **I want** to verify a webhook request really came from checkmeup **so that** I don't act on spoofed events.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] Every request includes an `X-Checkmeup-Signature` header: HMAC-SHA256 of the raw body using the org's signing secret (US-1401)
- [x] Signing secret viewable and regeneratable in Settings — regenerating invalidates the signature for future sends only, not retroactively
- [x] UI shows a short verification snippet so users can check the signature on their end

---

### US-1404: Handle webhook failures without retry storms

**As a** user, **I want** a failing webhook endpoint to not generate runaway retries **so that** a broken integration on my end doesn't affect monitor checking.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] No retries on MVP — one attempt per event; failures are logged, not retried
- [x] Non-2xx response or timeout recorded and visible in Settings (e.g. "Last delivery: failed, 500, 2 min ago")
- [x] A failing webhook never blocks or delays the worker's check loop — fire-and-forget within the US-1402 timeout

---

### US-1405: Shared alert cap across all channels

**As a** user, **I want** the per-incident alert cap to include webhook events **so that** enabling all three channels doesn't multiply my alert volume.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [x] `max_alerts_per_incident` ([ADR-016](../decisions/016-alert-debounce.md)) counts a triggered alert once per incident-event regardless of how many channels it's delivered to — same shared-cap design as email ([US-1305](ep-13-email-alerts.md))
- [x] Recovery event is exempt from the cap on every enabled channel, including webhook
