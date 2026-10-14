import type {
  CronMonitor,
  DNSMonitor,
  DomainMonitor,
  PortMonitor,
  SSLMonitor,
  UptimeMonitor,
} from '@/api/monitors'

// One row per monitor of any type — the shape the dashboard and the
// /monitors list both render, so every type reads the same in a table.

export type MonitorType = 'cron' | 'uptime' | 'ssl' | 'domain' | 'port' | 'dns'

export const monitorTypes: MonitorType[] = ['cron', 'uptime', 'ssl', 'domain', 'port', 'dns']

// How a status reads at a glance: the color and the sort group it falls in.
export type Tone = 'down' | 'warn' | 'waiting' | 'up' | 'paused'

export interface MonitorRow {
  key: string
  type: MonitorType
  id: string
  name: string
  target: string
  // Type-specific second line under the target (keyword rule, expected
  // state, check error) — null when there's nothing to add.
  note: string | null
  status: string
  statusLabel: string
  tone: Tone
  metric: string
  metricSub: string
  lastChecked: string
  daysUntilExpiry: number | null
  // Uptime over the last 24h for monitors that track it (uptime, port, DNS).
  uptime24h: number | null
}

export const typeLabel: Record<MonitorType, string> = {
  cron: 'Cron',
  uptime: 'Uptime',
  ssl: 'SSL',
  domain: 'Domain',
  port: 'Port',
  dns: 'DNS',
}

export const detailRouteName: Record<MonitorType, string> = {
  cron: 'cron-monitor-detail',
  uptime: 'uptime-monitor-detail',
  ssl: 'ssl-monitor-detail',
  domain: 'domain-monitor-detail',
  port: 'port-monitor-detail',
  dns: 'dns-monitor-detail',
}

export const createRouteName: Record<MonitorType, string> = {
  cron: 'cron-monitor-create',
  uptime: 'uptime-monitor-create',
  ssl: 'ssl-monitor-create',
  domain: 'domain-monitor-create',
  port: 'port-monitor-create',
  dns: 'dns-monitor-create',
}

export const toneColor: Record<Tone, string> = {
  down: 'var(--status-down)',
  warn: 'var(--status-degraded)',
  waiting: 'var(--text-muted)',
  up: 'var(--status-up)',
  paused: 'var(--status-paused)',
}

const toneOrder: Record<Tone, number> = { down: 0, warn: 1, waiting: 2, up: 3, paused: 4 }

export function toneOf(status: string): Tone {
  if (status === 'down' || status === 'error' || status === 'expired') return 'down'
  if (status === 'expiring_soon') return 'warn'
  if (status === 'up') return 'up'
  if (status === 'paused') return 'paused'
  return 'waiting'
}

const statusLabels: Record<string, string> = {
  up: 'Up',
  down: 'Down',
  expired: 'Expired',
  error: 'Error',
  expiring_soon: 'Expiring soon',
  waiting: 'Waiting',
  paused: 'Paused',
}

export function statusLabel(type: MonitorType, status: string): string {
  // A certificate or registration that's fine is "valid", not "up".
  if (status === 'up' && (type === 'ssl' || type === 'domain')) return 'Valid'
  return statusLabels[status] ?? status
}

export function relativeTime(iso: string | null, now = Date.now()): string {
  if (!iso) return '—'
  const diff = now - new Date(iso).getTime()
  const m = Math.floor(Math.abs(diff) / 60000)
  const h = Math.floor(m / 60)
  const magnitude = m < 60 ? `${m}m` : h < 24 ? `${h}h` : `${Math.floor(h / 24)}d`
  if (diff >= 0) return m < 1 ? 'just now' : `${magnitude} ago`
  return m < 1 ? 'in <1m' : `in ${magnitude}`
}

export function fmtPct(v: number | null): string {
  if (v === null || v === undefined) return '—'
  return `${v.toFixed(2)}%`
}

export function fmtDaysLeft(days: number | null): string {
  if (days === null) return '—'
  if (days < 0) return 'Expired'
  return days === 1 ? '1 day' : `${days} days`
}

function fmtDate(iso: string | null): string {
  return iso ? new Date(iso).toLocaleDateString() : ''
}

function truncate(s: string, n: number) {
  return s.length > n ? s.slice(0, n) + '…' : s
}

function base(type: MonitorType, id: string, name: string, status: string) {
  return {
    key: `${type}:${id}`,
    type,
    id,
    name,
    status,
    statusLabel: statusLabel(type, status),
    tone: toneOf(status),
  }
}

export function cronRow(m: CronMonitor, now = Date.now()): MonitorRow {
  return {
    ...base('cron', m.id, m.name, m.status),
    target: m.schedule,
    note: null,
    metric: relativeTime(m.lastPingAt, now),
    metricSub: m.nextPingAt ? `next ping ${relativeTime(m.nextPingAt, now)}` : 'last ping',
    lastChecked: relativeTime(m.lastPingAt, now),
    daysUntilExpiry: null,
    uptime24h: null,
  }
}

