---
title: Feature Bucket List — Competitor Gap Analysis
type: proposal
status: proposed
updated: 2026-07-05
tags: [features, competitors, backlog]
---

# Feature bucket list — competitor gap analysis

Source: feature audit of [healthchecks.io](https://healthchecks.io/), [uptimerobot.com](https://uptimerobot.com/), and [cronitor.io](https://cronitor.io/) (checkmeup's three named competitors — see `project_checkmeup` memory), done 2026-06-23. Cross-referenced against [`roadmap.md`](../roadmap.md), `stories/`, and `decisions/` to exclude anything already shipped or already queued — this is a list of *gaps*, not a restatement of the existing backlog.

Not a committed plan. Pull items into `roadmap.md`/`stories/` individually when one gets prioritized, same as any other epic.

---

## Already covered (not repeated below)

Multi-channel alerts (Telegram, email, webhook, SMS shipped; Slack/Teams/WhatsApp/Signal/Viber queued), maintenance windows, status pages with manual incidents (EP-24), keyword monitoring on uptime checks, 2FA (EP-25), public API + keys (EP-26), team management (EP-12), annual billing.

---

## New monitor types

- **DNS record monitoring** — alert when a DNS record changes or resolves unexpectedly (e.g. an A record pointing somewhere it shouldn't). UptimeRobot.
- **ICMP ping monitoring** — simplest possible "is it reachable at all" check, below HTTP/TCP. UptimeRobot.
- **30-second check interval** — already identified as a competitive gap in [ADR-019](../decisions/019-plan-limits.md) ("30-second interval — deferred"), not new to this audit. Flagging here for visibility since it's exactly the kind of thing this doc is meant to surface, but it's already tracked, not orphaned.

> Domain expiry monitoring, assertion-based API checks, and multi-region checking were pulled from this section into [EP-29](../stories/ep-29-domain-expiry-monitoring.md), [EP-31](../stories/ep-31-assertion-checks.md), and [EP-32](../stories/ep-32-multi-region-checking.md) respectively on 2026-06-23 — see [`roadmap.md`](../roadmap.md). TCP/port monitoring pulled in as [EP-33](../stories/ep-33-port-monitoring.md) on 2026-07-01.

## Alerting / ops integrations

- **PagerDuty / Opsgenie** (and similar: Spike.sh, PagerTree) — on-call incident routing, not just a notification. Healthchecks.io supports both. Higher value for Startup/Enterprise-tier teams already running on-call rotations; natural next channel type on top of the EP-28 notification-channels model once Slack/Teams (already queued) ship.
- **Voice call alerts** — UptimeRobot and Healthchecks.io both offer it as a more attention-grabbing escalation than SMS. Lower priority than the above; needs a telephony provider decision similar to the EP-19 SMS one.
- **Zapier** — both competitors have it, but checkmeup's generic webhook (EP-14, shipped v1.7) already covers most of the same ground (Zapier can itself receive a webhook as a trigger). Low incremental value; not worth a dedicated integration unless users specifically ask.

## Status page polish

> Public status badges pulled into [EP-30](../stories/ep-30-status-badges.md) on 2026-06-23 — see [`roadmap.md`](../roadmap.md).

- **Visitor email subscriptions** — let a status page visitor opt in to email updates when an incident is posted. UptimeRobot. Complements the manual-incident system (EP-24) once it ships — currently incidents are visible only if someone checks the page.
- **Password-protected / private status pages** — UptimeRobot. For agencies who want a client-only page instead of fully public, which is a real fit given checkmeup's existing white-label-for-agencies angle.
- **Search-engine indexing control (noindex toggle)** — UptimeRobot. Minor, cheap.
- **Custom domains for status pages** — UptimeRobot has it, but this is explicitly **out of scope**: [ADR-005](../decisions/005-status-page-same-domain.md) deliberately chose path-based `/status/:slug` over subdomains/custom domains, and it's a hard "Don't" in `CLAUDE.md`. Not re-suggesting it; noted here only so it isn't mistaken for an oversight.

## Developer experience

- **Execution log capture (cron monitor's stdout/stderr)** — none of the three named competitors do this, so it'd still be a real differentiator, but full output capture is **not shipped**. What *did* ship (2026-07-03, part of [EP-26](../stories/ep-26-public-api-keys.md)'s public API work) is a narrower, query-based version: `GET /ping/{token}` (`ping.go`) now captures the request's query-string params (e.g. `?build=142&state=success`) into a `metadata JSONB` column on `cron_pings` (`028_cron_ping_metadata.sql`), capped at 20 key/value pairs, 64-char keys, 256-char values — silently truncated past that, never rejected, since a ping must always return 200. Only the *most recent* ping's metadata is kept (overwritten each ping, not a history), and it's surfaced read-only via the public API's `GET /api/v1/public/monitors/cron/{id}/status` endpoint as `lastPingMetadata` — see the Public API section of [`/docs`](../../apps/web/src/views/DocsView.vue). This covers structured key/value tagging (a CI job reporting its build number and outcome) but not arbitrary free-text stdout/stderr capture — that would still need a request body (POST or a body on the existing GET), unbounded-ish text storage, and its own retention/size policy, which this feature deliberately avoided by staying query-param-sized.
- **CLI / language SDKs for cron jobs** — Cronitor publishes open-source SDKs that wrap a job command and auto-ping on success/failure, instead of users hand-rolling a `curl` call. Would lower the integration friction on EP-02's existing ping-URL mechanism without changing the underlying model.
- **Publish static monitoring IP ranges** — UptimeRobot publishes the IPs its checks come from so users behind a firewall can allowlist them. Cheap (a docs page + stable egress IP on the Hetzner box) and removes a real adoption blocker for users monitoring non-public endpoints.

## Mobile & access

- **Native mobile app with push notifications** — UptimeRobot has iOS/Android apps. Meaningful lift for a one-person team maintaining a Go+Vue stack; a responsive web app plus the existing/queued push-capable channels (webhook → push-notification bridge, or a future dedicated channel) likely covers the "get alerted on my phone" need without taking on two more codebases. Flagging as a gap, not recommending it outright.

## Trust signals (marketing, not engineering)

- **SOC 2 / GDPR / CCPA compliance messaging** — UptimeRobot markets this at the Enterprise tier. Worth a landing-page mention once actually true; not a code change.
- **Role-based granular permissions** — UptimeRobot. Depends on EP-12 (team management) landing first, which is blocked on the multi-user-org schema decision in the [decision backlog](../decisions/backlog.md). Note the dependency rather than treat this as standalone.

---

## Explicitly out of scope (existing architecture decisions)

Cross-checked against `CLAUDE.md`'s "Don't" list and the ADRs so nothing above contradicts a decision already made:

- No Redis/job queue/external broker ([ADR-001](../decisions/001-worker-model.md)) — rules out a queue-based approach to multi-region scheduling; if multi-region ever gets built, it needs to fit the existing goroutine-worker model (e.g. independent regional workers polling the same DB), not a new broker.
- No ORM ([ADR-004](../decisions/004-sqlc-over-orm.md)) — n/a to any item above.
- No status-page subdomains or custom domains ([ADR-005](../decisions/005-status-page-same-domain.md)) — see Status page polish above.
- No `Authorization` header for auth ([ADR-003](../decisions/003-auth-jwt-httponly-cookie.md)) — the planned public API (EP-26) already routes around this with a dedicated `X-API-Key` header per the decision backlog; any future integration work (Zapier, PagerDuty, etc.) should follow the same pattern.

---

## If picking a starting point

**Update 2026-06-23:** the four items previously listed here (domain expiry monitoring, public status badges, assertion-based API checks, multi-region checking) have been pulled into the roadmap as [EP-29](../stories/ep-29-domain-expiry-monitoring.md), [EP-30](../stories/ep-30-status-badges.md), [EP-31](../stories/ep-31-assertion-checks.md), and [EP-32](../stories/ep-32-multi-region-checking.md). EP-29 and EP-30 shipped the same day they were pulled in (see [`reports/2026-06.md`](../reports/2026-06.md)); EP-31 has mostly shipped, with its remaining piece (US-3105, chained API checks) now in `roadmap.md`'s **Later** section; EP-32 is also in **Later**, blocked on the multi-region infra decision in the [decision backlog](../decisions/backlog.md) — it was the only one of the four that breaks the single-Hetzner-CX23 model (needs compute in more than one region), so it's tracked as an infra decision rather than folded into a quick epic.

Remaining items in this doc are still unprioritized gaps — pull individually into `roadmap.md`/`stories/` as above when one gets picked up.
