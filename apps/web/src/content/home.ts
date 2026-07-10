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
    extras: ['Cron, uptime, SSL, domain & port monitors', 'Telegram, Slack & email alerts'],
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
    extras: ['Cron, uptime, SSL, domain & port monitors', 'Telegram, Slack, email & SMS alerts'],
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
    extras: ['Cron, uptime, SSL, domain & port monitors', 'Telegram, Slack, email & SMS alerts'],
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
    extras: ['Cron, uptime, SSL, domain & port monitors', 'Telegram, Slack, email & SMS alerts'],
  },
]

export const customers = [
  {
    name: 'Alex M.',
    role: 'DevOps Engineer',
    avatar: 'img/customer1.png',
  },
  {
    name: 'Priya S.',
    role: 'Full-Stack Dev',
    avatar: 'img/customer2.png',
  },
  {
    name: 'James K.',
    role: 'Agency CTO',
    avatar: 'img/customer3.png',
  },
  {
    name: 'Layla H.',
    role: 'Startup Founder',
    avatar: 'img/customer4.png',
  },
  {
    name: 'Dan W.',
    role: 'Backend Engineer',
    avatar: 'img/customer5.png',
  },
  {
    name: 'Zoe C.',
    role: 'Platform Engineer',
    avatar: 'img/customer6.png',
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
