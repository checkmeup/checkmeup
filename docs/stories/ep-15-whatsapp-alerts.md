# EP-15: WhatsApp alerts

A fourth alert channel alongside Telegram ([EP-05](ep-05-telegram-alerts.md)), email ([EP-13](ep-13-email-alerts.md)), and webhook ([EP-14](ep-14-webhook-alerts.md)).

Also builds on the multi-channel model in [EP-28](ep-28-notification-channels.md) ([ADR-023](../decisions/023-notification-channels.md)) — adds a `whatsapp` value to `notification_channel_type`. "Off at the org level" below should read "channel disabled or not attached to that monitor" once EP-28 lands.

**Needs a provider decision before implementation** (add to [decision backlog](../decisions/backlog.md) / write an ADR): unlike Telegram's free bot API and freeform messages, WhatsApp business-initiated messages outside an active user conversation must use a **pre-approved Meta message template** — arbitrary text isn't allowed. This needs: a choice of provider (Meta Cloud API directly vs. a BSP like Twilio or MessageBird), template content submitted for Meta's approval (lead time, not instant), and a per-message cost (unlike Telegram, which is free) that varies by recipient country.

---

### US-1501: Connect a WhatsApp number

**As a** user, **I want** to connect my WhatsApp number to my organization **so that** I can receive alerts there.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [ ] User provides their WhatsApp number in Settings
- [ ] Opt-in confirmation step required before alerts start (WhatsApp policy for business-initiated templates) — same role as Telegram's `/start` step ([US-0501](ep-05-telegram-alerts.md))
- [ ] "Send test alert" button verifies the connection before saving
- [ ] Multiple WhatsApp numbers can be added per org, each its own `notification_channels` row (EP-28) — no longer limited to one

---

### US-1502: Receive a down alert via WhatsApp

**As a** user, **I want** a WhatsApp message when a monitor goes down **so that** I'm notified on a channel I check constantly.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] Sent using a pre-approved template with monitor name, type, reason, and timestamp as template variables — no freeform text
- [ ] Sent within one check cycle of the transition to "down"
- [ ] Not sent if alerts are disabled for that monitor or WhatsApp alerts are off at the org level

---

### US-1503: Receive a recovery alert via WhatsApp

**As a** user, **I want** a WhatsApp message when a monitor recovers **so that** I know the incident is resolved.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] Separate pre-approved template for recovery, with monitor name and downtime duration as variables
- [ ] Always sent on genuine recovery regardless of the per-incident alert cap ([ADR-016](../decisions/016-alert-debounce.md)) — consistent with Telegram and email
- [ ] Not sent if alerts are disabled for that monitor or WhatsApp alerts are off at the org level

---

### US-1504: Disable WhatsApp alerts

**As a** user, **I want** to turn WhatsApp alerts off **so that** I'm not charged for messages or notified on a channel I don't want.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] Org-level toggle for WhatsApp alerts, independent of Telegram/email/webhook
- [ ] Disconnecting removes the stored number; re-enabling requires the opt-in flow again (US-1501)
- [ ] Per-monitor mute ([US-0504](ep-05-telegram-alerts.md)) suppresses WhatsApp alerts the same as the other channels

---

### US-1505: Shared alert cap across all channels

**As a** user, **I want** the per-incident alert cap to include WhatsApp **so that** enabling every channel doesn't multiply my alert volume (or my WhatsApp bill).

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] `max_alerts_per_incident` ([ADR-016](../decisions/016-alert-debounce.md)) counts a triggered alert once per incident-event across Telegram/email/webhook/WhatsApp together, consistent with [US-1305](ep-13-email-alerts.md) and [US-1405](ep-14-webhook-alerts.md)
- [ ] Recovery event is exempt from the cap on every enabled channel, including WhatsApp
