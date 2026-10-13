# Frontend structure and web analytics

Date: 2026-10-13

The `use*Monitors.ts` composables are shallow on purpose: they keep the TanStack Query keys in one place. An architecture review should not flag them as pass-throughs to inline.

Web analytics is Google Analytics 4 through a Google Tag Manager container, loaded only after the visitor accepts the consent banner; declining loads nothing, and a missing `VITE_GTM_ID` loads nothing either. GTM lets new tags be added from its dashboard without a deploy. The `<noscript>` GTM fallback is left out on purpose: it would track without consent.

## Considered options

- PostHog: autocapture turns one visit into many billed events, so cost grows with interactions, not visits.
- Plausible, Fathom or Simple Analytics: no free hosted tier, and self-hosting is a server to run.

## Consequences

- No product analytics, session replay or feature flags; needing them is a new tool, not a GA4 change.
