import { describe, expect, it } from 'vitest'
import { liveStrip, type StatusSummary } from './statusSummary'

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
