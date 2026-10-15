// The status page mockup: the kind of monitors a freelancer puts on a
// client's page.
export const statusRows = [
  { name: 'acme-shop.com', pct: '100%' },
  { name: 'api.client.io', pct: '98%' },
  { name: 'backup: nightly', pct: '99%' },
]

export interface PricingPlan {
  name: string
  price: number
  description: string
  highlight: boolean
  cta: string
}

export const plans: PricingPlan[] = [
  {
    name: 'Hobby',
    price: 0,
    description: 'Your own side projects',
    highlight: false,
    cta: 'Start free',
  },
  {
    name: 'Solo',
    price: 9,
    description: 'A freelancer with a few clients',
    highlight: false,
    cta: 'Start Solo',
  },
  {
    name: 'Startup',
    price: 29,
    description: 'A small agency',
    highlight: true,
    cta: 'Start Startup',
  },
  {
    name: 'Enterprise',
    price: 99,
    description: 'An agency with a page per client',
    highlight: false,
    cta: 'Start Enterprise',
  },
]

// The landing page's plan comparison: only what differs between plans, one
// value per plan in `plans` order. Values match PricingView's featureRows —
// keep the two in sync.
export const planRows = [
  { label: 'Monitors', values: ['10', '30', '100', '1000'] },
  { label: 'Fastest check', values: ['5 min', '1 min', '1 min', '1 min'] },
  { label: 'Status pages', values: ['1', '3', '10', '100'] },
  { label: 'White-label status pages', values: ['—', '✓', '✓', '✓'] },
  { label: 'Notification channels', values: ['5', '20', '50', '100'] },
  { label: 'SMS credits / month', values: ['—', '10', '30', '100'] },
]

// What every plan gets, so the comparison above doesn't repeat it per column.
export const everyPlan =
  'Every plan: all six monitor types, execution logs, incidents, maintenance windows, and Telegram, email, Slack and webhook alerts.'

// The landing page's spec sheet of monitor types. Intervals and thresholds
// match DocsView — keep the two in sync. `id` is the row's anchor, which the
// hero's eyebrow links to.
export const monitorTypes = [
  {
    id: 'cron',
    name: 'Cron job monitoring',
    watches: 'A ping your job sends when it finishes',
    often: 'your schedule',
    oftenNote: 'plus a grace period',
    alerts: 'A ping is later than the grace period you chose, or a started run never finishes',
  },
  {
    id: 'uptime',
    name: 'Uptime monitoring',
    watches: 'A URL: its status code, a keyword, or JSON fields',
    often: '1 min',
    oftenNote: '5 min on Hobby',
    alerts: 'The response is an error, too slow, or missing what you expect',
  },
  {
    id: 'ssl',
    name: 'SSL expiry monitoring',
    watches: 'The certificate on port 443',
    often: 'daily',
    oftenNote: null,
    alerts: 'It expires in 30, 14 or 7 days, or is already invalid',
  },
  {
    id: 'domain',
    name: 'Domain expiry monitoring',
    watches: "The domain's registration",
    often: 'daily',
    oftenNote: null,
    alerts: 'It expires in 30, 14 or 7 days, or already has',
  },
  {
    id: 'port',
    name: 'Port (TCP) monitoring',
    watches: 'A raw TCP connect to host:port',
    often: '1 min',
    oftenNote: '5 min on Hobby',
    alerts: 'The port stops answering — or, for one that should stay closed, starts',
  },
  {
    id: 'dns',
    name: 'DNS record monitoring',
    watches: 'An A, AAAA, CNAME, MX, TXT or NS record',
    often: '1 min',
    oftenNote: '5 min on Hobby',
    alerts: 'The record changes or stops resolving',
  },
]

// What a missed ping actually sends, as the landing page shows it under the
// spec table. The text mirrors buildOverdueAlert and buildAlert in
// apps/api/internal/worker — keep them in sync. The expected time already
// includes the grace period, so a 03:00 job with 5 minutes' grace is
// expected at 03:05.
export const sampleTelegram = `🔴 nightly-backup missed its ping

Schedule: 0 3 * * *
Expected at: 03:05:00 UTC
Missed by: 18s`

export const sampleWebhook = `POST /hooks/checkmeup HTTP/1.1
Content-Type: application/json
X-Checkmeup-Signature: 7c1e9a…f03b

{
  "eventType": "down",
  "monitorName": "nightly-backup",
  "monitorType": "cron",
  "reason": "missed its ping — expected at 03:05:00 UTC, missed by 18s",
  "timestamp": "2026-10-15T03:05:18Z"
}`

// The production status page of checkmeup.net itself; the landing hero shows
// its live 90-day strip.
export const ownStatusSlug = 'checkmeup-net'
