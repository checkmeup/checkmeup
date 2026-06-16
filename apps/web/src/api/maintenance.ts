import { api } from './client'

export interface MaintenanceMonitorRef {
  monitorType: 'cron' | 'uptime' | 'ssl'
  monitorId: string
  name: string
}

export interface MaintenanceWindow {
  id: string
  title: string
  message: string
  startsAt: string
  endsAt: string | null
  status: 'upcoming' | 'active' | 'ended'
  monitors?: MaintenanceMonitorRef[]
  monitorCount: number
  createdAt: string
}

export interface SaveMaintenanceWindowInput {
  title: string
  message: string
  startsAt: string
  endsAt: string | null
  monitors: { monitorType: 'cron' | 'uptime' | 'ssl'; monitorId: string }[]
}

export const maintenanceApi = {
  list: () => api.get<MaintenanceWindow[]>('/api/v1/maintenance-windows/'),
  get: (id: string) => api.get<MaintenanceWindow>(`/api/v1/maintenance-windows/${id}/`),
  create: (input: SaveMaintenanceWindowInput) =>
    api.post<MaintenanceWindow>('/api/v1/maintenance-windows/', input),
  update: (id: string, input: SaveMaintenanceWindowInput) =>
    api.patch<MaintenanceWindow>(`/api/v1/maintenance-windows/${id}/`, input),
  delete: (id: string) => api.delete<void>(`/api/v1/maintenance-windows/${id}/`),
  endNow: (id: string) => api.post<MaintenanceWindow>(`/api/v1/maintenance-windows/${id}/end`, {}),
}
