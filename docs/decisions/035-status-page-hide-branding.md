# ADR-035: Per-status-page "hide branding" toggle, gated to paid plans

**Date:** 2026-07-12
**Status:** Accepted

---

## Context

The public status page footer has always unconditionally rendered "Powered by Checkmeup" plus FAQ/Terms/Privacy links, with no plan gating — [ADR-033](033-target-customer-freelancers.md) explicitly documented this as an intentional non-priority, deferred until formal multi-person agencies became the target customer ("no need to prioritize gating it until/unless formal agencies become the target").

That's now being prioritized directly: paid-plan orgs should be able to remove the branding footer from an individual status page.

---

## Decision

- Add `status_pages.hide_branding` (`BOOLEAN NOT NULL DEFAULT false`), a per-page setting rather than an org-wide one — an org can run multiple status pages ([ADR-005](005-status-page-same-domain.md)) and may want branding on some but not others (e.g. an internal page vs. a client-facing one).
- Settable only on paid plans (Solo/Startup/Enterprise) — Hobby ($0) requests to set it `true` are rejected with `402 plan_limit_reached`, the same pattern already used for SMS channel gating ([ADR-032](032-sms-credit-quotas.md)).
- Enforced defensively on downgrade: if an org drops to Hobby, `hide_branding` is cleared back to `false` on all of that org's status pages — mirroring the existing `EnforceMonitorLimit`/`EnforceNotificationChannelLimit` calls in the plan-change path, extended with the same idempotent no-op-if-compliant shape.
- When set, the public status page footer omits both the "Powered by Checkmeup" line and the FAQ/Terms/Privacy links — the whole footer becomes just the "Last updated" line.

---

## Consequences

- Supersedes ADR-033's framing of the unconditional footer as low-priority — it's no longer unconditional.
- No public-render-time plan lookup: gating happens at write-time (can't set `true` without a paid plan) and at downgrade-time (cleared on downgrade), so the unauthenticated status-page render path (`status_public.go`) stays a single query against `status_pages`, unchanged in shape.
- A brief downgrade window exists between "plan lapses" and "downgrade enforcement runs" where a stale `hide_branding=true` could still render on a Hobby org's page — accepted as consistent with how the existing count-based limits already behave (enforcement runs synchronously on the plan-change webhook, not on every page load).
