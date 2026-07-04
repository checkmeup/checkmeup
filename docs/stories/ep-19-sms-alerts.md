# EP-19: SMS alerts

An eighth alert channel — unlike Slack/Teams ([EP-17](ep-17-slack-alerts.md)/[EP-18](ep-18-teams-alerts.md)), this needs a real SMS provider (Twilio, Vonage, AWS SNS) and a per-message cost, and is subject to anti-spam regulation (TCPA-style rules in the US and equivalents elsewhere) that requires explicit opt-in before sending automated texts — not just collecting a phone number.

Also builds on the multi-channel model in [EP-28](ep-28-notification-channels.md) ([ADR-023](../decisions/023-notification-channels.md)) — adds an `sms` value to `notification_channel_type`. "Off at the org level" below should read "channel disabled or not attached to that monitor" once EP-28 lands.

Provider and opt-in flow decided in [ADR-029](../decisions/029-sms-alerts-twilio.md): Twilio, with an explicit consent checkbox + timestamp recorded at connect time.

---

### US-1901: Connect a phone number for SMS alerts

**As a** user, **I want** to provide a phone number for SMS alerts **so that** I'm notified even without an internet connection.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [ ] Org setting: phone number, validated in E.164 format
- [ ] Explicit opt-in confirmation required before alerts start (regulatory requirement for automated texts, not just providing a number)
- [ ] "Send test SMS" button verifies delivery before saving
- [ ] Multiple phone numbers can be added per org, each its own `notification_channels` row (EP-28) — no longer limited to one

---

### US-1902: Receive a down alert via SMS

**As a** user, **I want** a text message when a monitor goes down **so that** I'm notified even with no data connection.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] Message kept within a single SMS segment (160 chars) where possible: monitor name, status, short reason
- [ ] Sent within one check cycle of the transition to "down"
- [ ] Not sent if alerts are disabled for that monitor or SMS alerts are off at the org level

---

### US-1903: Receive a recovery alert via SMS

**As a** user, **I want** a text message when a monitor recovers **so that** I know the incident is resolved.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] Message: monitor name, downtime duration, kept within a single SMS segment where possible
- [ ] Always sent on genuine recovery regardless of the per-incident alert cap ([ADR-016](../decisions/016-alert-debounce.md))
- [ ] Not sent if alerts are disabled for that monitor or SMS alerts are off at the org level

---

### US-1904: Handle SMS delivery failures and cost control

**As a** user, **I want** SMS failures handled gracefully and my SMS volume kept predictable **so that** a flapping monitor doesn't generate a surprise bill.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] Non-success delivery status (carrier rejection, invalid number) logged and visible in Settings — same pattern as webhook delivery status ([US-1404](ep-14-webhook-alerts.md))
- [ ] No automatic retries on MVP, consistent with webhook (US-1404)
- [ ] SMS strictly respects `max_alerts_per_incident` — no segment-splitting or other workaround that effectively sends more notifications than the cap allows

---

### US-1905: Shared alert cap across all channels

**As a** user, **I want** the per-incident alert cap to include SMS **so that** enabling every channel doesn't multiply my alert volume (or my SMS bill).

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] `max_alerts_per_incident` ([ADR-016](../decisions/016-alert-debounce.md)) counts a triggered alert once per incident-event across all enabled channels together, consistent with [US-1305](ep-13-email-alerts.md), [US-1405](ep-14-webhook-alerts.md), [US-1505](ep-15-whatsapp-alerts.md), [US-1605](ep-16-signal-alerts.md), [US-1705](ep-17-slack-alerts.md), and [US-1805](ep-18-teams-alerts.md)
- [ ] Recovery event is exempt from the cap on every enabled channel, including SMS
