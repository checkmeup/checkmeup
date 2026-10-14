---
title: Feature Bucket List — Competitor Gap Analysis
type: proposal
status: proposed
updated: 2026-10-14
tags: [features, competitors, backlog]
---

# Feature bucket list — competitor gap analysis

Features [healthchecks.io](https://healthchecks.io/), [uptimerobot.com](https://uptimerobot.com/) and [cronitor.io](https://cronitor.io/) have and Checkmeup doesn't, from an audit on 2026-06-23. Only gaps are listed: anything shipped or already a PRD is left out.

Not a committed plan. When an item gets prioritized, write its PRD in [`.about/prd/`](../../.about/prd/) and add it to [`roadmap.md`](../roadmap.md).

---

## Monitor types

- **ICMP ping monitoring** — the simplest "is it reachable at all" check, below HTTP and TCP. UptimeRobot.
- **30-second check interval** — deferred in [ADR billing](../../.about/adr/billing.md): the schema and the worker count in minutes.

## Alerting integrations

- **PagerDuty / Opsgenie** (and Spike.sh, PagerTree) — on-call routing, not just a notification. Healthchecks.io has both. Most useful to Startup and Enterprise teams already running rotations; it would be one more channel type on [PRD notification-channels](../../.about/prd/notification-channels.md).
- **Voice call alerts** — a louder escalation than SMS. UptimeRobot and Healthchecks.io. Needs a telephony provider decision, like SMS did.
- **Zapier** — both competitors have it, but the generic webhook ([PRD webhook-alerts](../../.about/prd/webhook-alerts.md)) already covers most of it, since Zapier can be triggered by a webhook. Not worth a dedicated integration unless users ask.

## Status pages

- **Visitor email subscriptions** — a visitor opts in to email when an incident is posted. UptimeRobot. Today an incident is seen only by someone who opens the page.
- **Password-protected status pages** — a client-only page instead of a public one. UptimeRobot. Fits the white-label-for-agencies angle.
- **Noindex toggle** — keep a status page out of search engines. UptimeRobot. Small and cheap.

Custom domains are out of scope: [ADR status-pages](../../.about/adr/status-pages.md) chose `/status/:slug` paths over subdomains and custom domains.

## Developer experience

- **Cron job output capture (stdout/stderr)** — none of the three competitors does this, so it would be a differentiator. What exists now is narrower: `GET /ping/{token}` stores the query-string params of the latest ping (up to 20 pairs) and the public API returns them as `lastPingMetadata`. Free-text output would need a request body, larger storage and its own retention and size policy.
- **CLI / language SDKs for cron jobs** — Cronitor ships open-source wrappers that run a command and ping on success or failure, instead of a hand-written `curl`. Lowers the setup effort of [PRD cron-monitoring](../../.about/prd/cron-monitoring.md) without changing the ping model.
- **Published check IP ranges** — UptimeRobot publishes the IPs its checks come from so users behind a firewall can allowlist them. A docs page plus a stable egress IP; removes a real blocker for monitoring non-public endpoints.

## Mobile

- **Native app with push notifications** — UptimeRobot has iOS and Android apps. A heavy lift for a one-person team; the responsive web app and existing channels cover "alert me on my phone". Listed as a gap, not a recommendation.

## Trust and access

- **SOC 2 / GDPR / CCPA messaging** — UptimeRobot markets it on Enterprise. A landing-page mention once it is true; no code.
- **Role-based permissions** — UptimeRobot. Depends on [PRD team-management](../../.about/prd/team-management.md) landing first.

---

## Ruled out by existing decisions

- A queue or broker for multi-region scheduling ([ADR workers](../../.about/adr/workers.md)): multi-region checks, if built, have to fit the goroutine worker, e.g. regional workers polling the same database.
- Status-page subdomains or custom domains ([ADR status-pages](../../.about/adr/status-pages.md)).
- The `Authorization` header for integrations ([ADR auth](../../.about/adr/auth.md)): anything machine-facing authenticates with `X-API-Key`, like the public API.
