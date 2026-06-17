# EP-18: Microsoft Teams alerts

A seventh alert channel, built the same way as [Slack](ep-17-slack-alerts.md): a thin formatter on top of the generic webhook delivery already built in [EP-14](ep-14-webhook-alerts.md). One wrinkle specific to Teams: Microsoft is retiring the legacy Office 365 Connector "Incoming Webhook" in favor of **Power Automate workflows** ("Post to a channel when a webhook request is received"), which also hand the user a POST URL but expect an **Adaptive Card** JSON payload instead of Slack's `text`/Block Kit shape. Verify Microsoft's current retirement timeline before implementing — don't build against the legacy connector.

---

### US-1801: Connect a Teams channel

**As a** user, **I want** to connect a Microsoft Teams channel by pasting a workflow webhook URL **so that** I receive alerts there.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] Setup instructions: create a Power Automate workflow ("Post to a channel when a webhook request is received") in the target Teams channel, paste the generated URL into Settings
- [ ] URL validated against the expected Power Automate webhook URL pattern
- [ ] "Send test message" button verifies the connection before saving
- [ ] One Teams webhook per org on MVP, matching the existing one-Telegram-chat-per-org pattern ([US-0501](ep-05-telegram-alerts.md))

---

### US-1802: Receive a down alert in Teams

**As a** user, **I want** a formatted Teams message when a monitor goes down **so that** my team sees it in the channel we already watch.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] Message formatted as an Adaptive Card with monitor name, type, reason, timestamp — readable directly in Teams, not raw JSON
- [ ] Sent within one check cycle of the transition to "down", reusing the existing webhook delivery mechanics ([US-1402](ep-14-webhook-alerts.md): timeout, fire-and-forget)
- [ ] Not sent if alerts are disabled for that monitor or Teams alerts are off at the org level

---

### US-1803: Receive a recovery alert in Teams

**As a** user, **I want** a formatted Teams message when a monitor recovers **so that** my team knows the incident is resolved.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] Adaptive Card includes monitor name and downtime duration
- [ ] Always sent on genuine recovery regardless of the per-incident alert cap ([ADR-016](../decisions/016-alert-debounce.md))
- [ ] Not sent if alerts are disabled for that monitor or Teams alerts are off at the org level

---

### US-1804: Handle Teams delivery failures

**As a** user, **I want** to know if my Teams webhook stops working (e.g. the workflow is deleted or disabled) **so that** I'm not silently missing alerts.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] Non-2xx response logged and visible in Settings — same pattern as the generic webhook channel ([US-1404](ep-14-webhook-alerts.md))
- [ ] No automatic retries on MVP, consistent with the generic webhook channel
- [ ] A failing Teams webhook never blocks or delays the worker's check loop

---

### US-1805: Shared alert cap across all channels

**As a** user, **I want** the per-incident alert cap to include Teams **so that** enabling every channel doesn't multiply my alert volume.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] `max_alerts_per_incident` ([ADR-016](../decisions/016-alert-debounce.md)) counts a triggered alert once per incident-event across all enabled channels together, consistent with [US-1305](ep-13-email-alerts.md), [US-1405](ep-14-webhook-alerts.md), [US-1505](ep-15-whatsapp-alerts.md), [US-1605](ep-16-signal-alerts.md), and [US-1705](ep-17-slack-alerts.md)
- [ ] Recovery event is exempt from the cap on every enabled channel, including Teams
