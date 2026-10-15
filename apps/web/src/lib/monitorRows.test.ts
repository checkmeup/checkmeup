import { describe, expect, it } from 'vitest'
import type {
  CronMonitor,
  DNSMonitor,
  PortMonitor,
  SSLMonitor,
  UptimeMonitor,
} from '@/api/monitors'
import {
  cronRow,
  attentionSentence,
  dnsRow,
  fmtDaysLeft,
  groupByProject,
  matchesSearch,
  portRow,
  relativeTime,
  sortRows,
  sslRow,
  statusLabel,
  toneOf,
  uptimeRow,
} from './monitorRows'

const now = Date.parse('2026-10-14T12:00:00Z')

function uptime(overrides: Partial<UptimeMonitor> = {}): UptimeMonitor {
  return {
    id: 'u1',
    name: 'shop',
    url: 'https://shop.example.com',
    status: 'up',
    uptime24h: 99.5,
    lastCheckedAt: '2026-10-14T11:58:00Z',
    keyword: null,
    keywordMode: 'contains',
    ...overrides,
  } as UptimeMonitor
}

function ssl(overrides: Partial<SSLMonitor> = {}): SSLMonitor {
  return {
    id: 's1',
    name: 'cert',
    hostname: 'example.com',
    status: 'up',
    expiresAt: '2026-10-26T00:00:00Z',
    daysUntilExpiry: 12,
    issuer: 'R11',
    errorMsg: null,
    lastCheckedAt: null,
    ...overrides,
  } as SSLMonitor
}

describe('toneOf', () => {
  it('groups failed, expired and errored checks as down', () => {
    expect(toneOf('down')).toBe('down')
    expect(toneOf('expired')).toBe('down')
    expect(toneOf('error')).toBe('down')
  })

  it('treats expiring soon as a warning and unknown states as waiting', () => {
    expect(toneOf('expiring_soon')).toBe('warn')
    expect(toneOf('waiting')).toBe('waiting')
    expect(toneOf('something-new')).toBe('waiting')
  })
})

describe('statusLabel', () => {
  it('calls a healthy certificate or domain "Valid", anything else "Up"', () => {
    expect(statusLabel('ssl', 'up')).toBe('Valid')
    expect(statusLabel('domain', 'up')).toBe('Valid')
    expect(statusLabel('uptime', 'up')).toBe('Up')
    expect(statusLabel('ssl', 'expiring_soon')).toBe('Expiring soon')
  })
})

describe('relativeTime', () => {
  it('formats past and future times and a missing one', () => {
    expect(relativeTime(null, now)).toBe('—')
    expect(relativeTime('2026-10-14T11:59:40Z', now)).toBe('just now')
    expect(relativeTime('2026-10-14T11:45:00Z', now)).toBe('15m ago')
    expect(relativeTime('2026-10-14T09:00:00Z', now)).toBe('3h ago')
    expect(relativeTime('2026-10-12T12:00:00Z', now)).toBe('2d ago')
    expect(relativeTime('2026-10-14T12:30:00Z', now)).toBe('in 30m')
  })
})

describe('fmtDaysLeft', () => {
  it('pluralizes days and calls a negative count expired', () => {
    expect(fmtDaysLeft(null)).toBe('—')
    expect(fmtDaysLeft(1)).toBe('1 day')
    expect(fmtDaysLeft(12)).toBe('12 days')
    expect(fmtDaysLeft(-1)).toBe('Expired')
  })
})

describe('row builders', () => {
  it("shows an uptime monitor's keyword rule as its note", () => {
    expect(uptimeRow(uptime({ keyword: 'Welcome' }), now).note).toBe('contains "Welcome"')
    expect(uptimeRow(uptime({ keyword: 'Error', keywordMode: 'not_contains' }), now).note).toBe(
      'does not contain "Error"',
    )
    expect(uptimeRow(uptime(), now).note).toBeNull()
  })

  it("shows an SSL monitor's days left with its expiry date and issuer", () => {
    const row = sslRow(ssl(), now)
    expect(row.metric).toBe('12 days')
    expect(row.metricSub).toContain('R11')
  })

  it('shows the check error as the note only while the monitor is in error', () => {
    expect(sslRow(ssl({ status: 'error', errorMsg: 'handshake failed' }), now).note).toBe(
      'handshake failed',
    )
    expect(sslRow(ssl({ errorMsg: 'stale' }), now).note).toBeNull()
  })

  it("says when a cron monitor's next ping is due, or how long ago it was due", () => {
    const cron = (nextPingAt: string | null) =>
      ({
        id: 'c1',
        name: 'backup',
        schedule: '0 3 * * *',
        status: 'up',
        lastPingAt: null,
        nextPingAt,
      }) as CronMonitor
    expect(cronRow(cron('2026-10-14T16:00:00Z'), now).metricSub).toBe('next ping in 4h')
    expect(cronRow(cron('2026-10-14T11:53:00Z'), now).metricSub).toBe('was due 7m ago')
    expect(cronRow(cron(null), now).metricSub).toBe('last ping')
  })

  it('notes a port monitor that expects the port closed', () => {
    const port = {
      id: 'p1',
      name: 'db',
      host: 'db.local',
      port: 5432,
      status: 'up',
      expectedState: 'closed',
      uptime24h: 100,
      lastCheckedAt: null,
    } as PortMonitor
    expect(portRow(port, now).target).toBe('db.local:5432')
    expect(portRow(port, now).note).toBe('expects the port closed')
  })

  it("shows a DNS monitor's last resolved value", () => {
    const dns = {
      id: 'd1',
      name: 'apex',
      hostname: 'example.com',
      recordType: 'A',
      lastResolvedValue: '93.184.216.34',
      status: 'up',
      uptime24h: 100,
      lastCheckedAt: null,
    } as DNSMonitor
    expect(dnsRow(dns, now).metric).toBe('93.184.216.34')
    expect(dnsRow(dns, now).metricSub).toBe('A record')
  })
})

