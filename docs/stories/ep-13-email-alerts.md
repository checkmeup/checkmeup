# EP-13: Email alerts

A second alert channel alongside Telegram ([EP-05](ep-05-telegram-alerts.md)), reusing the Resend integration already built for password reset ([ADR-012](../decisions/012-email-resend.md)). Note: ADR-012 currently states "alerts remain Telegram-only" — that line needs updating once this ships, since it's no longer accurate.

**Update 2026-06-20:** migrated to the multi-channel model in [EP-28](ep-28-notification-channels.md) ([ADR-023](../decisions/023-notification-channels.md)) — the org-level `alert_email` field below is replaced by `notification_channels` rows; the "off at the org level" phrasing in US-1302/1303/1304 describes the original MVP behavior, not the current one.

---

### US-1301: Set an alert email address

**As a** user, **I want** to set an email address for alerts **so that** I can receive notifications without relying on Telegram.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] Org setting: alert email address, defaults to the account owner's email at signup
- [x] Editable independently — e.g. to route to a shared address like `alerts@yourteam.com`
- [x] "Send test email" button verifies deliverability before saving
- [x] Email alerts can be enabled whether or not Telegram is connected — neither channel is required

---

### US-1302: Receive a down alert by email

**As a** user, **I want** an email when a monitor goes down **so that** I'm notified even if I miss the Telegram message.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] Content: monitor name, type (cron / uptime / SSL), reason, timestamp — same information as the Telegram message ([US-0502](ep-05-telegram-alerts.md))
- [x] Sent within one check cycle of the transition to "down"
- [x] Not sent if alerts are disabled for that monitor ([US-0504](ep-05-telegram-alerts.md)) or email alerts are off at the org level
- [x] Subject line includes the monitor name and "DOWN" for inbox scanning/filtering

---

### US-1303: Receive a recovery alert by email

**As a** user, **I want** an email when a monitor recovers **so that** I know the incident is resolved without checking Telegram.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [x] Content: monitor name, downtime duration — same information as the Telegram recovery message ([US-0503](ep-05-telegram-alerts.md))
- [x] Always sent on genuine recovery regardless of the per-incident alert cap ([ADR-016](../decisions/016-alert-debounce.md)) — matches existing Telegram behavior
- [x] Not sent if alerts are disabled for that monitor or email alerts are off at the org level

---

### US-1304: Enable or disable the email channel

**As a** user, **I want** to turn email alerts on or off at the org level **so that** I can use Telegram only, email only, or both.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [x] Org-level toggle for email alerts, independent of the Telegram connection
- [x] Per-monitor mute ([US-0504](ep-05-telegram-alerts.md)) suppresses both channels — no separate per-monitor, per-channel mute for MVP
- [x] Both channels can be enabled at once; each fires independently for the same event

---

### US-1305: Shared alert cap across channels

**As a** user, **I want** the per-incident alert cap to apply across all channels **so that** enabling email doesn't multiply my alert volume.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] `max_alerts_per_incident` ([ADR-016](../decisions/016-alert-debounce.md)) counts alerts across all enabled channels together, not per channel — a cap of 3 means 3 total notification events, regardless of how many channels each was sent on
- [x] When both channels are enabled, a single triggered alert is sent on every enabled channel and counts as one increment of `alert_count`
- [x] Recovery alert is exempt from the cap on every enabled channel, same as today's Telegram-only behavior
