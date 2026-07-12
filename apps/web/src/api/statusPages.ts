import { api } from './client'

export interface StatusPage {
  id: string
  slug: string
  title: string
  description: string
  logoUrl: string
  hideBranding: boolean
  publicUrl: string
  createdAt: string
}

export interface StatusPageMonitorItem {
  id: string
  monitorType: 'cron' | 'uptime' | 'ssl' | 'domain' | 'port'
  monitorId: string
  displayName: string
  displayOrder: number
}

export interface StatusPageDetail extends StatusPage {
  monitors: StatusPageMonitorItem[]
}

export interface CreateStatusPageInput {
  slug: string
  title: string
  description: string
  logoUrl: string
}

export interface UpdateStatusPageInput {
  title: string
  description: string
  logoUrl: string
  hideBranding: boolean
}

export interface SetMonitorsInput {
  monitors: {
    monitorType: 'cron' | 'uptime' | 'ssl' | 'domain' | 'port'
    monitorId: string
    displayName: string
    displayOrder: number
  }[]
}

export interface SlugCheckResult {
  available: boolean
  reason: string
}

export const statusPagesApi = {
  checkSlug: (slug: string) =>
    api.get<SlugCheckResult>(`/api/v1/status-pages/check-slug?slug=${encodeURIComponent(slug)}`),
  list: () => api.get<StatusPage[]>('/api/v1/status-pages/'),
  create: (input: CreateStatusPageInput) => api.post<StatusPage>('/api/v1/status-pages/', input),
  get: (id: string) => api.get<StatusPageDetail>(`/api/v1/status-pages/${id}/`),
  update: (id: string, input: UpdateStatusPageInput) =>
    api.patch<StatusPage>(`/api/v1/status-pages/${id}/`, input),
  delete: (id: string) => api.delete(`/api/v1/status-pages/${id}/`),
  setMonitors: (id: string, input: SetMonitorsInput) =>
    api.put<StatusPageMonitorItem[]>(`/api/v1/status-pages/${id}/monitors`, input),
}
