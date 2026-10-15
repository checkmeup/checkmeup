import { describe, expect, it } from 'vitest'
import { liveStrip, stripTotals, type StatusSummary } from './statusSummary'

describe('liveStrip', () => {
  it('maps each day to a tick and the headline to a status label', () => {
    const summary: StatusSummary = {
      title: 'checkmeup.net',
      overall: 'Partial outage',
      tone: 'warn',
      monitors: 3,
      days: [
        { date: 'Jul 1', tone: 'up' },
        { date: 'Jul 2', tone: 'down' },
        { date: 'Jul 3', tone: 'warn' },
        { date: 'Jul 4', tone: 'none' },
      ],
    }

    expect(liveStrip(summary)).toEqual({
      tone: 'warn',
      label: 'Partial outage',
      checks: [
        { ok: true, label: 'Jul 1 · Operational' },
        { ok: false, label: 'Jul 2 · Outage' },
        { ok: true, slow: true, label: 'Jul 3 · Degraded' },
        { ok: null, label: 'Jul 4 · No data' },
      ],
    })
  })

  it("adds a day's uptime and latency to its label when they were measured", () => {
    const summary: StatusSummary = {
      title: 't',
      overall: 'x',
      tone: 'up',
      monitors: 2,
      days: [
        { date: 'Oct 14', tone: 'down', uptime: 99.93, latencyMs: 182 },
        { date: 'Oct 15', tone: 'up', uptime: 100 },
      ],
    }

    expect(liveStrip(summary).checks.map((c) => c.label)).toEqual([
      'Oct 14 · Outage · 99.93% uptime · 182 ms',
      'Oct 15 · Operational · 100% uptime',
    ])
  })

  it('shows a page with no data yet as waiting', () => {
    const summary: StatusSummary = { title: 't', overall: 'x', tone: 'none', monitors: 0, days: [] }
    expect(liveStrip(summary).tone).toBe('waiting')
  })
})

describe('stripTotals', () => {
  const summary: StatusSummary = {
    title: 't',
    overall: 'x',
    tone: 'up',
    monitors: 2,
    days: [
      { date: 'Oct 12', tone: 'none' },
      { date: 'Oct 13', tone: 'up', uptime: 100, latencyMs: 200 },
      { date: 'Oct 14', tone: 'down', uptime: 99.97, latencyMs: 181 },
      { date: 'Oct 15', tone: 'up', uptime: 100, latencyMs: 190 },
    ],
  }

  it('averages the measured days in range', () => {
    expect(stripTotals(summary, 4)).toEqual({ uptime: 99.99 })
  })

  it('truncates uptime, so 99.996% shows as 99.99% rather than 100%', () => {
    const almost: StatusSummary = { ...summary, days: [{ date: 'Oct 15', tone: 'down', uptime: 99.996 }] }
    expect(stripTotals(almost, 90).uptime).toBe(99.99)
  })

  it('only looks at the last N days', () => {
    expect(stripTotals(summary, 1)).toEqual({ uptime: 100 })
  })

  it('leaves uptime out when no day measured it', () => {
    expect(stripTotals({ ...summary, days: [{ date: 'Oct 12', tone: 'up' }] }, 90)).toEqual({
      uptime: undefined,
    })
  })
})
