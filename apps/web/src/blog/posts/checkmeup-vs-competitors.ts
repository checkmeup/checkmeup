import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "These are the three products I researched most carefully before writing a line of Checkmeup's code, and they're the three that show up in almost every 'what do you use to monitor cron jobs / uptime' thread. Each one is good at what it was built for. This post is the comparison I wish existed when I was deciding what gap was actually still open — feature matrix, pricing at the tier people actually buy, and an honest list of what Checkmeup doesn't do yet.",
  },
  {
    type: 'h2',
    text: 'The three tools, in one line each',
  },
  {
    type: 'ul',
    items: [
      "Healthchecks.io — the deepest, most mature cron-job monitoring tool that exists. It does one thing (dead man's switches for scheduled jobs) and does it very well. It does not do uptime, SSL, domain, or port checks, and it does not have a real status page product — only public status badges.",
      'UptimeRobot — the default answer for "is my website up." Huge free tier (50 monitors), and it also covers SSL/domain expiry and TCP port checks. Cron-style heartbeat monitoring exists, but as an add-on on top of a product built around polling a URL, not the other way around.',
      'Cronitor — cron and uptime monitoring in one product, with real per-check assertions (including SSL certificate expiry conditions). The catch is the pricing model: monitors are $2/month each with no bundled tier, so the bill scales linearly with how much you watch.',
      'Checkmeup — cron, uptime, SSL, domain, and port monitoring in one flat-rate product, with white-label status pages included from the cheapest paid tier. Built specifically because none of the three above bundle all five monitor types at a flat, predictable price.',
    ],
  },
  {
    type: 'h2',
    text: 'What each product actually monitors',
  },
  {
    type: 'table',
    headers: ['Monitor type', 'Healthchecks.io', 'UptimeRobot', 'Cronitor', 'Checkmeup'],
    rows: [
      [
        'Cron / heartbeat monitoring',
        '✓ (core product)',
        '✓ (paid plans)',
        '✓ (core product)',
        '✓',
      ],
      ['Uptime (HTTP/keyword) monitoring', '—', '✓', '✓', '✓'],
      ['SSL certificate expiry', '—', '✓', '✓ (via assertions)', '✓'],
      ['Domain (WHOIS) expiry', '—', '✓', '—', '✓'],
      ['Port (TCP) monitoring', '—', '✓', '—', '✓'],
      [
        'White-label status pages',
        'Badges only',
        'From ~$33–38/mo tier',
        'Add-on, +$25–50/mo',
        'From $9/mo tier',
      ],
    ],
  },
  {
    type: 'p',
    text: "Healthchecks.io not doing uptime/SSL/domain/port isn't a knock on it — it's a deliberate, disciplined choice by a tool that's excellent at one job. The gap is real either way: if you run cron jobs and also want to know when a cert or a domain is about to expire, Healthchecks.io alone won't cover it. UptimeRobot and Cronitor both stretch across more monitor types, but neither bundles all five that Checkmeup does, and (see below) neither does it at a flat price once you're past the free tier.",
  },
  {
    type: 'h2',
    text: 'Pricing at the tier people actually buy',
  },
  {
    type: 'p',
    text: "Free tiers are easy to compare and rarely where the real decision gets made. The tier that matters is the cheapest paid one — what you hit the moment the free plan's monitor count runs out.",
  },
  {
    type: 'table',
    headers: [
      '',
      'Healthchecks.io Business',
      'UptimeRobot Solo',
      'Cronitor Business',
      'Checkmeup Solo',
    ],
    rows: [
      ['Price', '$20/mo', '~$9–10/mo', '$2 / monitor / mo', '$9/mo'],
      [
        'Monitors included',
        '100 (cron only)',
        '~50 (uptime only)',
        'none bundled',
        '30 (all 5 types combined)',
      ],
      ['Min. check interval', 'push-based, n/a', '60 sec', '30 sec', '1 min'],
      [
        'Status pages',
        'badges only',
        '3, not white-labeled',
        'branded costs extra',
        '3, white-labeled, included',
      ],
    ],
  },
  {
    type: 'p',
    text: "Cronitor's per-monitor pricing is the one worth doing the arithmetic on. At $2/monitor/month with no bundled allowance, watching the same 30 things Checkmeup's Solo plan covers for $9/month would run about $60/month on Cronitor — before branded status pages, which cost extra on top of that. That's not a criticism of Cronitor's product; the assertion-based checks it offers are genuinely more powerful than a simple uptime ping. It's a pricing-model difference: pay-per-monitor scales your bill with your infrastructure's growth, flat tiers don't.",
  },
  {
    type: 'h2',
    text: 'Where Checkmeup wins',
  },
  {
    type: 'ul',
    items: [
      'One product, five monitor types, one flat price — no other tool here bundles cron + uptime + SSL + domain + port without either leaving one out or charging per monitor for it.',
      'White-label status pages from the $9/mo tier, not gated behind a $33+/mo plan or a separate branded-page add-on — relevant specifically for agencies running status pages for multiple clients.',
      'Predictable cost as you grow: going from 10 to 30 to 100 monitors moves you between three fixed price points ($0 → $9 → $29), never a per-monitor multiplication.',
      'A public read-only API (X-API-Key auth) with per-ping metadata for CI pipelines and status LEDs — none of the three named here expose that combination the same way.',
      'SMS alerts with credits bundled into the flat plan price (Solo 10/mo, Startup 30/mo, Enterprise 100/mo) rather than a separate paid add-on or metered per-message charge — same flat-pricing philosophy as the monitor tiers above, applied to the one channel that actually has a real per-message cost behind it.',
    ],
  },
  {
    type: 'h2',
    text: "Where Checkmeup doesn't win yet",
  },
  {
    type: 'p',
    text: "Being straight about this matters more to me than the pitch above, so here's what's genuinely still a gap as of today:",
  },
  {
    type: 'ul',
    items: [
      'No voice-call alerts — UptimeRobot and Healthchecks.io both offer it as a more attention-grabbing escalation than SMS. Checkmeup has Telegram, email, Slack, webhooks, and SMS today; voice, WhatsApp, and Signal are on the roadmap but not shipped.',
      'No native mobile app — UptimeRobot has iOS/Android apps with push notifications. Checkmeup is a responsive web app; there is no dedicated mobile client.',
      'No two-factor authentication yet — UptimeRobot offers it on paid plans today. It is designed and queued on the Checkmeup roadmap but not live.',
      'No team/multi-user accounts yet — Cronitor and UptimeRobot both support additional seats. Checkmeup is single-user per organization for now; team management is planned but not built.',
    ],
  },
  {
    type: 'blockquote',
    text: "If you only run cron jobs and never touch a status page, Healthchecks.io is a mature, purpose-built tool and a perfectly good choice. If you only need uptime pings, UptimeRobot's free tier is hard to argue with. If you want deep per-check assertions and don't mind per-monitor billing, Cronitor. Checkmeup exists for the specific case those three don't cover well together: cron, uptime, SSL, domain, and port monitoring, plus a white-label status page, at one flat price that doesn't grow per monitor.",
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