describe('sortRows', () => {
  it('puts down first, then warnings, then the rest by name', () => {
    const rows = sortRows([
      uptimeRow(uptime({ id: '1', name: 'b-up' }), now),
      uptimeRow(uptime({ id: '2', name: 'a-paused', status: 'paused' }), now),
      sslRow(ssl({ id: '3', name: 'c-expiring', status: 'expiring_soon' }), now),
      uptimeRow(uptime({ id: '4', name: 'd-down', status: 'down' }), now),
      uptimeRow(uptime({ id: '5', name: 'a-up' }), now),
    ])
    expect(rows.map((r) => r.name)).toEqual(['d-down', 'c-expiring', 'a-up', 'b-up', 'a-paused'])
  })
})

describe('matchesSearch', () => {
  it('matches name or target, case-insensitively, and everything when empty', () => {
    const row = uptimeRow(uptime(), now)
    expect(matchesSearch(row, '')).toBe(true)
    expect(matchesSearch(row, 'SHOP')).toBe(true)
    expect(matchesSearch(row, 'example.com')).toBe(true)
    expect(matchesSearch(row, 'billing')).toBe(false)
  })
})

describe('attentionSentence', () => {
  it('explains each problem in words', () => {
    expect(attentionSentence(uptimeRow(uptime({ status: 'down' }), now))).toBe(
      'Down — the last check of https://shop.example.com failed.',
    )
    expect(attentionSentence(sslRow(ssl({ status: 'expiring_soon' }), now))).toBe(
      'Certificate expires in 12 days.',
    )
    expect(attentionSentence(sslRow(ssl({ status: 'expired', daysUntilExpiry: -2 }), now))).toBe(
      'The certificate for example.com has expired.',
    )
    expect(
      attentionSentence(sslRow(ssl({ status: 'error', errorMsg: 'no route to host' }), now)),
    ).toBe('Check failed: no route to host')
  })
})

describe('groupByProject (PRD projects)', () => {
  const projects = [
    { id: 'p2', name: 'Globex' },
    { id: 'p1', name: 'Acme' },
    { id: 'p3', name: 'Empty' },
  ]

  it('carries each monitor\'s project on its row, null without one', () => {
    expect(uptimeRow(uptime({ projectId: 'p1' }), now).projectId).toBe('p1')
    expect(uptimeRow(uptime(), now).projectId).toBeNull()
  })

  it('groups rows by project name, "No project" last, leaving out empty projects', () => {
    const rows = [
      uptimeRow(uptime({ id: '1', name: 'a', projectId: 'p2' }), now),
      uptimeRow(uptime({ id: '2', name: 'b', projectId: 'p1' }), now),
      uptimeRow(uptime({ id: '3', name: 'c' }), now),
      uptimeRow(uptime({ id: '4', name: 'd', projectId: 'p1' }), now),
    ]
    const groups = groupByProject(rows, projects)
    expect(groups.map((g) => [g.name, g.rows.map((r) => r.name)])).toEqual([
      ['Acme', ['b', 'd']],
      ['Globex', ['a']],
      ['No project', ['c']],
    ])
    expect(groups.map((g) => g.key)).toEqual(['p1', 'p2', 'none'])
  })

  it('counts the monitors that need attention in each group', () => {
    const rows = [
      uptimeRow(uptime({ id: '1', status: 'down', projectId: 'p1' }), now),
      sslRow(ssl({ id: '2', status: 'expiring_soon', projectId: 'p1' }), now),
      uptimeRow(uptime({ id: '3', status: 'paused', projectId: 'p1' }), now),
      uptimeRow(uptime({ id: '4' }), now),
    ]
    const [acme, none] = groupByProject(rows, projects)
    expect(acme!.attention).toBe(2)
    expect(none!.attention).toBe(0)
  })

  it('puts a row whose project is gone under "No project"', () => {
    const groups = groupByProject([uptimeRow(uptime({ projectId: 'deleted' }), now)], projects)
    expect(groups.map((g) => g.name)).toEqual(['No project'])
  })
})
