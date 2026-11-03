# Frontend structure and web analytics

Date: 2026-11-03

The `use*Monitors.ts` composables are shallow on purpose: they keep the TanStack Query keys in one place. An architecture review should not flag them as pass-throughs to inline.

Web analytics is Google Analytics 4 through a Google Tag Manager container, loaded only after the visitor accepts the consent banner; declining loads nothing, and a missing `VITE_GTM_ID` loads nothing either. GTM lets new tags be added from its dashboard without a deploy. The `<noscript>` GTM fallback is left out on purpose: it would track without consent.

The app only pushes events to the dataLayer; where they go is set in GTM, not in the repo. A successful sign-up pushes `{ event: 'sign_up', method: 'email' }` (`apps/web/src/lib/analytics.ts`), and a GTM Custom Event trigger on `sign_up` fires a GA4 Event tag of the same name, which GA4 counts as a key event. A run of a free checker (PRD free-tools) pushes `{ event: 'tool_check', tool, outcome }`; its GA4 Event tag reads `tool` and `outcome` through Data Layer Variables and sends them as parameters, registered in GA4 as the event-scoped custom dimensions Tool and Tool outcome. The owner's visits are tagged internal by a rule on their static IP in the Google tag's settings, and GA4's Internal Traffic data filter is Active, so a test sign-up from that IP never shows in GA4 reports or Realtime; check it in GTM Preview instead.

GA4 therefore counts only visitors who accept the banner, a small share of them: from 2026-10-25 to 2026-10-28 it recorded no session while kamal-proxy logged 1–12 people a day who loaded the app's JS. Traffic is counted from kamal-proxy's request log instead (`mise run traffic`, which logs the Referer to show where visitors came from), not from GA4. The log keeps only a few days, so a cron job on the server (`mise run traffic-install`) saves each finished day's aggregates, with no IPs, to `traffic/days.jsonl` in the deploy user's home, and `mise run traffic` shows those days next to the ones still in the log; GA4 stays for the `sign_up` and `tool_check` events of those who consent.

## Considered options

- PostHog: autocapture turns one visit into many billed events, so cost grows with interactions, not visits.
- Plausible, Fathom or Simple Analytics: no free hosted tier, and self-hosting is a server to run.
- GA4 consent mode's cookieless pings: GA4 models the missing visitors only from about 1,000 daily users, so below that the reports stay empty.
- Umami Cloud or Cloudflare Web Analytics (cookieless, free tiers): not needed while a few people a day can be counted from the log; they come back when channels need comparing over more than the few days the log keeps, or by sign-ups.

## Consequences

- No product analytics, session replay or feature flags; needing them is a new tool, not a GA4 change.
- Only the founder can read the proxy log and the saved days over SSH; bots with a browser User-Agent are told apart only by not loading the JS.
- A day is saved only by the cron job's run after it ends: days before the job was installed, or when it didn't run, keep only what the log still holds.
