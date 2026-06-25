import type { BlogPost } from '../types'

export const post: BlogPost = {
  slug: 'v1-11-release-notes',
  title: 'v1.11: Alert Noise Reduction',
  date: 'June 25, 2026',
  readTime: '3 min read',
  excerpt:
    'Configure how many consecutive failures must occur before the first alert fires — across all four monitor types. Pair it with the existing per-incident alert cap to control both the start and the volume of notifications for any incident.',
  content: [
    {
      type: 'p',
      text: "v1.10 added Slack as an alert channel, making it easy to route alerts to more places. This release is about the opposite: giving you control over when alerts start and how many arrive per incident, so adding channels doesn't mean more noise.",
    },
    {
      type: 'h3',
      text: 'Alert after N failures',
    },
    {
      type: 'p',
      text: "Every monitor now has an 'Alert after N failures' setting. Set it to 0 (the default) and the first failure fires an alert — same behaviour as before. Set it to 1 and the first failure is silently recorded; the alert fires only if a second consecutive failure follows. Set it to 2 and the first two failures are swallowed; the alert fires on the third.",
    },
    {
      type: 'p',
      text: "The counter resets on every successful check, so a blip that resolves on its own never wakes anyone up. For uptime monitors, 'consecutive failures' is the existing state-machine counter already used to trigger alerts. For cron, SSL, and domain monitors, a new `consecutive_failures` column was added to track this independently of the alert state.",
    },
    {
      type: 'h3',
      text: 'Alert cap extended to SSL and domain monitors',
    },
    {
      type: 'p',
      text: "The per-incident alert cap — introduced in v1.5 for cron and uptime monitors — now covers SSL and domain monitors too. For expiry-based monitors, 'per incident' means per expiry cycle: the cap resets when the certificate is renewed or the domain registration is extended, not when the monitor itself is saved. Set it to 3, and you get at most three expiry warnings before renewal silences the monitor.",
    },
    {
      type: 'h3',
      text: 'How the two settings fit together',
    },
    {
      type: 'p',
      text: "'Alert after N failures' controls when the first alert fires. The per-incident cap controls how many alerts fire total. They compose: set N=1 and cap=5, and you skip the first failure, then get up to five alerts for any incident that persists past that point. Recovery alerts are always sent regardless of either setting — the cap and the filter apply only to down/failure notifications.",
    },
    {
      type: 'h3',
      text: 'Notification channel limits per plan',
    },
    {
      type: 'p',
      text: "Plan limits now include notification channels: Hobby 5, Solo 20, Startup 50, Enterprise 100. The limit is enforced at creation time — you'll see a clear error if you hit the ceiling. Existing channels are unaffected. The Billing page and the Pricing page both show the channel limit so the ceiling is visible before you reach it.",
    },
    {
      type: 'h3',
      text: 'Dashboard notification channels card',
    },
    {
      type: 'p',
      text: 'The dashboard now shows a notification channels summary card alongside the monitor and status page counts. It shows how many channels you have configured and links directly to Settings. A small addition, but it means the channel count is visible at a glance without navigating away.',
    },
    {
      type: 'h2',
      text: 'Reliability & data retention',
    },
    {
      type: 'p',
      text: 'A few infrastructure improvements shipped alongside the feature work. Uptime check records are now pruned after 90 days — history older than that is removed automatically in the background, keeping storage lean as the platform grows. The background worker now caps outbound checks at 50 concurrent goroutines per tick across uptime, SSL, and domain loops, so a spike in monitor count cannot exhaust connections. Incident list queries are bounded to 200 rows, and the public status page is rate-limited to 300 requests per minute per IP — both changes harden the API against accidental or deliberate load.',
    },
    {
      type: 'h2',
      text: 'Follow along',
    },
    {
      type: 'p',
      text: 'Microsoft Teams alerts are next. Releases land on this blog as they ship; the GitHub repo has the full commit history and architecture decision records.',
    },
    {
      type: 'signature',
      text: '— Andrew',
    },
  ],
}
