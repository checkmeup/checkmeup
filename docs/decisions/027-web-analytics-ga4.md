# ADR-027: Google Analytics 4 for web analytics

**Status:** Accepted
**Date:** 2026-07-02

## Context

checkmeup.net has no web analytics in place yet — no visibility into landing-page traffic, signup funnel drop-off, or feature usage.

Options considered:

- **PostHog Cloud** — generous headline free tier (~1M events/month) and bundles product analytics, session replay, and feature flags. Rejected: autocapture logs every click/scroll/pageview by default, so a single visitor generates 5–20 events, not one — the free tier burns through fast under real traffic and the bill then scales with *interactions*, not *visits*. Could be tamed by disabling autocapture and sending only pageview + a handful of custom events, but that gives up most of the product-analytics value while still carrying open-ended cost risk as traffic grows.
- **Plausible / Fathom / Simple Analytics** — privacy-first, no cookie banner needed, predictable flat-rate pricing by pageview volume. Rejected: no free hosted tier (from ~$9/mo); the only free option is self-hosting, which we explicitly don't want to run (extra infra, same bias as [ADR-001](001-worker-model.md)).
- **Google Analytics 4** — free and unlimited regardless of scale, hosted by Google, no infra to run. Tradeoffs: heavier setup than the privacy-first tools, cookie-consent implications in some jurisdictions, pageview/session analytics only (no built-in product analytics, session replay, or feature flags).

## Decision

Use **Google Analytics 4** for web analytics.

Cost predictability (free, unbounded by traffic) outweighs the product-analytics/session-replay features PostHog would have bundled in, and outweighs Plausible's cleaner privacy story — this is a solo-founder, cost-sensitive operation (see [ADR-026](026-billing-paddle-mor.md) context) where an open-ended usage-based bill is a worse outcome than a slightly heavier setup or a less polished privacy posture.

If product analytics (funnels, session replay, feature flags) becomes a real need later, revisit — that's a genuinely different tool category, not a reason to swap GA4 out preemptively.

## Integration approach

GA4 is delivered through a **Google Tag Manager** container (`GTM-WZWK6LR4`) rather than a raw `gtag.js` snippet — GTM lets the GA4 tag (and any future tag) be configured from the GTM dashboard without another code change, and it has first-class consent-state support if the consent model below ever needs to grow (e.g. per-region rules).

**Consent gating (frontend, `apps/web`):**

- `src/lib/consent.ts` — module-level singleton (same pattern as `useTheme`) holding `'granted' | 'denied' | undefined`, persisted to `localStorage` under `cookie_consent`.
- `src/lib/analytics.ts` — `loadGtm()` injects the GTM script tag and seeds `window.dataLayer`; no-ops if `VITE_GTM_ID` is unset (same posture as the missing-`RESEND_API_KEY`/missing-Paddle-token cases) or if already loaded. `trackPageview(path)` pushes a `page_view` event for SPA route changes, since GTM's own pageview trigger only fires once on the initial hard load.
- `src/components/ConsentBanner.vue` — Accept/Decline banner, shown whenever `status` is `undefined`. Accept calls `grant()` then `loadGtm()` immediately; Decline calls `deny()` and nothing is ever loaded for that visitor.
- `App.vue` loads GTM on mount if consent was already granted in a prior visit; `router.afterEach` (in `router/index.ts`) calls `trackPageview()` on every navigation, gated on `status.value === 'granted'`.
- `VITE_GTM_ID` is a public container ID (safe to expose client-side, same posture as `VITE_PADDLE_CLIENT_TOKEN`).

**Deliberate gap:** the static `<noscript>` iframe fallback Google's install instructions suggest (for no-JS visitors) is intentionally omitted from `index.html` — it would fire a tracking hit unconditionally, bypassing the consent gate entirely. Given the audience (developers, JS-dependent SPA already required to use the product), the no-JS case is not worth breaking consent-gating for.

## Consequences

- Zero analytics cost at any traffic volume.
- No tracking of any kind happens until a visitor explicitly accepts the consent banner — GDPR-compliant by construction rather than by after-the-fact config.
- No product analytics, session replay, or feature flags bundled — if those are needed later, that's a separate tool/decision, not a GA4 config change.
- New tags (e.g. conversion pixels) can be added later via the GTM dashboard without another frontend deploy — the frontend only owns the consent gate and the container script injection, not individual tags.
