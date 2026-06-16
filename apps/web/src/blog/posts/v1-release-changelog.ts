import type { BlogPost } from '../types'

export const post: BlogPost = {
  slug: 'v1-release-changelog',
  title: "v1.0 — First Release: What Shipped and What's Next",
  date: 'June 16, 2026',
  readTime: '2 min read',
  excerpt:
    "checkmeup v1.0 is live. Here's the complete changelog, what we deliberately left out, and a foggy look at the road ahead.",
  content: [
    {
      type: 'p',
      text: 'checkmeup v1.0 shipped on June 16, 2026. This is the changelog.',
    },
    {
      type: 'h2',
      text: "What's in v1.0",
    },
    {
      type: 'h3',
      text: 'Cron job monitoring',
    },
    {
      type: 'p',
      text: 'Add a monitor, get a ping URL, call it at the end of your job. Miss a ping within the grace period and you get a Telegram alert. Recover and you get another. Execution history is stored so you can see exactly when jobs ran and for how long.',
    },
    {
      type: 'code',
      lang: 'bash',
      text: '# Add to the end of any cron job\ncurl -s https://checkmeup.net/ping/your-monitor-id',
    },
    {
      type: 'h3',
      text: 'Uptime monitoring',
    },
    {
      type: 'p',
      text: 'HTTP checks every 10 minutes. Alert on non-200 status or timeout. Response time tracked per check. Incident history with start/end times and duration.',
    },
    {
      type: 'h3',
      text: 'SSL expiry monitoring',
    },
    {
      type: 'p',
      text: 'Daily certificate checks. Alerts at 30, 14, and 7 days before expiry. Issuer and expiry date visible in the dashboard.',
    },
    {
      type: 'h3',
      text: 'Telegram alerts',
    },
    {
      type: 'p',
      text: 'Connect your Telegram account in settings. Alerts go to you directly — no shared channels, no noise from other users. Alert frequency is capped per incident (default: 3 alerts, then silence until recovery). Recovery alerts always send.',
    },
    {
      type: 'h3',
      text: 'Status pages',
    },
    {
      type: 'p',
      text: 'Public status pages at checkmeup.net/status/your-slug. Add any of your monitors to a page. Clients can bookmark it. No subdomain required — the same domain is intentional (avoids DNS setup overhead for every client).',
    },
    {
      type: 'h3',
      text: 'Maintenance windows',
    },
    {
      type: 'p',
      text: 'Schedule a maintenance window — or start one on the fly with no end date — covering any combination of cron, uptime, and SSL monitors. Covered monitors are skipped entirely while the window is active: no checks, no incidents, no alerts, and uptime stats stay untouched. Status pages show "Under maintenance" instead of up/down for the duration.',
    },
    {
      type: 'h3',
      text: 'Plans',
    },
    {
      type: 'ul',
      items: [
        'Hobby — Free — 10 monitors, 5-min checks, 1 status page',
        'Solo — $9/mo — 30 monitors, 1-min checks, 3 status pages',
        'Startup — $29/mo — 100 monitors, 1-min checks, 10 status pages',
        'Enterprise — $99/mo — 1000 monitors, 1-min checks, 100 status pages',
      ],
    },
    {
      type: 'h2',
      text: 'What we deliberately left out',
    },
    {
      type: 'p',
      text: "Email alerts, Slack integration, webhooks, API keys, and multi-user organizations are all missing from v1.0. That's intentional. Each of those is a real feature that would have added real complexity — and delays. Better to ship something that does the core well than to ship everything half-done.",
    },
    {
      type: 'p',
      text: "If one of those is blocking you from using checkmeup, email andrew@checkmeup.net and say so. That's how the roadmap gets prioritized.",
    },
    {
      type: 'h2',
      text: 'The foggy roadmap',
    },
    {
      type: 'p',
      text: 'These are the things most likely to happen next, in rough priority order. None of it is committed until it ships.',
    },
    {
      type: 'ul',
      items: [
        "Email alerts — for when you're not on Telegram",
        'Webhook alerts — POST to any URL on state change',
        'Public API — programmatic monitor creation and ping ingestion',
        'Slack integration — channel notifications for teams',
        'Multi-user organizations — invite teammates, shared monitors',
        'Check locations — run uptime checks from multiple regions',
      ],
    },
    {
      type: 'blockquote',
      text: 'The order will shift based on what users actually ask for. If you have strong feelings about any of these, reach out.',
    },
    {
      type: 'h2',
      text: 'Follow along',
    },
    {
      type: 'p',
      text: 'Releases and product updates go here on the blog. For faster updates, the Telegram channel (@checkmeup) gets announcements as they happen. The GitHub repo has the full commit history and all 19 architecture decision records if you want to see exactly how things are built.',
    },
    {
      type: 'p',
      text: "Thanks for being early. It means a lot — you're trusting a tool built by one person over a weekend-that-became-three-days to tell you when your stuff is on fire. I take that seriously.",
    },
    {
      type: 'signature',
      text: '— Andrew',
    },
  ],
}
