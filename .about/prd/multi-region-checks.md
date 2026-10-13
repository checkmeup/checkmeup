# Checks from several regions

Status: proposed

## Problem

Every check runs from one server, so a network blip between that server and the target looks like the target going down.

## Users

Uptime-monitor users who need fewer false alarms or want to know where a site is unreachable from.

## Success signal

A failure seen from one region only no longer marks a monitor down.

## Scope

- Choosing check regions per uptime monitor, the number limited per plan; one region by default.
- Each region checking independently, results stored per region.
- Down only when enough selected regions fail in the same window; a single-region failure shown as a regional anomaly.
- A per-region breakdown on the detail page and in alerts.

## Non-goals

- Region choice for cron, SSL, domain or DNS monitors.

## Acceptance criteria

- A monitor with no region chosen behaves exactly as today.
- A selected region being unavailable doesn't delay the other regions' checks.
- With three regions and the majority rule, one failing region shows as a regional anomaly and the monitor stays up; two failing mark it down.
- The detail page shows each region's status, response time and last check.
- A down alert names the failing and the passing regions.
- Choosing more regions than the plan allows gets `402 plan_limit_reached`.

## Open questions

- Regional VPSes reporting to the one database, or a check-execution service; which regions first (see the Inbox).
- Whether the down rule is fixed (majority) or per monitor (any, majority, all).
