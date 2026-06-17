# EP-20: Viber alerts

A ninth alert channel. Viber's official Bots REST API is closer in shape to Telegram ([EP-05](ep-05-telegram-alerts.md)) than WhatsApp ([EP-15](ep-15-whatsapp-alerts.md)): free, official, and once a user subscribes to the checkmeup bot, messages are sent as plain text with no pre-approved template required. The one prerequisite is that subscription step — a user must message the bot first before it can message them back, the same constraint as Telegram's `/start`.

---

### US-2001: Connect a Viber account

**As a** user, **I want** to connect my Viber account by subscribing to the checkmeup Viber bot **so that** I receive alerts there.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [ ] Setup instructions: open a link to start a conversation with the checkmeup Viber bot, which subscribes the user (required before the bot can message them, same constraint as Telegram's `/start` — [US-0501](ep-05-telegram-alerts.md))
- [ ] Viber user ID captured via the bot's subscribe webhook callback, or pasted manually into Settings
- [ ] "Send test message" button verifies the connection before saving
- [ ] One Viber subscriber per org on MVP, matching the existing one-Telegram-chat-per-org pattern

---

### US-2002: Receive a down alert via Viber

**As a** user, **I want** a Viber message when a monitor goes down **so that** I'm notified on another channel I already use daily.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] Content: monitor name, type (cron / uptime / SSL), reason, timestamp — same information as Telegram's down alert ([US-0502](ep-05-telegram-alerts.md))
- [ ] Sent within one check cycle of the transition to "down"
- [ ] Not sent if alerts are disabled for that monitor or Viber alerts are off at the org level

---

### US-2003: Receive a recovery alert via Viber

**As a** user, **I want** a Viber message when a monitor recovers **so that** I know the incident is resolved.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] Content: monitor name, downtime duration — same information as Telegram's recovery alert ([US-0503](ep-05-telegram-alerts.md))
- [ ] Always sent on genuine recovery regardless of the per-incident alert cap ([ADR-016](../decisions/016-alert-debounce.md))
- [ ] Not sent if alerts are disabled for that monitor or Viber alerts are off at the org level

---

### US-2004: Handle an unsubscribed or unreachable Viber account

**As a** user, **I want** to know if my Viber connection breaks (e.g. I unsubscribe from the bot) **so that** I'm not silently missing alerts.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] Viber API errors (e.g. user unsubscribed, invalid receiver) logged and visible in Settings — same pattern as webhook delivery status ([US-1404](ep-14-webhook-alerts.md))
- [ ] No automatic retries on MVP, consistent with the other channels
- [ ] A Viber API outage never blocks or delays the worker's check loop, or delivery on other enabled channels

---

### US-2005: Shared alert cap across all channels

**As a** user, **I want** the per-incident alert cap to include Viber **so that** enabling every channel doesn't multiply my alert volume.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] `max_alerts_per_incident` ([ADR-016](../decisions/016-alert-debounce.md)) counts a triggered alert once per incident-event across all enabled channels together, consistent with [US-1305](ep-13-email-alerts.md), [US-1405](ep-14-webhook-alerts.md), [US-1505](ep-15-whatsapp-alerts.md), [US-1605](ep-16-signal-alerts.md), [US-1705](ep-17-slack-alerts.md), [US-1805](ep-18-teams-alerts.md), and [US-1905](ep-19-sms-alerts.md)
- [ ] Recovery event is exempt from the cap on every enabled channel, including Viber