export function keywordRule(m: Pick<UptimeMonitor, 'keyword' | 'keywordMode'>): string | null {
  if (!m.keyword) return null
  const verb = m.keywordMode === 'not_contains' ? 'does not contain' : 'contains'
  return `${verb} "${truncate(m.keyword, 24)}"`
}

export function uptimeRow(m: UptimeMonitor, now = Date.now()): MonitorRow {
  return {
    ...base('uptime', m.id, m.name, m.status),
    target: m.url,
    note: keywordRule(m),
    metric: fmtPct(m.uptime24h),
    metricSub: 'uptime 24h',
    lastChecked: relativeTime(m.lastCheckedAt, now),
    daysUntilExpiry: null,
    uptime24h: m.uptime24h,
  }
}

export function sslRow(m: SSLMonitor, now = Date.now()): MonitorRow {
  return {
    ...base('ssl', m.id, m.name, m.status),
    target: m.hostname,
    note: m.status === 'error' ? m.errorMsg : null,
    metric: fmtDaysLeft(m.daysUntilExpiry),
    metricSub: [fmtDate(m.expiresAt), m.issuer].filter(Boolean).join(' · '),
    lastChecked: relativeTime(m.lastCheckedAt, now),
    daysUntilExpiry: m.daysUntilExpiry,
    uptime24h: null,
  }
}

export function domainRow(m: DomainMonitor, now = Date.now()): MonitorRow {
  return {
    ...base('domain', m.id, m.name, m.status),
    target: m.domain,
    note: m.status === 'error' ? m.errorMsg : null,
    metric: fmtDaysLeft(m.daysUntilExpiry),
    metricSub: [fmtDate(m.expiresAt), m.registrar].filter(Boolean).join(' · '),
    lastChecked: relativeTime(m.lastCheckedAt, now),
    daysUntilExpiry: m.daysUntilExpiry,
    uptime24h: null,
  }
}

export function portRow(m: PortMonitor, now = Date.now()): MonitorRow {
  return {
    ...base('port', m.id, m.name, m.status),
    target: `${m.host}:${m.port}`,
    note: m.expectedState === 'closed' ? 'expects the port closed' : null,
    metric: fmtPct(m.uptime24h),
    metricSub: 'uptime 24h',
    lastChecked: relativeTime(m.lastCheckedAt, now),
    daysUntilExpiry: null,
    uptime24h: m.uptime24h,
  }
}

export function dnsRow(m: DNSMonitor, now = Date.now()): MonitorRow {
  return {
    ...base('dns', m.id, m.name, m.status),
    target: `${m.hostname} ${m.recordType}`,
    note: null,
    metric: m.lastResolvedValue ? truncate(m.lastResolvedValue, 28) : '—',
    metricSub: `${m.recordType} record`,
    lastChecked: relativeTime(m.lastCheckedAt, now),
    daysUntilExpiry: null,
    uptime24h: m.uptime24h,
  }
}

// Problems first (down, then warnings), then by name — the order both the
// dashboard and /monitors show.
export function sortRows(rows: MonitorRow[]): MonitorRow[] {
  return [...rows].sort(
    (a, b) => toneOrder[a.tone] - toneOrder[b.tone] || a.name.localeCompare(b.name),
  )
}

export function matchesSearch(row: MonitorRow, query: string): boolean {
  const q = query.trim().toLowerCase()
  if (!q) return true
  return row.name.toLowerCase().includes(q) || row.target.toLowerCase().includes(q)
}

// The sentence the dashboard's "Needs attention" list shows for a row.
export function attentionSentence(row: MonitorRow): string {
  switch (row.status) {
    case 'down':
      return row.type === 'cron'
        ? `Missed its expected ping — last one ${row.lastChecked}.`
        : `Down — the last check of ${row.target} failed.`
    case 'error':
      return row.note ? `Check failed: ${row.note}` : `The last check of ${row.target} failed.`
    case 'expired':
      return row.type === 'ssl'
        ? `The certificate for ${row.target} has expired.`
        : `The registration for ${row.target} has expired.`
    case 'expiring_soon':
      return row.type === 'ssl'
        ? `Certificate expires in ${fmtDaysLeft(row.daysUntilExpiry)}.`
        : `Registration expires in ${fmtDaysLeft(row.daysUntilExpiry)}.`
    default:
      return row.statusLabel
  }
}

// How long an incident lasted, or "Ongoing" while it's open.
export function incidentDuration(start: string, end: string | null): string {
  if (!end) return 'Ongoing'
  const mins = Math.floor((new Date(end).getTime() - new Date(start).getTime()) / 60000)
  if (mins < 60) return `${mins}m`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ${mins % 60}m`
  return `${Math.floor(hrs / 24)}d ${hrs % 24}h`
}

// The /monitors list with this type's tab selected — where a monitor page's
// back link and post-delete redirect go.
export function monitorsListRoute(type: MonitorType) {
  return { name: 'monitors', query: { type } }
}
