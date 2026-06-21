# EP-28: Notification channels

Foundational multi-channel alert infrastructure ([ADR-023](../decisions/023-notification-channels.md)), replacing the single `orgs.telegram_chat_id` / `orgs.alert_email` fields with a `notification_channels` table and per-monitor routing. Every channel epic from here on — [EP-14](ep-14-webhook-alerts.md) (webhook), [EP-15](ep-15-whatsapp-alerts.md) (WhatsApp), [EP-16](ep-16-signal-alerts.md) (Signal), [EP-17](ep-17-slack-alerts.md) (Slack), [EP-18](ep-18-teams-alerts.md) (Teams), [EP-19](ep-19-sms-alerts.md) (SMS), [EP-20](ep-20-viber-alerts.md) (Viber) — adds a `notification_channel_type` on top of this instead of inventing its own org-level field, so this has to ship first.

**Shipped 2026-06-21 (v1.6), 5/5 stories.** Migration 017 added `notification_channels` + `monitor_notification_channels`. The backfill migrated every org's existing Telegram chat / alert email into the new model, attached to every monitor that had alerts on, and the `dispatchAlert` cutover to the channel-based path shipped in the same release as the backfill — per the ADR-023 migration plan, with no window where channel rows existed but the worker still read the old org fields. One AC remains deliberately open: dropping the legacy `orgs.telegram_chat_id` / `alert_email` / `email_alerts_enabled` columns, which per the ADR-023 migration plan only happens in a later, separate migration once the cutover has run clean in production.

---

### US-2801: Add and manage notification channels

**As a** user, **I want** to add multiple notification channels (e.g. two Telegram chats, a Slack webhook) **so that** I'm not limited to one of each type.

**Estimate:** 2 h

**Acceptance criteria:**

- [x] Settings page lists all channels for the org: type, name, enabled state
- [x] Add/edit/delete a channel, with a type-specific config form (chat ID for Telegram, webhook URL for Slack/Teams, address for email)
- [x] "Send test message" verifies a channel before saving, same UX as today's Telegram/email test buttons
- [x] No limit on number of channels per type on MVP — a per-plan limit, if any, is a separate pricing decision, not in scope here

---

### US-2802: Attach channels to a monitor

**As a** user, **I want** to choose which channels a specific monitor alerts on **so that** a critical payment monitor can page Slack while a low-priority one only emails me.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [x] Per-monitor channel picker (multi-select) in the monitor's settings, alongside the existing `alerts_enabled` mute toggle
- [x] A monitor with channels attached but `alerts_enabled = false` sends nothing — mute always wins over channel attachment
- [x] A new monitor defaults to all of the org's enabled channels attached — matches today's implicit "every monitor alerts on every org channel" behavior, so existing users see no surprise on day one

---

### US-2803: Fall back to account email when nothing is configured

**As a** user, **I want** to still get notified by email even if I haven't set up any channels **so that** I never silently miss a monitor going down.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] A monitor with zero attached, enabled channels sends to every user's login email in the org instead of nothing
- [x] Applies immediately to existing orgs that currently have neither Telegram nor email alerts configured — no grandfathering ([ADR-023](../decisions/023-notification-channels.md))
- [x] Degrades the same way `SendAlertEmail` already does when `RESEND_API_KEY` is unset (log warning, no crash) — no new failure mode introduced

---

### US-2804: Migrate existing Telegram/email settings without losing alerts

**As an** existing user, **I want** my current Telegram chat and alert email to keep working **so that** this migration doesn't silently break my alerting.

**Estimate:** 2 h

**Acceptance criteria:**

- [x] Backfill migration creates a `telegram` channel from `orgs.telegram_chat_id` and an `email` channel from `orgs.alert_email` (if `email_alerts_enabled`) for every org that has them set
- [x] Backfilled channels are attached to every monitor that currently has `alerts_enabled = true` — dispatch behavior is identical to pre-migration, aside from the new fallback (US-2803)
- [x] `dispatchAlert` cutover to the channel-based path ships in the same release as the backfill, not a separate one — no window where channel rows exist but the worker still reads the old org fields
- [ ] Legacy `orgs.telegram_chat_id` / `alert_email` / `email_alerts_enabled` columns are dropped in a **follow-up migration only**, after the cutover has run in production — keeps a rollback path. Deliberately still open: the cutover (this release) needs to run clean in production first; the drop migration is a separate, later PR per the ADR-023 migration plan

---

### US-2805: Shared alert cap across all channels

**As a** user, **I want** the per-incident alert cap to count once across every attached channel **so that** attaching multiple channels to a monitor doesn't multiply alert volume.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [x] `max_alerts_per_incident` ([ADR-016](../decisions/016-alert-debounce.md)) counts one notification event per incident regardless of how many channels (or the fallback) it goes out on — same semantics as today, now keyed off the monitor's attached-channel list instead of the two hardcoded org fields
- [x] Recovery alert is exempt from the cap on every attached channel, including the fallback
