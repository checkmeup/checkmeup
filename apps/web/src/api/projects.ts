import { api } from './client'
import type { MonitorType } from '@/lib/monitorRows'

// A named group of monitors, typically one per client (PRD projects).
export interface Project {
  id: string
  name: string
  monitorCount: number
  createdAt: string
}

export const projectsApi = {
  list: () => api.get<Project[]>('/api/v1/projects/'),
  create: (name: string) => api.post<Project>('/api/v1/projects/', { name }),
  rename: (id: string, name: string) => api.patch<Project>(`/api/v1/projects/${id}/`, { name }),
  delete: (id: string) => api.delete(`/api/v1/projects/${id}/`),
  // Makes exactly these monitors the project's: others in it leave it.
  setMonitors: (id: string, monitors: { type: MonitorType; id: string }[]) =>
    api.put<Project>(`/api/v1/projects/${id}/monitors`, { monitors }),
}
