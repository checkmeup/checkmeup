# ADR-023: Multi-channel notification model — `notification_channels` table + per-monitor routing

**Status:** Accepted
**Date:** 2026-06-20

## Context

Today alert delivery is hardcoded to two org-level singular fields: `orgs.telegram_chat_id` and `orgs.alert_email` (+ `orgs.email_alerts_enabled`). `dispatchAlert` ([worker.go](../../apps/api/internal/worker/worker.go)) sends to whichever of these two are set, for every monitor in the org — there's no per-monitor routing and no way to have more than one chat/address of a given type.

Seven epics already in `docs/stories/` each repeat the same assumption — "One X per org on MVP, matching the existing one-Telegram-chat-per-org pattern": [EP-14](../stories/ep-14-webhook-alerts.md) (webhook), [EP-15](../stories/ep-15-whatsapp-alerts.md) (WhatsApp), [EP-16](../stories/ep-16-signal-alerts.md) (Signal), [EP-17](../stories/ep-17-slack-alerts.md) (Slack), [EP-18](../stories/ep-18-teams-alerts.md) (Teams), [EP-19](../stories/ep-19-sms-alerts.md) (SMS), [EP-20](../stories/ep-20-viber-alerts.md) (Viber). EP-14/17/18/20 are next up on [the roadmap](../roadmap.md), so this needed deciding before any of them are built, not after.

healthchecks.io and UptimeRobot were checked for prior art. Neither has a "group of channels" concept: healthchecks.io calls a connection an "integration," scoped to the project, toggled per-check; UptimeRobot calls the generic destinations "notification channels" and third-party ones "integrations," attached to monitors via a flat list (their older "Alert Contacts" model). Neither needed a grouping layer on top of individual channels.

## Decision

**Schema** — two new tables, following the existing polymorphic-join pattern from `maintenance_window_monitors` ([ADR-020](020-maintenance-windows.md)):

```sql
CREATE TYPE notification_channel_type AS ENUM ('telegram', 'email', 'webhook');
-- Each later channel epic (Slack/Teams/Discord/Viber/WhatsApp/Signal/SMS) adds
-- its own value via `ALTER TYPE notification_channel_type ADD VALUE '...'` in
-- its own migration, when that epic is actually built — not reserved upfront.

CREATE TABLE notification_channels (
    id          UUID                       PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID                       NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
    type        notification_channel_type  NOT NULL,
    name        TEXT                       NOT NULL,
    config      JSONB                      NOT NULL, -- {"chat_id": "..."} / {"webhook_url": "..."} / {"email": "..."}
    enabled     BOOLEAN                    NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ                NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ                NOT NULL DEFAULT NOW()
);

-- monitor_type is 'cron', 'uptime', or 'ssl'; no FK constraint, same pattern as
-- maintenance_window_monitors / status_page_monitors
CREATE TABLE monitor_notification_channels (
    id            UUID NOT NULL DEFAULT gen_random_uuid(),
    channel_id    UUID NOT NULL REFERENCES notification_channels(id) ON DELETE CASCADE,
    monitor_type  TEXT NOT NULL,
    monitor_id    UUID NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (channel_id, monitor_type, monitor_id)
);
```

**Naming** — "channel" for a single configured destination (one Telegram chat, one Slack webhook). No "group" entity: a monitor attaches directly to N channels via the join table, matching both competitors above and this repo's preference for not building structure ahead of actual demand.

**Monitor-level mute stays** — each monitor's existing `alerts_enabled` boolean ([ADR-016](016-alert-debounce.md)) remains a master kill-switch, independent of which channels are attached. Muting temporarily shouldn't require detaching and re-attaching channels.

**Fallback channel** — if a monitor has zero attached, enabled channels, alerts fall back to every user's login email (`SELECT email FROM users WHERE org_id = $1`), computed at dispatch time, not persisted as a row. Every org has exactly one user today ([EP-12](../stories/ep-12-team-management.md) — team invites aren't built), so this is unambiguous; writing the query against `org_id` rather than a single user means it doesn't need touching when EP-12 ships multi-user orgs. Delivery reuses `email.Sender.SendAlertEmail`, which already no-ops with a log warning when `RESEND_API_KEY` is unset ([ADR-012](012-email-resend.md)) — the fallback degrades the same way.

**This is a behavior change, applied retroactively, by deliberate choice.** Today an org with neither Telegram nor `alert_email` configured gets zero alerts on an incident (`hasAlertChannel` returns false). After this ships, every org always has at least the fallback. Confirmed 2026-06-20: apply to all existing orgs immediately, not just new ones going forward — an unannounced monitor going down with nobody told is worse than an unsolicited alert email.

## Migration path

1. **Additive migration** — create `notification_channels` + `monitor_notification_channels`. No existing data touched.
2. **Backfill** — for each org, create a `telegram` channel from `telegram_chat_id` (if set) and an `email` channel from `alert_email` (if `email_alerts_enabled`); attach both to every monitor that currently has `alerts_enabled = true`. Must produce bit-for-bit identical dispatch behavior to today, aside from the new fallback.
3. **Cutover** — rewrite `dispatchAlert` to loop over a monitor's attached channels through a `Notifier` registry keyed by `notification_channel_type`, instead of the two hardcoded `org.TelegramChatID` / `org.EmailAlertsEnabled` branches. Ships in the **same release** as the backfill — no window where channel rows exist but the worker still reads the old org columns.
4. **Drop legacy columns** (`orgs.telegram_chat_id`, `orgs.alert_email`, `orgs.email_alerts_enabled`) in a **separate, later migration**, only once step 3 has run in production and is confirmed working — keeps a rollback path if the cutover has a bug.

## Breaking changes

- **API contract**: `GET/PUT /api/v1/settings/telegram`, `/settings/email`, `/settings/email/enabled` ([settings.go](../../apps/api/internal/handler/settings.go)) are replaced by a `notification_channels` CRUD API plus a per-monitor attach/detach endpoint. No `/api/v2/` needed per [ADR-007](007-api-versioning.md) — single Docker image, frontend and backend always deploy together, no external API consumers yet ([EP-26](../stories/ep-26-public-api-keys.md) isn't built).
- **Settings UI**: the single Telegram-field / single-email-field forms become a channel list (add/remove/test per channel); a per-monitor channel picker is new UI that didn't exist before.
- **Dispatch behavior**: orgs with no channel configured now receive fallback alert emails they previously wouldn't have gotten — not a bug, a deliberate default (see above).
- **Downstream epics**: EP-14/15/16/17/18/19/20 each assert "one X per org on MVP" — no longer true once this ships, corrected in each epic as part of [EP-28](../stories/ep-28-notification-channels.md).

## Consequences

- New foundational epic [EP-28](../stories/ep-28-notification-channels.md) must land before EP-14/17/18/20 — moved to the front of the roadmap's "Next."
- `dispatchAlert` and the worker's three call sites gain a DB query per check cycle (load the monitor's channels) where today it reads two already-loaded org fields — negligible at current scale, revisit if monitor count grows large enough to matter.
- Secrets in `config` JSONB (Slack/Teams webhook URLs, future OAuth tokens) are bearer-token-like and should never round-trip unmasked in API responses after creation — same treatment future API keys ([EP-26](../stories/ep-26-public-api-keys.md)) will need; no existing masking pattern in this codebase yet, so EP-28 sets the precedent.
- Signal ([EP-16](../stories/ep-16-signal-alerts.md)) still doesn't fit the `config` JSONB / stateless-webhook shape the other types share — its `signal-cli` daemon dependency is a separate operational decision, unaffected by this ADR.
