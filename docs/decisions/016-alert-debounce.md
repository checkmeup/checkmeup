# ADR-016: Per-incident alert limit (alert debounce)

**Status:** Accepted  
**Date:** 2026-06-15

## Context

When a monitor stays down across multiple check cycles, the system can send a large number of Telegram alerts. The backlog question was: when to stay silent vs. keep alerting, and whether this is global or per-monitor.

## Decision

Each monitor has a `max_alerts_per_incident` integer field:

- **0 = always alert** — send a Telegram message on every failed check for the duration of the incident.
- **N > 0 = cap at N** — send at most N alerts per incident, then go silent until the monitor recovers. **Default: 3.**

The count (`alert_count`) is stored on the incident row and incremented after each alert is sent. On recovery, the incident is resolved and the next incident starts fresh with `alert_count = 0`.

## Why per-monitor, not global

Different monitors have different urgency. A payment processor going down warrants every alert; a low-priority internal tool does not. Per-monitor config matches how users already think about alerts (the `alerts_enabled` toggle is already per-monitor).

## Consequences

- New column `max_alerts_per_incident INTEGER NOT NULL DEFAULT 3` on `cron_monitors` (and future `uptime_monitors`).
- New column `alert_count INTEGER NOT NULL DEFAULT 0` on `cron_incidents` (and future `uptime_incidents`).
- Worker checks `alert_count < max_alerts_per_incident` (or `max_alerts_per_incident = 0`) before sending. Increments `alert_count` after a successful send.
- For **cron monitors**: the worker detects down once per incident — the limit mainly controls the initial alert (N ≥ 1 always alerts; N = 0 also always alerts on first detection). The limit becomes meaningful when uptime-style re-alerting is added later.
- For **uptime monitors** (Phase 3): each failed check while down is one potential alert, so the cap actively suppresses noise after N failures.
- Recovery alerts are always sent regardless of the cap — the user always knows when a monitor comes back up.
