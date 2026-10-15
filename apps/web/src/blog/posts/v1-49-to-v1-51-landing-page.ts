import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "After the redesign, I asked for an outside review of the landing page. The verdict was fair: the top half felt technical, with a real crontab line leading to the alert it would trigger, but the second half went back to stock SaaS sections and undid some of that. These three releases are mostly about the second half.",
  },
  {
    type: 'h2',
    text: 'The testimonials are gone',
  },
  {
    type: 'p',
    text: "The review called the testimonials section the most generic part of the page: first names, last initials, stock-looking avatars, the kind of block every SaaS site has. A quote is only worth something if you can see who said it, so the section is gone, and it won't come back until it can link to the people quoted.",
  },
  {
    type: 'p',
    text: "In their place I tried real numbers: monitors running, pings received, checks run and outages caught across Checkmeup in the last 30 days. The first version showed all four, and it backfired. \"0 cron pings\" next to \"9 monitors\" honestly says this is a small, young service, and that is all a visitor takes away. Now a number only appears once it's big enough to mean something. Today that's just the checks.",
  },
  {
    type: 'h2',
    text: 'Show the payload, not a description',
  },
  {
    type: 'p',
    text: "Developers trust a payload more than a feature list. So under the monitor table, the page now shows exactly what a missed nightly backup sends: the Telegram message and the signed webhook request, with the same fields and wording the alert code produces. One detail that looks like a typo but isn't: the job runs at 03:00, and the alert says it was expected at 03:05. That's the five-minute grace period, already included in the expected time.",
  },
  {
    type: 'h2',
    text: 'A live strip that is actually live',
  },
  {
    type: 'p',
    text: "Under the hero sits checkmeup.net's own status for the last 90 days: not a picture, the same data its public status page shows. Each day has a tooltip with its date, status, uptime and average response time. The status page summary gained per-day uptime and latency for this. Above the strip is the uptime for the whole range. It's truncated rather than rounded, so 99.996% shows as 99.99%, never as a flattering 100%.",
  },
  {
    type: 'p',
    text: "The page is prerendered for search engines, without a backend, so it used to show an example strip until the real one loaded. That example ended on a red day with a \"Down\" label, which means every visitor briefly saw Checkmeup apparently down. It now shows a gray skeleton of the same size instead, and the real colors fill in once the data arrives.",
  },
  {
    type: 'h2',
    text: 'Pricing as one table',
  },
  {
    type: 'p',
    text: "The four pricing cards repeated the same two lines in every column. They're now one comparison table that lists only what differs between plans, with everything every plan gets in a single line underneath. Renaming the Enterprise tier came up too, and I deliberately left it alone for now: plan names are tied to billing, so that's a bigger change than a heading.",
  },
  {
    type: 'h2',
    text: 'Also these releases',
  },
  {
    type: 'ul',
    items: [
      'The logo in the prerendered page pointed at a development-only path, so search engines and anyone with JavaScript off got a broken image. Fixed.',
      'On phones, the crontab line wraps instead of being cut off. It wraps visually rather than with a backslash, because crontab has no line continuation and a backslash there would be wrong.',
      'On phones, the monitor table becomes stacked cards that keep every column, and the 90-day strip shows the last 30 days, since 90 ticks on a narrow screen come out as hairlines.',
      'The status badge has a copy-ready Markdown or HTML snippet next to it.',
      'The pricing page dropped "Simple, honest pricing" for something more specific, and the footer version number links to the changelog.',
      'v1.49 redrew the social preview image around the same cron line, missed-ping alert and 90-day strip.',
      "Reloading a page could scroll it to wherever you were the last time you switched tabs, rather than where you were when you reloaded. That turned out to be a bug in the router library version Checkmeup had pinned; updating it fixed it.",
    ],
  },
  {
    type: 'h2',
    text: 'Follow along',
  },
  {
    type: 'p',
    text: "The landing page will keep changing as there's more real data to show; the numbers line will fill in on its own as usage grows. Full commit history and every architecture decision record are on GitHub if you want the unfiltered version.",
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
