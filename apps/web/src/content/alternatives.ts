// The "<tool> alternative" pages at /alternatives/<slug> (PRD
// marketing-pages). Every competitor fact comes from that vendor's own pages,
// checked on `asOf` and listed in `sources`; recheck them all and move `asOf`
// when you change one. Checkmeup's side matches ADR billing and the PRDs.
export type AlternativeSlug = 'healthchecks-io' | 'uptimerobot' | 'cronitor'

export interface Alternative {
  slug: AlternativeSlug
  name: string
  seoTitle: string
  description: string
  title: string
  lead: string
  asOf: string
  rows: { label: string; checkmeup: string; them: string }[]
  reasons: string[]
  theyWin: string[]
  switching: string[]
  sources: { label: string; url: string }[]
}

export const alternatives: Alternative[] = [
  {
    slug: 'healthchecks-io',
    name: 'Healthchecks.io',
    seoTitle: 'Healthchecks.io Alternative with Uptime and Status Pages — Checkmeup',
    description:
      'Checkmeup vs Healthchecks.io: cron monitoring plus uptime, SSL, domain, port and DNS checks and client status pages, in one account. Free for 10 monitors.',
    title: 'A Healthchecks.io alternative that also watches your sites',
    lead: 'Healthchecks.io does cron monitoring well and only that. Checkmeup puts cron jobs, uptime, SSL and domain expiry, ports, DNS and public status pages in one account.',
    asOf: '2026-10-18',
    rows: [
      {
        label: 'Monitor types',
        checkmeup:
          'Cron, uptime (keyword and JSON checks), SSL expiry, domain expiry, TCP port, DNS record',
        them: 'Cron and heartbeat jobs only, pinged over HTTP or email',
      },
      {
        label: 'Free plan',
        checkmeup: '10 monitors of any type, 1 status page',
        them: '20 checks, 100 log entries each',
      },
      {
        label: 'First paid plan with more monitors',
        checkmeup: 'Solo, $9/mo: 30 monitors, 3 status pages, 10 SMS a month',
        them: 'Business, $20/mo: 100 checks, 50 SMS or WhatsApp and 20 phone calls a month',
      },
      {
        label: 'Status pages',
        checkmeup: 'On every plan; branding can be hidden from $9/mo',
        them: 'None; status badges only',
      },
      {
        label: 'Alerts on the free plan',
        checkmeup: 'Email, Telegram, Slack, webhook',
        them: 'Email, webhooks, Slack, Telegram, Discord and about 20 more',
      },
      {
        label: 'Team members',
        checkmeup: 'One user per account today',
        them: 'Unlimited on every plan',
      },
      {
        label: 'Source code',
        checkmeup: 'Public under the Business Source License 1.1',
        them: 'Open source under BSD-3-Clause, self-hostable',
      },
    ],
    reasons: [
      'One account for everything a client site needs watched: the backup job, the site itself, its certificate, its domain, its mail server port and its DNS.',
      'A public status page you can give a client, on the free plan too.',
      'SMS credits on every paid plan from $9 a month.',
    ],
    theyWin: [
      'You only need cron monitoring and want more free checks: 20 against 10.',
      'You want a failure signal from the job itself: Healthchecks.io takes /fail pings and exit codes, while Checkmeup only knows whether a ping arrived.',
      'You need many people in one account, phone-call alerts, or integrations such as PagerDuty, Opsgenie or Discord.',
      'You want to self-host under a permissive open-source license.',
    ],
    switching: [
      'Create a free Checkmeup account.',
      "For each check, create a cron monitor with the same schedule and grace period, then swap the hc-ping.com URL in your job for the monitor's ping URL. A /start ping maps to Checkmeup's /ping/{token}/start.",
      'Leave both running for a day or two, then delete the old check.',
    ],
    sources: [
      { label: 'Healthchecks.io pricing', url: 'https://healthchecks.io/pricing/' },
      { label: 'Healthchecks.io status badges', url: 'https://healthchecks.io/docs/badges/' },
      {
        label: 'Healthchecks.io: unlimited team sizes for all (2026-02)',
        url: 'https://blog.healthchecks.io/2026/02/unlimited-team-sizes-for-all/',
      },
      { label: 'Healthchecks.io self-hosting', url: 'https://healthchecks.io/docs/self_hosted/' },
    ],
  },
  {
    slug: 'uptimerobot',
    name: 'UptimeRobot',
    seoTitle: 'UptimeRobot Alternative with Cron and SSL Monitoring — Checkmeup',
    description:
      'Checkmeup vs UptimeRobot: cron, SSL and domain monitoring and webhook, Telegram and Slack alerts on the free plan, white-label status pages from $9.',
    title: 'An UptimeRobot alternative with cron and SSL checks on every plan',
    lead: "UptimeRobot's free plan watches many URLs; cron jobs, SSL and domain expiry, and webhook alerts take a paid plan. Checkmeup includes them all from the free plan up.",
    asOf: '2026-10-18',
    rows: [
      {
        label: 'Free plan',
        checkmeup: '10 monitors at 5 min: cron, uptime, SSL, domain, port, DNS',
        them: '50 monitors at 5 min: HTTP(S), keyword, ping, port',
      },
      {
        label: 'Cron, SSL and domain monitoring',
        checkmeup: 'Every plan',
        them: 'Paid plans',
      },
      {
        label: 'Cheapest paid plan',
        checkmeup: 'Solo, $9/mo: 30 monitors at 1 min',
        them: 'Solo, $10/mo: 10 monitors at 60 s ($22/mo for 50)',
      },
      {
        label: 'Webhook alerts',
        checkmeup: 'Every plan',
        them: 'Team, $41/mo, and up',
      },
      {
        label: 'Telegram and Slack alerts',
        checkmeup: 'Every plan',
        them: 'Solo and up',
      },
      {
        label: 'White-label status pages',
        checkmeup: 'From Solo, $9/mo',
        them: 'From Team, $41/mo',
      },
      {
        label: 'SMS',
        checkmeup: '10, 30 or 100 credits included each month on paid plans',
        them: 'Credits bought in packs; they do not renew with the plan',
      },
      {
        label: 'Fastest check, check locations',
        checkmeup: '1 min, from one location',
        them: '30 s on Team, 15 s on Scale; several regions on paid plans',
      },
    ],
    reasons: [
      'Cron jobs, certificates and domains are watched on the free plan, next to your sites.',
      'Webhook, Telegram and Slack alerts on every plan, not from $41 a month.',
      'Status pages without Checkmeup branding from $9 a month, for pages you give clients.',
    ],
    theyWin: [
      'You need more than 10 free monitors and only check URLs and ports.',
      'You need checks faster than a minute, or confirmation from several regions before an alert.',
      'You want ICMP ping or UDP checks, a mobile app, or several people in one account.',
    ],
    switching: [
      'Create a free Checkmeup account.',
      'Recreate each monitor: an uptime monitor for each HTTP or keyword monitor, a port monitor for each port monitor. There is no import yet.',
      'Add your alert channels, rebuild your status page, and run both for a day before switching off the old monitors.',
    ],
    sources: [
      { label: 'UptimeRobot pricing', url: 'https://uptimerobot.com/pricing/' },
      { label: 'UptimeRobot integrations', url: 'https://uptimerobot.com/integrations/' },
      { label: 'UptimeRobot terms', url: 'https://uptimerobot.com/terms/' },
    ],
  },
  {
    slug: 'cronitor',
    name: 'Cronitor',
    seoTitle: 'Cronitor Alternative with Flat Pricing — Checkmeup',
    description:
      'Checkmeup vs Cronitor: flat plans instead of $2 a monitor plus $5 a user, with domain expiry and DNS checks and white-label status pages from $9.',
    title: 'A Cronitor alternative with flat pricing',
    lead: "Cronitor charges per monitor and per user, so the bill grows with every job you add. Checkmeup's plans are flat: 30 monitors for $9 a month, 100 for $29.",
    asOf: '2026-10-18',
    rows: [
      {
        label: 'Pricing',
        checkmeup: 'Flat plans: $0, $9, $29, $99 a month',
        them: 'Business: $2 a monitor plus $5 a user, a month',
      },
      {
        label: 'Free plan',
        checkmeup: '10 monitors at 5 min, 1 status page',
        them: '5 monitors at 5 min, 1 basic status page',
      },
      {
        label: '30 monitors, one user',
        checkmeup: 'Solo, $9/mo',
        them: 'Business, $65/mo',
      },
      {
        label: 'Monitor types',
        checkmeup: 'Cron, uptime, SSL expiry, domain expiry, TCP port, DNS record',
        them: 'Jobs, heartbeats, HTTP, browser, TCP and UDP checks, with SSL expiry as an assertion',
      },
      {
        label: 'Status pages',
        checkmeup: '1 to 100 by plan; branding can be hidden from $9/mo',
        them: '1 included; a branded page $25/mo, a private one $50/mo',
      },
      {
        label: 'SMS alerts',
        checkmeup: 'Credits included on every paid plan',
        them: 'Business plan',
      },
      {
        label: 'Fastest check, check locations',
        checkmeup: '1 min, from one location',
        them: '30 s on paid plans, up to 12 locations',
      },
    ],
    reasons: [
      'A bill that stays the same as you add jobs: 30 monitors cost $9 a month, not $65.',
      'Domain expiry and DNS record checks, which Cronitor does not document.',
      'Status pages without branding included in the plan, not $25 a page.',
    ],
    theyWin: [
      'You want job telemetry: run metrics, exit codes, and SDKs or a CLI wrapper instead of a curl.',
      'You need checks every 30 seconds, from several locations, or browser checks.',
      'You need a team with roles, SSO or on-call integrations such as PagerDuty and Opsgenie.',
    ],
    switching: [
      'Create a free Checkmeup account.',
      "For each job, create a cron monitor with the same schedule and grace period, and have the job call the monitor's ping URL with curl where it called Cronitor. A run's start maps to /ping/{token}/start.",
      'Recreate your HTTP checks as uptime monitors, run both for a day, then remove the Cronitor monitors.',
    ],
    sources: [
      { label: 'Cronitor pricing', url: 'https://cronitor.io/pricing' },
      {
        label: 'Cronitor uptime monitoring docs',
        url: 'https://cronitor.io/docs/uptime-monitoring',
      },
      { label: 'Cronitor status page docs', url: 'https://cronitor.io/docs/status-pages' },
    ],
  },
]
