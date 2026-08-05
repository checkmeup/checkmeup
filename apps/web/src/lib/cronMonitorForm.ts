// Static/pure pieces of the cron monitor form — shared by the create and edit
// views via CronMonitorForm.vue. Same split as lib/uptimeMonitorForm.ts.
import type { CronMonitor } from '@/api/monitors'

export interface CronMonitorFormState {
  name: string
  schedule: string
  gracePeriodMins: number
  alertsEnabled: boolean
  maxAlertsPerIncident: number
  alertAfterNFailures: number
  maxDurationMins: number | null
  channelIds: string[]
}

export function createCronMonitorFormState(): CronMonitorFormState {
  return {
    name: '',
    schedule: '',
    gracePeriodMins: 5,
    alertsEnabled: true,
    maxAlertsPerIncident: 3,
    alertAfterNFailures: 0,
    maxDurationMins: null,
    channelIds: [],
  }
}

export function cronMonitorToFormState(m: CronMonitor): CronMonitorFormState {
  return {
    name: m.name,
    schedule: m.schedule,
    gracePeriodMins: m.gracePeriodMins,
    alertsEnabled: m.alertsEnabled,
    maxAlertsPerIncident: m.maxAlertsPerIncident,
    alertAfterNFailures: m.alertAfterNFailures,
    maxDurationMins: m.maxDurationMins,
    channelIds: [...(m.channelIds ?? [])],
  }
}

export function cronMonitorFormPayload(f: CronMonitorFormState) {
  return {
    name: f.name.trim(),
    schedule: f.schedule.trim(),
    gracePeriodMins: f.gracePeriodMins,
    maxAlertsPerIncident: f.maxAlertsPerIncident,
    alertAfterNFailures: f.alertAfterNFailures,
    maxDurationMins: f.maxDurationMins,
    channelIds: f.channelIds,
  }
}

export function validateCronMonitorForm(f: CronMonitorFormState): string {
  if (!f.name.trim()) return 'Name is required'
  if (!f.schedule.trim()) return 'Schedule is required'
  return ''
}

export const scheduleExamples = [
  { label: 'Every hour', value: 'every 1h' },
  { label: 'Every 30 min', value: 'every 30m' },
  { label: 'Daily at midnight', value: '0 0 * * *' },
  { label: 'Daily at 9am', value: '0 9 * * *' },
  { label: 'Every weekday', value: '0 9 * * 1-5' },
]

export const graceOptions = [
  { label: '1 min', value: 1 },
  { label: '5 min', value: 5 },
  { label: '10 min', value: 10 },
  { label: '30 min', value: 30 },
  { label: '1 hour', value: 60 },
]

export const alertLimitOptions = [
  { label: '1 time', value: 1 },
  { label: '2 times', value: 2 },
  { label: '3 times (default)', value: 3 },
  { label: '5 times', value: 5 },
  { label: '10 times', value: 10 },
]

export const alertFilterOptions = [
  { label: 'Alert immediately (default)', value: 0 },
  { label: 'Skip first 1 failure', value: 1 },
  { label: 'Skip first 2 failures', value: 2 },
  { label: 'Skip first 3 failures', value: 3 },
  { label: 'Skip first 5 failures', value: 5 },
]

export const maxDurationOptions: { label: string; value: number | null }[] = [
  { label: 'Off (default)', value: null },
  { label: '5 min', value: 5 },
  { label: '15 min', value: 15 },
  { label: '30 min', value: 30 },
  { label: '1 hour', value: 60 },
  { label: '2 hours', value: 120 },
]
