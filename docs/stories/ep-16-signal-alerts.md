# EP-16: Signal alerts

A fifth alert channel alongside Telegram ([EP-05](ep-05-telegram-alerts.md)), email ([EP-13](ep-13-email-alerts.md)), webhook ([EP-14](ep-14-webhook-alerts.md)), and WhatsApp ([EP-15](ep-15-whatsapp-alerts.md)).

Also builds on the multi-channel model in [EP-28](ep-28-notification-channels.md) ([ADR-023](../decisions/023-notification-channels.md)) — adds a `signal` value to `notification_channel_type`. "Off at the org level" below should read "channel disabled or not attached to that monitor" once EP-28 lands. Note ADR-023 also flags that Signal's `signal-cli` dependency doesn't fit the stateless `config` JSONB shape the other types share as cleanly — worth a second look when this epic is actually picked up.

**Needs a go/no-go decision before implementation** (add to [decision backlog](../decisions/backlog.md) / write an ADR): unlike Telegram (official, free Bot API) or WhatsApp (official, paid Business API), **Signal has no official bot or business messaging API at all**. The only practical way to send Signal messages programmatically is self-hosting `signal-cli` (or a REST wrapper around it) and registering a dedicated phone number as a bot account — an always-on extra process, plus a phone number Signal's anti-abuse systems could rate-limit or ban for automated use. This conflicts with this repo's minimal-infra philosophy ([ADR-001](../decisions/001-worker-model.md): no Redis, job queue, or external broker — goroutine workers only) and is a meaningfully different operational commitment than the other channels. Worth confirming there's real user demand before taking this on.

---

### US-1601: Connect a Signal number

**As a** user, **I want** to link my Signal number to my organization **so that** I receive alerts there.

**Estimate:** 2 h

**Acceptance criteria:**

- [ ] User provides their Signal-registered phone number in Settings
- [ ] "Send test alert" button verifies delivery through the self-hosted `signal-cli` service before saving
- [ ] Multiple Signal numbers can be added per org, each its own `notification_channels` row (EP-28) — no longer limited to one
- [ ] Setup instructions note this depends on a self-hosted `signal-cli` service, not a first-party Signal feature

---

### US-1602: Receive a down alert via Signal

**As a** user, **I want** a Signal message when a monitor goes down **so that** I'm notified on a channel I already trust for sensitive messages.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] Content: monitor name, type (cron / uptime / SSL), reason, timestamp — same information as Telegram's down alert ([US-0502](ep-05-telegram-alerts.md))
- [ ] Sent within one check cycle of the transition to "down"
- [ ] Not sent if alerts are disabled for that monitor or Signal alerts are off at the org level

---

### US-1603: Receive a recovery alert via Signal

**As a** user, **I want** a Signal message when a monitor recovers **so that** I know the incident is resolved.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] Content: monitor name, downtime duration — same information as Telegram's recovery alert ([US-0503](ep-05-telegram-alerts.md))
- [ ] Always sent on genuine recovery regardless of the per-incident alert cap ([ADR-016](../decisions/016-alert-debounce.md))
- [ ] Not sent if alerts are disabled for that monitor or Signal alerts are off at the org level

---

### US-1604: Handle signal-cli unavailability gracefully

**As a** user, **I want** to know if my Signal alerts aren't being delivered **so that** I'm not silently missing notifications.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] If the `signal-cli` service is unreachable or errors, the failure is logged and visible in Settings (e.g. "Last delivery: failed, connection refused, 5 min ago") — same pattern as webhook delivery status ([US-1404](ep-14-webhook-alerts.md))
- [ ] A `signal-cli` outage never blocks or delays the worker's check loop, and never blocks delivery on other enabled channels
- [ ] No automatic retries on MVP, consistent with webhook (US-1404)

---

### US-1605: Shared alert cap across all channels

**As a** user, **I want** the per-incident alert cap to include Signal **so that** enabling every channel doesn't multiply my alert volume.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] `max_alerts_per_incident` ([ADR-016](../decisions/016-alert-debounce.md)) counts a triggered alert once per incident-event across all enabled channels together, consistent with [US-1305](ep-13-email-alerts.md), [US-1405](ep-14-webhook-alerts.md), and [US-1505](ep-15-whatsapp-alerts.md)
- [ ] Recovery event is exempt from the cap on every enabled channel, including Signal
