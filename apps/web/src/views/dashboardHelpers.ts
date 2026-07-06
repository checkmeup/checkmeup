export type MonitorType = 'cron' | 'uptime' | 'ssl' | 'domain' | 'port'

export interface Row {
  key: string
  type: MonitorType
  id: string
  name: string
  target: string
  status: string
  metricLabel: string
  metricValue: string
  lastChecked: string
}

export const typeLabel: Record<MonitorType, string> = {
  cron: 'Cron',
  uptime: 'Uptime',
  ssl: 'SSL',
  domain: 'Domain',
  port: 'Port',
}

export const detailRouteName: Record<MonitorType, string> = {
  cron: 'cron-monitor-detail',
  uptime: 'uptime-monitor-detail',
  ssl: 'ssl-monitor-detail',
  domain: 'domain-monitor-detail',
  port: 'port-monitor-detail',
}

export const statusColors: Record<string, string> = {
  up: 'var(--status-up)',
  down: 'var(--status-down)',
  expired: 'var(--status-down)',
  error: 'var(--status-down)',
  expiring_soon: 'var(--status-degraded)',
  waiting: 'var(--text-muted)',
  paused: 'var(--status-paused)',
}

export const statusLabels: Record<string, string> = {
  up: 'Up',
  down: 'Down',
  expired: 'Expired',
  error: 'Error',
  expiring_soon: 'Expiring soon',
  waiting: 'Waiting',
  paused: 'Paused',
}

export function relativeTime(iso: string | null): string {
  if (!iso) return '—'
  const diff = Date.now() - new Date(iso).getTime()
  const abs = Math.abs(diff)
  const m = Math.floor(abs / 60000)
  const h = Math.floor(m / 60)
  const magnitude = m < 60 ? `${m}m` : h < 24 ? `${h}h` : `${Math.floor(h / 24)}d`
  if (diff >= 0) return m < 1 ? 'just now' : `${magnitude} ago`
  return m < 1 ? 'in <1m' : `in ${magnitude}`
}

export function fmtPct(v: number | null): string {
  if (v === null) return '—'
  return `${v.toFixed(2)}%`
}

export function fmtExpiry(daysUntilExpiry: number | null): string {
  if (daysUntilExpiry === null) return '—'
  if (daysUntilExpiry < 0) return 'Expired'
  return `${daysUntilExpiry}d`
}

export function buildRow(
  type: MonitorType,
  id: string,
  name: string,
  target: string,
  status: string,
  metricLabel: string,
  metricValue: string,
  lastChecked: string,
): Row {
  return {
    key: `${type}:${id}`,
    type,
    id,
    name,
    target,
    status,
    metricLabel,
    metricValue,
    lastChecked,
  }
}
