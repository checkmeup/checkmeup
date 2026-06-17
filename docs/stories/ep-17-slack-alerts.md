# EP-17: Slack alerts

A sixth alert channel, built as a thin Slack-specific formatter on top of the generic webhook delivery already built in [EP-14](ep-14-webhook-alerts.md) — Slack Incoming Webhooks are a plain HTTPS POST endpoint, official, free, and need no OAuth app or approval process (unlike [WhatsApp](ep-15-whatsapp-alerts.md)). The only real difference from EP-14 is the payload shape: Slack expects `text`/Block Kit, not checkmeup's generic event JSON.

---

### US-1701: Connect a Slack channel

**As a** user, **I want** to connect a Slack channel by pasting an Incoming Webhook URL **so that** I receive alerts there.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] Setup instructions: create a Slack Incoming Webhook in the target workspace/channel, paste the URL into Settings
- [ ] URL validated against the `hooks.slack.com/...` pattern
- [ ] "Send test message" button verifies the connection before saving
- [ ] One Slack webhook per org on MVP, matching the existing one-Telegram-chat-per-org pattern ([US-0501](ep-05-telegram-alerts.md))

---

### US-1702: Receive a down alert in Slack

**As a** user, **I want** a formatted Slack message when a monitor goes down **so that** my team sees it in the channel we already watch.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] Message formatted for Slack (Block Kit or `text` fallback) with monitor name, type, reason, timestamp — readable directly in Slack, not raw JSON
- [ ] Sent within one check cycle of the transition to "down", reusing the existing webhook delivery mechanics ([US-1402](ep-14-webhook-alerts.md): timeout, fire-and-forget)
- [ ] Not sent if alerts are disabled for that monitor or Slack alerts are off at the org level

---

### US-1703: Receive a recovery alert in Slack

**As a** user, **I want** a formatted Slack message when a monitor recovers **so that** my team knows the incident is resolved.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] Message includes monitor name and downtime duration, formatted for Slack
- [ ] Always sent on genuine recovery regardless of the per-incident alert cap ([ADR-016](../decisions/016-alert-debounce.md))
- [ ] Not sent if alerts are disabled for that monitor or Slack alerts are off at the org level

---

### US-1704: Handle Slack delivery failures

**As a** user, **I want** to know if my Slack webhook stops working (e.g. someone deletes it in the workspace) **so that** I'm not silently missing alerts.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] Non-2xx response (e.g. a removed webhook returning `404`) logged and visible in Settings — same pattern as the generic webhook channel ([US-1404](ep-14-webhook-alerts.md))
- [ ] No automatic retries on MVP, consistent with the generic webhook channel
- [ ] A failing Slack webhook never blocks or delays the worker's check loop

---

### US-1705: Shared alert cap across all channels

**As a** user, **I want** the per-incident alert cap to include Slack **so that** enabling every channel doesn't multiply my alert volume.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] `max_alerts_per_incident` ([ADR-016](../decisions/016-alert-debounce.md)) counts a triggered alert once per incident-event across all enabled channels together, consistent with [US-1305](ep-13-email-alerts.md), [US-1405](ep-14-webhook-alerts.md), [US-1505](ep-15-whatsapp-alerts.md), and [US-1605](ep-16-signal-alerts.md)
- [ ] Recovery event is exempt from the cap on every enabled channel, including Slack
