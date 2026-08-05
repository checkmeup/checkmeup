import { describe, expect, it } from 'vitest'
import type { UptimeMonitor } from '@/api/monitors'
import {
  uptimeMonitorToFormState,
  validateUptimeMonitorForm,
  createUptimeMonitorFormState,
} from './uptimeMonitorForm'

function monitorWith(overrides: Partial<UptimeMonitor> = {}): UptimeMonitor {
  return {
    id: 'u1',
    name: 'API uptime',
    url: 'https://api.example.com/health',
    intervalMins: 5,
    status: 'up',
    alertsEnabled: true,
    maxAlertsPerIncident: 3,
    alertAfterNFailures: 0,
    lastCheckedAt: null,
    createdAt: '2026-01-01T00:00:00Z',
    uptime24h: 99.9,
    keyword: null,
    keywordMode: 'contains',
    keywordCaseSensitive: false,
    jsonAssertions: [],
    maxResponseTimeMs: 10000,
    httpMethod: 'GET',
    acceptedStatusCodes: [200],
    channelIds: [],
    ...overrides,
  } as UptimeMonitor
}

describe('uptimeMonitorToFormState', () => {
  it('does not share JsonAssertion objects with the source monitor', () => {
    // The source monitor here stands in for the object TanStack Query holds in
    // its cache: the edit view populates its form straight from `detail.monitor`.
    const source = monitorWith({
      jsonAssertions: [{ path: '$.status', comparator: 'equals', expected: 'ok' }],
    })

    const form = uptimeMonitorToFormState(source)
    expect(form.jsonAssertions).toEqual([
      { path: '$.status', comparator: 'equals', expected: 'ok' },
    ])

    form.jsonAssertions[0].path = '$.edited'

    expect(source.jsonAssertions![0].path).toBe('$.status')
  })

  it('does not share the acceptedStatusCodes array with the source monitor', () => {
    const source = monitorWith({ acceptedStatusCodes: [200] })

    const form = uptimeMonitorToFormState(source)
    expect(form.acceptedStatusCodes).toEqual([200])

    form.acceptedStatusCodes.push(201)

    expect(source.acceptedStatusCodes).toEqual([200])
  })

  it('does not share the channelIds array with the source monitor', () => {
    const source = monitorWith({ channelIds: ['c1'] })

    const form = uptimeMonitorToFormState(source)
    expect(form.channelIds).toEqual(['c1'])

    form.channelIds.push('c2')

    expect(source.channelIds).toEqual(['c1'])
  })
})

describe('validateUptimeMonitorForm', () => {
  const valid = () => ({
    ...createUptimeMonitorFormState(),
    name: 'API',
    url: 'https://example.com',
  })

  it('accepts a keyword of exactly 500 characters', () => {
    expect(validateUptimeMonitorForm({ ...valid(), keyword: 'a'.repeat(500) })).toBe('')
  })

  it('rejects a keyword of 501 characters', () => {
    expect(validateUptimeMonitorForm({ ...valid(), keyword: 'a'.repeat(501) })).toBe(
      'Keyword must be 500 characters or fewer',
    )
  })

  it('rejects an empty accepted-status-code set', () => {
    expect(validateUptimeMonitorForm({ ...valid(), acceptedStatusCodes: [] })).toBe(
      'Select at least one accepted status code',
    )
  })

  it('accepts the timeout boundaries and rejects just outside them', () => {
    expect(validateUptimeMonitorForm({ ...valid(), maxResponseTimeMs: 1000 })).toBe('')
    expect(validateUptimeMonitorForm({ ...valid(), maxResponseTimeMs: 30000 })).toBe('')
    const msg = 'Request timeout must be between 1000 and 30000 ms'
    expect(validateUptimeMonitorForm({ ...valid(), maxResponseTimeMs: 999 })).toBe(msg)
    expect(validateUptimeMonitorForm({ ...valid(), maxResponseTimeMs: 30001 })).toBe(msg)
  })
})
