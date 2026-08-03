// Static/pure pieces of the uptime monitor form — shared by the create and
// edit views via UptimeMonitorForm.vue. Kept out of the component (same split
// as lib/notificationChannelTypes.ts vs useNotificationChannelForm.ts) so the
// views can validate and build a payload without rendering anything.
import type {
  AssertionComparator,
  HttpMethod,
  JsonAssertion,
  KeywordMode,
  UptimeMonitor,
} from '@/api/monitors'

export interface UptimeMonitorFormState {
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
  httpMethod: HttpMethod
  acceptedStatusCodes: number[]
  maxResponseTimeMs: number
  channelIds: string[]
}

export const DEFAULT_MAX_RESPONSE_TIME_MS = 10000

export function createUptimeMonitorFormState(): UptimeMonitorFormState {
  return {
    name: '',
    url: '',
    intervalMins: 10,
    alertsEnabled: true,
    maxAlertsPerIncident: 3,
    alertAfterNFailures: 0,
    keyword: '',
    keywordMode: 'contains',
    keywordCaseSensitive: false,
    jsonAssertions: [],
    httpMethod: 'GET',
    acceptedStatusCodes: [200],
    maxResponseTimeMs: DEFAULT_MAX_RESPONSE_TIME_MS,
    channelIds: [],
  }
}

// Fields arriving possibly-null from the API get a form-friendly default here,
// so neither view carries the ?? branches.
export function uptimeMonitorToFormState(m: UptimeMonitor): UptimeMonitorFormState {
  return {
    name: m.name,
    url: m.url,
    intervalMins: m.intervalMins,
    alertsEnabled: m.alertsEnabled,
    maxAlertsPerIncident: m.maxAlertsPerIncident,
    alertAfterNFailures: m.alertAfterNFailures,
    keyword: m.keyword ?? '',
    keywordMode: m.keywordMode,
    keywordCaseSensitive: m.keywordCaseSensitive,
    jsonAssertions: [...(m.jsonAssertions ?? [])],
    httpMethod: m.httpMethod,
    acceptedStatusCodes: [...m.acceptedStatusCodes],
    maxResponseTimeMs: m.maxResponseTimeMs,
    channelIds: m.channelIds ?? [],
  }
}

// Trimmed payload shared by createUptime and updateUptime — both take the same
// field set, so building it twice was one of the ways the two views drifted.
export function uptimeMonitorFormPayload(f: UptimeMonitorFormState) {
  return {
    name: f.name.trim(),
    url: f.url.trim(),
    intervalMins: f.intervalMins,
    maxAlertsPerIncident: f.maxAlertsPerIncident,
    alertAfterNFailures: f.alertAfterNFailures,
    keyword: f.keyword.trim(),
    keywordMode: f.keywordMode,
    keywordCaseSensitive: f.keywordCaseSensitive,
    jsonAssertions: f.jsonAssertions,
    maxResponseTimeMs: f.maxResponseTimeMs,
    httpMethod: f.httpMethod,
    acceptedStatusCodes: f.acceptedStatusCodes,
    channelIds: f.channelIds,
  }
}

export function validateUptimeMonitorForm(f: UptimeMonitorFormState): string {
  if (!f.name.trim()) return 'Name is required'
  if (!f.url.trim()) return 'URL is required'
  if (!f.url.match(/^https?:\/\//)) return 'URL must start with http:// or https://'
  if (f.keyword.trim().length > 500) return 'Keyword must be 500 characters or fewer'
  if (f.acceptedStatusCodes.length === 0) return 'Select at least one accepted status code'
  if (f.maxResponseTimeMs < 1000 || f.maxResponseTimeMs > 30000)
    return 'Request timeout must be between 1000 and 30000 ms'
  return ''
}

export const keywordModeOptions: { label: string; value: KeywordMode }[] = [
  { label: 'Contains', value: 'contains' },
  { label: 'Does not contain', value: 'not_contains' },
]

export const httpMethodOptions: { label: string; value: HttpMethod }[] = [
  { label: 'GET', value: 'GET' },
  { label: 'HEAD', value: 'HEAD' },
  { label: 'POST', value: 'POST' },
]

export const statusCodeOptions = [200, 201, 202, 203, 204, 205, 206]

export const comparatorOptions: { label: string; value: AssertionComparator }[] = [
  { label: 'equals', value: 'equals' },
  { label: 'not equals', value: 'not_equals' },
  { label: 'contains', value: 'contains' },
  { label: '>', value: 'greater_than' },
  { label: '<', value: 'less_than' },
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
