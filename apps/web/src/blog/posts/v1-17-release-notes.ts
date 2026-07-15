import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: 'Every release note on this blog so far has been written blind to how many people actually read it. Checkmeup.net had no analytics at all — no idea how visitors find the landing page, where they drop off before signing up, or which docs pages get traffic. v1.17 fixes that with Google Analytics 4, delivered through a Google Tag Manager container rather than a raw gtag.js snippet, so tags can be added or changed later from the GTM dashboard without another frontend deploy.',
  },
  {
    type: 'h3',
    text: 'Why GA4 over the privacy-first options',
  },
  {
    type: 'p',
    text: "PostHog and Plausible were both on the table first. PostHog's free tier looks generous until you notice autocapture logs every click and scroll by default — a single visitor becomes 5-20 events, not one, and the bill scales with interactions rather than visits. Plausible and its peers are cleaner on privacy but have no free hosted tier, and self-hosting is the same infra burden this project has avoided since ADR-001. GA4 is free and unlimited regardless of traffic, which matters more than a nicer privacy story for a solo-founder, cost-sensitive product. The full comparison is in ADR-027 if you want the reasoning written out.",
  },
  {
    type: 'h3',
    text: 'Nothing loads until you say yes',
  },
  {
    type: 'p',
    text: "The tradeoff for going with GA4 is that it needs a real consent gate, not a cosmetic one. A new banner offers Accept or Decline; declining means the GTM script never gets injected into the page at all, not just that events stop being sent. Accepting loads the container immediately and persists the choice to localStorage, so returning visitors who already said yes don't see the banner again — and on every route change afterward, a page_view event fires manually, since GTM's own pageview trigger only catches the initial hard load, not client-side SPA navigation. The one thing deliberately left out is the static noscript fallback Google's own install instructions suggest — it fires a tracking hit unconditionally, which would defeat the entire consent gate for the sake of a no-JS audience that doesn't really exist for a JS-dependent SPA.",
  },
  {
    type: 'p',
    text: "Getting the GTM container ID into production took two follow-up fixes after the feature itself shipped. Vite only bakes VITE_* variables in at build time, not at container runtime, so VITE_GTM_ID needed the same builder-arg treatment already in place for the Paddle client token — first threaded through as a Kamal build arg, then through the Dockerfile itself. Without both, the production bundle would have shipped with an empty ID and the loader would have silently no-op'd, which is a particularly quiet way for a feature to not work.",
  },
  {
    type: 'h2',
    text: 'Also this release',
  },
  {
    type: 'p',
    text: "A Paddle webhook race could leave a Hobby org in a dead end: if a cancel_scheduled event arrived out of order or got retried after the org had already reverted to Hobby, the Billing page would keep hiding the upgrade options — reading the stale status as an active cancellation with nothing left to do. Fixed so the cancel_scheduled state is only trusted while the org is still on a paid plan; once it's back on Hobby, the upgrade options are always visible again.",
  },
  {
    type: 'h2',
    text: 'Follow along',
  },
  {
    type: 'p',
    text: 'Microsoft Teams alerts are still next on the board. Releases land on this blog as they ship; the GitHub repo has the full commit history and architecture decision records if you want the why behind any of this.',
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
