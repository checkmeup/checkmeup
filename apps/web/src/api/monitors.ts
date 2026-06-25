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
  alertAfterNFailures: number
  lastPingAt: string | null
  nextPingAt: string | null
  createdAt: string
  channelIds?: string[]
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
  alertAfterNFailures: number
  channelIds?: string[]
}

export interface UpdateCronMonitorInput {
  name: string
  schedule: string
  gracePeriodMins: number
  alertsEnabled: boolean
  maxAlertsPerIncident: number
  alertAfterNFailures: number
  channelIds: string[]
}

export type KeywordMode = 'contains' | 'not_contains'
export type AssertionComparator =
  | 'equals'
  | 'not_equals'
  | 'contains'
  | 'greater_than'
  | 'less_than'

export interface JsonAssertion {
  path: string
  comparator: AssertionComparator
  expected: string
}

export interface UptimeMonitor {
  id: string
  name: string
  url: string
  intervalMins: number
  status: 'waiting' | 'up' | 'down' | 'paused'
  alertsEnabled: boolean
  maxAlertsPerIncident: number
  alertAfterNFailures: number
  lastCheckedAt: string | null
  createdAt: string
  uptime24h: number | null
  keyword: string | null
  keywordMode: KeywordMode
  keywordCaseSensitive: boolean
  jsonAssertions: JsonAssertion[]
  maxResponseTimeMs: number | null
  channelIds?: string[]
}

export interface UptimeCheck {
  id: string
  checkedAt: string
  statusCode: number | null
  responseTimeMs: number
  isUp: boolean
  failureReason: string | null
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
  alertAfterNFailures: number
  keyword: string
  keywordMode: KeywordMode
  keywordCaseSensitive: boolean
  jsonAssertions: JsonAssertion[]
  maxResponseTimeMs: number | null
  channelIds?: string[]
}

export interface UpdateUptimeMonitorInput {
  name: string
  url: string
  intervalMins: number
  alertsEnabled: boolean
  maxAlertsPerIncident: number
  alertAfterNFailures: number
  keyword: string
  keywordMode: KeywordMode
  keywordCaseSensitive: boolean
  jsonAssertions: JsonAssertion[]
  maxResponseTimeMs: number | null
  channelIds: string[]
}

export interface SSLMonitor {
  id: string
  name: string
  hostname: string
  status: 'waiting' | 'up' | 'expiring_soon' | 'expired' | 'error' | 'paused'
  alertsEnabled: boolean
  alertAfterNFailures: number
  maxAlertsPerIncident: number
  expiresAt: string | null
  issuer: string | null
  errorMsg: string | null
  daysUntilExpiry: number | null
  lastCheckedAt: string | null
  createdAt: string
  channelIds?: string[]
}

export interface CreateSSLMonitorInput {
  name: string
  hostname: string
  channelIds?: string[]
}

export interface UpdateSSLMonitorInput {
  name: string
  hostname: string // passed through but not shown in UI (domain changes require delete + recreate)
  alertsEnabled: boolean
  alertAfterNFailures: number
  maxAlertsPerIncident: number
  channelIds: string[]
}

export interface DomainMonitor {
  id: string
  name: string
  domain: string
  status: 'waiting' | 'up' | 'expiring_soon' | 'expired' | 'error' | 'paused'
  alertsEnabled: boolean
  alertAfterNFailures: number
  maxAlertsPerIncident: number
  expiresAt: string | null
  registrar: string | null
  errorMsg: string | null
  daysUntilExpiry: number | null
  lastCheckedAt: string | null
  createdAt: string
  channelIds?: string[]
}

export interface CreateDomainMonitorInput {
  name: string
  domain: string
  channelIds?: string[]
}

export interface UpdateDomainMonitorInput {
  name: string
  domain: string // passed through but not shown in UI (domain changes require delete + recreate)
  alertsEnabled: boolean
  alertAfterNFailures: number
  maxAlertsPerIncident: number
  channelIds: string[]
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

  listDomain: () => api.get<DomainMonitor[]>('/api/v1/monitors/domain/'),
  getDomain: (id: string) => api.get<DomainMonitor>(`/api/v1/monitors/domain/${id}/`),
  createDomain: (input: CreateDomainMonitorInput) =>
    api.post<DomainMonitor>('/api/v1/monitors/domain/', input),
  updateDomain: (id: string, input: UpdateDomainMonitorInput) =>
    api.patch<DomainMonitor>(`/api/v1/monitors/domain/${id}/`, input),
  pauseDomain: (id: string) => api.post<DomainMonitor>(`/api/v1/monitors/domain/${id}/pause`, {}),
  resumeDomain: (id: string) => api.post<DomainMonitor>(`/api/v1/monitors/domain/${id}/resume`, {}),
  deleteDomain: (id: string) => api.delete(`/api/v1/monitors/domain/${id}/`),

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
