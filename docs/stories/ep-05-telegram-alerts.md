# EP-05: Telegram alerts

One Telegram chat per organization on MVP. Alert logic (debounce, thresholds) is shared across all monitor types.

**Update 2026-06-20:** migrated to the multi-channel model in [EP-28](ep-28-notification-channels.md) ([ADR-023](../decisions/023-notification-channels.md)) — the "one chat per org" constraint in US-0501 below describes the original MVP behavior, not the current one.

---

### US-0501: Connect a Telegram chat

**As a** user, **I want** to connect a Telegram chat to my organization **so that** I receive alerts there.

**Acceptance criteria:**

- [x] Setup instructions shown in the UI:
  1. Start the checkmeup bot in Telegram
  2. Send `/start` to get a chat ID
  3. Paste the chat ID into the settings form
- [x] "Send test message" button verifies the connection before saving
- [x] Success and failure states shown clearly (wrong ID, bot not started, etc.)
- [x] One chat per org on MVP (replacing the previous one if changed)

---

### US-0502: Receive a down alert

**As a** user, **I want** to receive a Telegram message when a monitor goes down **so that** I can act immediately.

**Acceptance criteria:**

- [x] Message format: monitor name, type (cron / uptime / SSL), reason, timestamp
- [x] Sent within one check cycle of the status transition to "down"
- [x] Not sent if alerts are disabled for that monitor (US-0504)

---

### US-0503: Receive a recovery alert

**As a** user, **I want** to receive a Telegram message when a monitor recovers **so that** I know the incident is resolved.

**Acceptance criteria:**

- [x] Message includes: monitor name, downtime duration
- [x] Only sent after a genuine "down" state — not after a transient check failure that never crossed the threshold
- [x] Not sent if alerts are disabled for that monitor

---

### US-0504: Disable alerts per monitor

**As a** user, **I want** to mute alerts for a specific monitor **so that** noisy or low-priority monitors don't interrupt me.

**Acceptance criteria:**

- [x] Alert toggle in each monitor's settings (on by default)
- [x] Muted monitors still track status and history — only notifications are suppressed
- [x] Muted state visible in the monitor list (icon or badge)
