// Static/pure pieces of the maintenance window form — shared by the create
// and edit views via MaintenanceWindowForm.vue. Same split as
// lib/uptimeMonitorForm.ts.
import type { MaintenanceWindow } from '@/api/maintenance'

export type MaintenanceMonitorRef = {
  monitorType: 'cron' | 'uptime' | 'ssl' | 'domain' | 'port'
  monitorId: string
  name: string
}

export interface MaintenanceWindowFormState {
  title: string
  message: string
  startsAt: string
  noEnd: boolean
  endsAt: string
  monitors: MaintenanceMonitorRef[]
}

export function createMaintenanceWindowFormState(): MaintenanceWindowFormState {
  return { title: '', message: '', startsAt: '', noEnd: false, endsAt: '', monitors: [] }
}

// The datetime-local input wants local wall-clock time, not the ISO string the
// API returns.
export function toLocalInputValue(iso: string): string {
  const d = new Date(iso)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

export function maintenanceWindowToFormState(w: MaintenanceWindow): MaintenanceWindowFormState {
  return {
    title: w.title,
    message: w.message,
    startsAt: toLocalInputValue(w.startsAt),
    noEnd: !w.endsAt,
    endsAt: w.endsAt ? toLocalInputValue(w.endsAt) : '',
    monitors: (w.monitors ?? []).map((m) => ({
      monitorType: m.monitorType,
      monitorId: m.monitorId,
      name: m.name,
    })),
  }
}

export function maintenanceWindowFormPayload(f: MaintenanceWindowFormState) {
  return {
    title: f.title.trim(),
    message: f.message.trim(),
    startsAt: new Date(f.startsAt).toISOString(),
    endsAt: f.noEnd ? null : new Date(f.endsAt).toISOString(),
    monitors: f.monitors.map((m) => ({ monitorType: m.monitorType, monitorId: m.monitorId })),
  }
}

export function validateMaintenanceWindowForm(f: MaintenanceWindowFormState): string {
  if (!f.title.trim()) return 'Title is required'
  if (!f.startsAt) return 'Start time is required'
  if (!f.noEnd && !f.endsAt) return 'End time is required, or check "no end date"'
  if (f.monitors.length === 0) return 'Select at least one monitor'
  return ''
}
