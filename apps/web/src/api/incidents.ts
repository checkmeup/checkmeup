import { api } from './client'

export type IncidentSeverity = 'minor' | 'major' | 'critical'
export type IncidentStatus = 'investigating' | 'identified' | 'monitoring' | 'resolved'

export interface IncidentMonitorRef {
  monitorType: 'cron' | 'uptime' | 'ssl' | 'domain' | 'port'
  monitorId: string
  name: string
}

export interface IncidentUpdate {
  id: string
  message: string
  status: IncidentStatus
  createdAt: string
}

export interface Incident {
  id: string
  title: string
  severity: IncidentSeverity
  status: IncidentStatus
  monitors?: IncidentMonitorRef[]
  monitorCount: number
  updates?: IncidentUpdate[]
  createdAt: string
  updatedAt: string
  resolvedAt: string | null
}

export interface CreateIncidentInput {
  title: string
  message: string
  severity: IncidentSeverity
  monitors: { monitorType: 'cron' | 'uptime' | 'ssl' | 'domain' | 'port'; monitorId: string }[]
  confirmOverlap?: boolean
}

export const incidentsApi = {
  list: () => api.get<Incident[]>('/api/v1/incidents/'),
  get: (id: string) => api.get<Incident>(`/api/v1/incidents/${id}/`),
  create: (input: CreateIncidentInput) => api.post<Incident>('/api/v1/incidents/', input),
  updateTitle: (id: string, title: string) =>
    api.patch<Incident>(`/api/v1/incidents/${id}/`, { title }),
  delete: (id: string) => api.delete(`/api/v1/incidents/${id}/`),
  postUpdate: (id: string, message: string, status: IncidentStatus) =>
    api.post<Incident>(`/api/v1/incidents/${id}/updates`, { message, status }),
  updateUpdateMessage: (id: string, updateId: string, message: string) =>
    api.patch<Incident>(`/api/v1/incidents/${id}/updates/${updateId}`, { message }),
}
