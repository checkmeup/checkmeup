// Static/pure pieces of the port monitor form — shared by the create and edit
// views via PortMonitorForm.vue. Same split as lib/uptimeMonitorForm.ts.
import type { ExpectedState, PortMonitor } from '@/api/monitors'

export interface PortMonitorFormState {
  name: string
  host: string
  port: number | undefined
  expectedState: ExpectedState
  intervalMins: number
  alertsEnabled: boolean
  maxAlertsPerIncident: number
  alertAfterNFailures: number
  channelIds: string[]
}

export function createPortMonitorFormState(): PortMonitorFormState {
  return {
    name: '',
    host: '',
    port: undefined,
    expectedState: 'open',
    intervalMins: 10,
    alertsEnabled: true,
    maxAlertsPerIncident: 3,
    alertAfterNFailures: 0,
    channelIds: [],
  }
}

export function portMonitorToFormState(m: PortMonitor): PortMonitorFormState {
  return {
    name: m.name,
    host: m.host,
    port: m.port,
    expectedState: m.expectedState,
    intervalMins: m.intervalMins,
    alertsEnabled: m.alertsEnabled,
    maxAlertsPerIncident: m.maxAlertsPerIncident,
    alertAfterNFailures: m.alertAfterNFailures,
    channelIds: [...(m.channelIds ?? [])],
  }
}

export function portMonitorFormPayload(f: PortMonitorFormState) {
  return {
    name: f.name.trim(),
    host: f.host.trim(),
    port: f.port as number,
    expectedState: f.expectedState,
    intervalMins: f.intervalMins,
    maxAlertsPerIncident: f.maxAlertsPerIncident,
    alertAfterNFailures: f.alertAfterNFailures,
    channelIds: f.channelIds,
  }
}

export function validatePortMonitorForm(f: PortMonitorFormState): string {
  if (!f.name.trim()) return 'Name is required'
  if (!f.host.trim()) return 'Host is required'
  if (!f.port || f.port < 1 || f.port > 65535) return 'Port must be between 1 and 65535'
  return ''
}

export const expectedStateOptions: { label: string; value: ExpectedState }[] = [
  { label: 'Open — alert if it stops accepting connections', value: 'open' },
  { label: 'Closed — alert if it becomes reachable', value: 'closed' },
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

export function intervalOptionsFor(minIntervalMins: number) {
  return [
    ...(minIntervalMins === 1 ? [{ label: '1 minute', value: 1 }] : []),
    { label: '5 minutes', value: 5 },
    { label: '10 minutes', value: 10 },
    { label: '30 minutes', value: 30 },
  ]
}
