import { api } from './client'

export interface CronMonitor {
  id: string
  name: string
  schedule: string
  gracePeriodMins: number
  pingToken: string
  pingUrl: string
  status: 'waiting' | 'up' | 'down' | 'paused'
  alertsEnabled: boolean
  maxAlertsPerIncident: number
  lastPingAt: string | null
  nextPingAt: string | null
  createdAt: string
}

export interface CronPing {
  id: string
  receivedAt: string
  sourceIp: string
}

export interface CronIncident {
  id: string
  startedAt: string
  resolvedAt: string | null
}

export interface CronMonitorDetail {
  monitor: CronMonitor
  pings: CronPing[]
  incidents: CronIncident[]
}

export interface CreateCronMonitorInput {
  name: string
  schedule: string
  gracePeriodMins: number
  maxAlertsPerIncident: number
}

export interface UpdateCronMonitorInput {
  name: string
  schedule: string
  gracePeriodMins: number
  alertsEnabled: boolean
  maxAlertsPerIncident: number
}

export interface UptimeMonitor {
  id: string
  name: string
  url: string
  intervalMins: number
  status: 'waiting' | 'up' | 'down' | 'paused'
  alertsEnabled: boolean
  maxAlertsPerIncident: number
  lastCheckedAt: string | null
  createdAt: string
  uptime24h: number | null
}

export interface UptimeCheck {
  id: string
  checkedAt: string
  statusCode: number | null
  responseTimeMs: number
  isUp: boolean
}

export interface UptimeIncident {
  id: string
  startedAt: string
  resolvedAt: string | null
}

export interface UptimeStats {
  uptime24h: number | null
  uptime7d: number | null
  uptime30d: number | null
}

export interface UptimeMonitorDetail {
  monitor: UptimeMonitor
  chartData: UptimeCheck[]
  checks: UptimeCheck[]
  incidents: UptimeIncident[]
  stats: UptimeStats
}

export interface CreateUptimeMonitorInput {
  name: string
  url: string
  intervalMins: number
  maxAlertsPerIncident: number
}

export interface UpdateUptimeMonitorInput {
  name: string
  url: string
  intervalMins: number
  alertsEnabled: boolean
  maxAlertsPerIncident: number
}

export interface SSLMonitor {
  id: string
  name: string
  hostname: string
  status: 'waiting' | 'up' | 'expiring_soon' | 'expired' | 'error' | 'paused'
  alertsEnabled: boolean
  expiresAt: string | null
  issuer: string | null
  errorMsg: string | null
  daysUntilExpiry: number | null
  lastCheckedAt: string | null
  createdAt: string
}

export interface CreateSSLMonitorInput {
  name: string
  hostname: string
}

export interface UpdateSSLMonitorInput {
  name: string
  hostname: string // passed through but not shown in UI (domain changes require delete + recreate)
  alertsEnabled: boolean
}

export const monitorsApi = {
  listCron: () => api.get<CronMonitor[]>('/api/v1/monitors/cron/'),
  getCron: (id: string) => api.get<CronMonitorDetail>(`/api/v1/monitors/cron/${id}/`),
  createCron: (input: CreateCronMonitorInput) =>
    api.post<CronMonitor>('/api/v1/monitors/cron/', input),
  updateCron: (id: string, input: UpdateCronMonitorInput) =>
    api.patch<CronMonitor>(`/api/v1/monitors/cron/${id}/`, input),
  pauseCron: (id: string) => api.post<CronMonitor>(`/api/v1/monitors/cron/${id}/pause`, {}),
  resumeCron: (id: string) => api.post<CronMonitor>(`/api/v1/monitors/cron/${id}/resume`, {}),
  deleteCron: (id: string) => api.delete<void>(`/api/v1/monitors/cron/${id}/`),
  getCronPings: (id: string, page = 1) =>
    api.get<CronPing[]>(`/api/v1/monitors/cron/${id}/pings?page=${page}`),

  listSSL: () => api.get<SSLMonitor[]>('/api/v1/monitors/ssl/'),
  getSSL: (id: string) => api.get<SSLMonitor>(`/api/v1/monitors/ssl/${id}/`),
  createSSL: (input: CreateSSLMonitorInput) => api.post<SSLMonitor>('/api/v1/monitors/ssl/', input),
  updateSSL: (id: string, input: UpdateSSLMonitorInput) =>
    api.patch<SSLMonitor>(`/api/v1/monitors/ssl/${id}/`, input),
  pauseSSL: (id: string) => api.post<SSLMonitor>(`/api/v1/monitors/ssl/${id}/pause`, {}),
  resumeSSL: (id: string) => api.post<SSLMonitor>(`/api/v1/monitors/ssl/${id}/resume`, {}),
  deleteSSL: (id: string) => api.delete<void>(`/api/v1/monitors/ssl/${id}/`),

  listUptime: () => api.get<UptimeMonitor[]>('/api/v1/monitors/uptime/'),
  getUptime: (id: string) => api.get<UptimeMonitorDetail>(`/api/v1/monitors/uptime/${id}/`),
  createUptime: (input: CreateUptimeMonitorInput) =>
    api.post<UptimeMonitor>('/api/v1/monitors/uptime/', input),
  updateUptime: (id: string, input: UpdateUptimeMonitorInput) =>
    api.patch<UptimeMonitor>(`/api/v1/monitors/uptime/${id}/`, input),
  pauseUptime: (id: string) => api.post<UptimeMonitor>(`/api/v1/monitors/uptime/${id}/pause`, {}),
  resumeUptime: (id: string) => api.post<UptimeMonitor>(`/api/v1/monitors/uptime/${id}/resume`, {}),
  deleteUptime: (id: string) => api.delete<void>(`/api/v1/monitors/uptime/${id}/`),
}
