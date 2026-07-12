# ADR-036: Flat safety caps on unbounded per-org creation

**Date:** 2026-07-12
**Status:** Accepted
**Revised:** 2026-07-12 — a follow-up re-audit of `status_public.go` found that the caps below, while each individually bounded, composed into an N+1 query pattern in `loadActiveIncidents`; batched into one query — see Consequences

---

## Context

Auditing [ADR-015](015-cron-pings-retention.md)'s active-incident cap for completeness surfaced three more unbounded-creation gaps in the same shape, none introduced today — all pre-existing:

- `PostIncidentUpdate` had no cap on updates per incident, and `ListStatusPageIncidentUpdates` had no `LIMIT`. The more serious of the three: `status_public.go`'s `loadActiveIncidents` renders an incident's full update timeline on every unauthenticated visitor's page load, so this was the one gap directly exposed to the public internet, not just a private dashboard.
- `CreateMaintenanceWindow` had no limit check at all, and `ListMaintenanceWindows` had no `LIMIT`.
- `CreateAPIKey` had no limit check at all, and `ListAPIKeys` had no `LIMIT`.

None of these are metered, per-plan-differentiated resources the way monitors/status pages/notification channels are (ADR-019) — they're operational/narration data, not something a higher plan should get more of. A flat cap, uniform across every plan, fits the same reasoning [ADR-015](015-cron-pings-retention.md) already used for the active-incident cap: nothing to upgrade past it, only something to clear (resolve, delete, revoke) to free a slot.

## Decision

Four flat, uniform-across-every-plan safety caps, each paired with a defensive `LIMIT 200` on the corresponding list query:

| Resource | Cap | Enforced in | List query LIMIT |
|---|---|---|---|
| Updates per incident | 100 | `incidents.go`'s `checkUpdateCap`, called from `PostIncidentUpdate` | `ListStatusPageIncidentUpdates` |
| Maintenance windows per org (cumulative — no retention exists for these) | 100 | `maintenance.go`'s `CreateMaintenanceWindow` | `ListMaintenanceWindows` |
| Active API keys per org | 100 | `api_keys.go`'s `CreateAPIKey` | `ListAPIKeys` |
| Active incidents per org (from [ADR-015](015-cron-pings-retention.md), restated here for the complete picture) | 100 | `incidents.go`'s `checkActiveIncidentCap`, called from `CreateIncident` | `ListStatusPageIncidents` / `ListActiveStatusPageIncidentsForPage` |

Each rejection is `409` with a distinct machine-readable code (`too_many_incident_updates`, `too_many_maintenance_windows`, `too_many_api_keys`, `too_many_active_incidents`) rather than reusing `plan_limit_reached` — that code specifically means "upgrade to raise this," which doesn't apply here.

The `100` number itself isn't derived from a capacity calculation the way [ADR-015](015-cron-pings-retention.md)'s original 30-day cron_pings retention was (~13M rows/month, ~1.3 GB/month math) — these are all low-volume, human-driven resources where 100 is simply generous enough that no legitimate use hits it, while still bounding worst-case growth to something trivial to store and render. If real usage ever approaches any of these caps, that's a signal to revisit the number, not evidence the cap was wrong to add.

## Consequences

- Closes the last of the unbounded-creation gaps found in this pass; `docs/reference/limits.md` and the `overload-audit` skill (`.claude/skills/overload-audit/audit.py`) both updated to track all four caps and list-query `LIMIT`s mechanically, so a future removal of any of them gets caught the same way a removed rate limit or check-loop semaphore would.
- No plan differentiation anywhere in this ADR — Hobby and Enterprise get the identical 100. If a future need arises to scale one of these with plan tier (e.g. Enterprise wanting more than 100 maintenance windows), that's a new decision, not an extension of this one.
- Maintenance windows now have a hard ceiling with no corresponding cleanup mechanism (unlike incidents, which pair their cap with 90-day retention on resolved ones). An org that hits 100 windows must delete an old one manually; there's no auto-expiry for ended windows. Acceptable for now since maintenance windows are typically short-lived scheduled events an org already manages by hand, but worth revisiting if this cap is ever actually reached in practice.
- Individually-bounded pieces can still compose into an unbounded query pattern. `loadActiveIncidents` (`status_public.go`) previously issued one `ListStatusPageIncidentUpdates` call per active incident on a page — with the caps above, that's now a bounded-but-still-large worst case (up to 100 active incidents × 100 updates each = up to 10,000 rows across up to 101 round-trips, all on one unauthenticated page load). Fixed by adding `ListStatusPageIncidentUpdatesForIncidents` (`WHERE incident_id = ANY($1::uuid[])`) and batching all of a page's active-incident updates into a single query, grouped in Go — a page's public render is now always 2 queries for its incident data (one for the incidents, one for all their updates), regardless of how many active incidents apply. Total row count is unchanged (still bounded by the per-resource caps, not reduced by this fix); what changed is round-trip count, not data volume.
