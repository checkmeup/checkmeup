import type { BlogPost } from '../types'

export const post: BlogPost = {
  slug: 'v1-8-release-notes',
  title: 'v1.8: Domain Expiry Monitoring',
  date: 'June 23, 2026',
  readTime: '3 min read',
  excerpt:
    'A fourth monitor type alongside cron, uptime, and SSL: domain registration expiry. checkmeup already told you when a certificate was about to lapse — now it tells you when the domain underneath it is about to disappear entirely, which is rarer but a lot worse when it happens.',
  content: [
    {
      type: 'p',
      text: "SSL monitoring (shipped back in EP-04) watches a certificate's expiry. That's a different failure from the domain registration itself lapsing — a cert renews on its own schedule, but if the domain isn't renewed, the whole site goes dark and the name is sometimes snapped up by a squatter before the owner notices. Domain monitors close that gap: same shape as SSL monitoring, different data source.",
    },
    {
      type: 'h3',
      text: 'Adding a domain monitor',
    },
    {
      type: 'p',
      text: "Add a domain the same way you'd add any other monitor: give it a name and an apex domain (example.com — same validation pattern as SSL's hostname field), and the first lookup runs immediately on creation. The list view shows every domain at a glance — status (valid, expiring soon, expired, or error), expiry date, and days remaining. The detail view adds registrar and last-checked time, plus the error message if the most recent lookup failed.",
    },
    {
      type: 'p',
      text: 'Domain monitors count toward the org\'s aggregate monitor limit alongside cron, uptime, and SSL — the plan limits table now reads "cron + uptime + SSL + domain," same enforcement as the other three. They also plug into maintenance windows and the monitor picker used there, so a planned domain transfer doesn\'t fire false alerts.',
    },
    {
      type: 'h3',
      text: 'How the lookup works',
    },
    {
      type: 'p',
      text: 'Registration data comes from RDAP (RFC 9082/9083) — the structured-JSON successor to WHOIS — via rdap.org\'s public bootstrap redirector, which resolves the right authoritative RDAP server per TLD without checkmeup having to maintain its own IANA bootstrap mapping. Each domain is checked once a day, same cadence as SSL. A lookup failure — registry timeout, an unsupported TLD, a response with no expiration date — is recorded as an error state, never reported as "expired" when the real status is simply unknown.',
    },
    {
      type: 'h3',
      text: 'Alerts',
    },
    {
      type: 'p',
      text: 'Alerts fire at 30, 14, and 7 days before expiry — one alert per threshold, no repeats at the same one — and immediately if a domain is already expired. Same debounce pattern as SSL monitoring, delivered over whichever channels (Telegram, email, webhook) the monitor has attached.',
    },
    {
      type: 'h3',
      text: "What's not in this release",
    },
    {
      type: 'p',
      text: "Two pieces of the original plan got cut to ship this without rushing them. RDAP doesn't cover every TLD, and the WHOIS text-parsing fallback for the ones it doesn't — which needs a separate parser per registry's response format — wasn't built; those domains report as an error rather than silently guessing. And the RDAP client doesn't parse a domain's status array yet, so an immediate alert for a registry hold or pending-delete state isn't wired up — only the expiration date and registrar are read out of the response today. Both are reasonable follow-ups if a real domain hits either gap.",
    },
    {
      type: 'h2',
      text: 'Follow along',
    },
    {
      type: 'p',
      text: 'Public status badges and assertion-based API checks are next up. Releases land on this blog as they ship; the GitHub repo has the full commit history and architecture decision records if you want the why behind any of this.',
    },
    {
      type: 'signature',
      text: '— Andrew',
    },
  ],
}
