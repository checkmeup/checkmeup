export const statusRows = [
  { name: 'Hourly Cron Monitor', pct: '98%' },
  { name: 'Hourly Cron Monitor Ping', pct: '100%' },
  { name: 'checkmeup.net', pct: '99%' },
]

export interface PricingPlan {
  name: string
  price: number
  description: string
  highlight: boolean
  cta: string
  stats: { value: string; label: string }[]
  extras: string[]
}

export const plans: PricingPlan[] = [
  {
    name: 'Hobby',
    price: 0,
    description: 'For personal projects',
    highlight: false,
    cta: 'Get started free',
    stats: [
      { value: '10', label: 'monitors' },
      { value: '5 min', label: 'check interval' },
      { value: '1', label: 'status page' },
      { value: '5', label: 'notification channels' },
    ],
    extras: ['Cron, uptime, SSL, domain, port & DNS monitors', 'Telegram, Slack & email alerts'],
  },
  {
    name: 'Solo',
    price: 9,
    description: 'For solo builders',
    highlight: false,
    cta: 'Start Solo',
    stats: [
      { value: '30', label: 'monitors' },
      { value: '1 min', label: 'check interval' },
      { value: '3', label: 'status pages' },
      { value: '20', label: 'notification channels' },
      { value: '10', label: 'SMS credits / month' },
    ],
    extras: [
      'Cron, uptime, SSL, domain, port & DNS monitors',
      'Telegram, Slack, email & SMS alerts',
    ],
  },
  {
    name: 'Startup',
    price: 29,
    description: 'For small agencies',
    highlight: true,
    cta: 'Start Startup',
    stats: [
      { value: '100', label: 'monitors' },
      { value: '1 min', label: 'check interval' },
      { value: '10', label: 'status pages' },
      { value: '50', label: 'notification channels' },
      { value: '30', label: 'SMS credits / month' },
    ],
    extras: [
      'Cron, uptime, SSL, domain, port & DNS monitors',
      'Telegram, Slack, email & SMS alerts',
    ],
  },
  {
    name: 'Enterprise',
    price: 99,
    description: 'For growing agencies',
    highlight: false,
    cta: 'Start Enterprise',
    stats: [
      { value: '1000', label: 'monitors' },
      { value: '1 min', label: 'check interval' },
      { value: '100', label: 'status pages' },
      { value: '100', label: 'notification channels' },
      { value: '100', label: 'SMS credits / month' },
    ],
    extras: [
      'Cron, uptime, SSL, domain, port & DNS monitors',
      'Telegram, Slack, email & SMS alerts',
    ],
  },
]

// The landing page's spec sheet of monitor types. Intervals and thresholds
// match DocsView — keep the two in sync.
export const monitorTypes = [
  {
    name: 'Cron job monitoring',
    watches: 'A ping your job sends when it finishes',
    often: 'your schedule',
    oftenNote: 'plus a grace period',
    alerts: 'A ping is later than the grace period you chose, or a started run never finishes',
  },
  {
    name: 'Uptime monitoring',
    watches: 'A URL: its status code, a keyword, or JSON fields',
    often: '1 min',
    oftenNote: '5 min on Hobby',
    alerts: 'The response is an error, too slow, or missing what you expect',
  },
  {
    name: 'SSL expiry monitoring',
    watches: 'The certificate on port 443',
    often: 'daily',
    oftenNote: null,
    alerts: 'It expires in 30, 14 or 7 days, or is already invalid',
  },
  {
    name: 'Domain expiry monitoring',
    watches: "The domain's registration",
    often: 'daily',
    oftenNote: null,
    alerts: 'It expires in 30, 14 or 7 days, or already has',
  },
  {
    name: 'Port (TCP) monitoring',
    watches: 'A raw TCP connect to host:port',
    often: '1 min',
    oftenNote: '5 min on Hobby',
    alerts: 'The port stops answering — or, for one that should stay closed, starts',
  },
  {
    name: 'DNS record monitoring',
    watches: 'An A, AAAA, CNAME, MX, TXT or NS record',
    often: '1 min',
    oftenNote: '5 min on Hobby',
    alerts: 'The record changes or stops resolving',
  },
]

export const testimonials = [
  {
    name: 'Sarah K.',
    role: 'Backend Engineer',
    avatar: 'img/avatar1.png',
    quote:
      'My nightly backup job silently failed for two weeks before I set up Checkmeup. Now I get a Telegram ping within minutes if anything misses its window.',
  },
  {
    name: 'Marcus T.',
    role: 'Freelance Web Developer',
    avatar: 'img/avatar2.png',
    quote:
      "I maintain 25+ client sites solo. Having a branded status page for each one changed how my clients perceive me. They feel taken care of, and I don't get 3am calls.",
  },
  {
    name: 'Gracy R.',
    role: 'Indie Developer',
    avatar: 'img/avatar3.png',
    quote:
      "Woke up to a Slack alert that my SSL cert was expiring in 7 days. I'd completely forgotten about it. Renewed it in 10 minutes. Crisis averted.",
  },
]
