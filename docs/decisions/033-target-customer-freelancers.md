# ADR-033: Primary target customer — freelance web devs / solo agency operators

**Date:** 2026-07-04
**Status:** Accepted

---

## Context

Early "how do we get first customers" discussion started from the differentiators list's assumption that agencies (wanting white-label status pages) were the target. Checking the actual status page implementation (`apps/api/internal/handler/status_public.go`) found that claim overstated: the URL is always `checkmeup.net/status/:slug` (custom domain deferred to Enterprise per [ADR-005](005-status-page-same-domain.md)), and the footer unconditionally renders "Powered by checkmeup" with no plan gating. Pitching formal multi-person agencies on "white-label" today would get relitigated in the first call.

Widening the target to "small businesses" was too broad on its own — non-technical local business owners (a restaurant, a shop) wouldn't self-serve configure a cron/SSL/uptime monitor. Narrowing that to freelance web devs / solo agency operators resolved it: this group builds and maintains sites (WordPress/Webflow/Shopify) for exactly those non-technical local businesses, is technical enough to self-serve, and faces the recurring pain checkmeup solves directly — SSL/domain expiry surprises and silent downtime on a client's site are a classic freelancer reputation risk.

## Decision

Primary target customer: **freelance web devs / solo agency operators who build and maintain multiple client sites for local small businesses.**

Not targeting for now:

- **Non-technical small businesses directly** — wouldn't self-serve configure monitors; would need a done-for-you wrapper, out of scope.
- **Formal multi-person agencies** — reachable, but a credible pitch needs the not-yet-shipped white-label features (custom domain, gated "Powered by" footer); revisit once those ship.

Secondary/parallel channel: broad developer self-monitoring audience (HN, Indie Hackers, r/webdev, r/selfhosted) for free-tier volume and product feedback — this competes directly with Healthchecks.io/Cronitor/UptimeRobot's actual customer base, so treated as top-of-funnel, not the primary paid-conversion path.

## Rationale

- **Self-serve-able** — no sales process needed; a freelancer can sign up and configure a monitor same-day.
- **Real recurring pain** — SSL/domain expiry and silent downtime on a client's site damages the freelancer's own reputation with that client.
- **Less crowded** — "monitor my own infra" is the saturated self-dev market; "monitor other people's sites and hand them a status page" is a narrower, less-contested angle that fits checkmeup's actually-shipped differentiators (multi-channel alerts, status page, flat pricing) without needing custom domains.
- **Reachable without a gatekeeper** — Upwork/Fiverr profiles (search "WordPress maintenance," "website support"), r/freelance, r/webdev, r/WordPress, local (Israeli) freelancer communities.

## Consequences

- Outreach and landing-page copy should default to this ICP going forward: "give your clients a professional status page" framing, not "white-label."
- The status page's unconditional "Powered by checkmeup" footer and lack of custom domain are acceptable for this segment (they don't expect enterprise-grade white-labeling) — no need to prioritize gating it until/unless formal agencies become the target.
- Positioning/landing-page copy may eventually need to speak to this ICP specifically rather than a generic "developer monitoring" framing — not addressed by this ADR, a future copy pass.
