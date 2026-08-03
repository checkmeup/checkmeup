// Static/pure pieces of the DNS monitor form — shared by the create and edit
// views via DNSMonitorForm.vue. Same split as lib/uptimeMonitorForm.ts.
import type { DNSMonitor, DNSRecordType } from '@/api/monitors'

export interface DNSMonitorFormState {
  name: string
  hostname: string
  recordType: DNSRecordType
  expectedValue: string
  intervalMins: number
  alertsEnabled: boolean
  maxAlertsPerIncident: number
  alertAfterNFailures: number
  channelIds: string[]
}

export function createDNSMonitorFormState(): DNSMonitorFormState {
  return {
    name: '',
    hostname: '',
    recordType: 'A',
    expectedValue: '',
    intervalMins: 10,
    alertsEnabled: true,
    maxAlertsPerIncident: 3,
    alertAfterNFailures: 0,
    channelIds: [],
  }
}

export function dnsMonitorToFormState(m: DNSMonitor): DNSMonitorFormState {
  return {
    name: m.name,
    hostname: m.hostname,
    recordType: m.recordType,
    expectedValue: m.expectedValue ?? '',
    intervalMins: m.intervalMins,
    alertsEnabled: m.alertsEnabled,
    maxAlertsPerIncident: m.maxAlertsPerIncident,
    alertAfterNFailures: m.alertAfterNFailures,
    channelIds: m.channelIds ?? [],
  }
}

// expectedValue is deliberately NOT here: the two screens mean different
// things by a blank field. Create omits it entirely (`undefined`) so the
// worker captures a baseline on first check; edit sends `''` to actively
// re-arm that baseline. Each view supplies its own.
export function dnsMonitorFormPayload(f: DNSMonitorFormState) {
  return {
    name: f.name.trim(),
    hostname: f.hostname.trim(),
    recordType: f.recordType,
    intervalMins: f.intervalMins,
    maxAlertsPerIncident: f.maxAlertsPerIncident,
    alertAfterNFailures: f.alertAfterNFailures,
    channelIds: f.channelIds,
  }
}

export function validateDNSMonitorForm(f: DNSMonitorFormState): string {
  if (!f.name.trim()) return 'Name is required'
  if (!f.hostname.trim()) return 'Hostname is required'
  return ''
}

export const recordTypeOptions: { label: string; value: DNSRecordType }[] = [
  { label: 'A — IPv4 address', value: 'A' },
  { label: 'AAAA — IPv6 address', value: 'AAAA' },
  { label: 'CNAME — canonical name', value: 'CNAME' },
  { label: 'MX — mail exchange', value: 'MX' },
  { label: 'TXT — text record', value: 'TXT' },
  { label: 'NS — nameserver', value: 'NS' },
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
