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
}
