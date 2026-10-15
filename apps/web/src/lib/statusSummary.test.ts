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

  it('shows a page with no data yet as waiting', () => {
    const summary: StatusSummary = { title: 't', overall: 'x', tone: 'none', monitors: 0, days: [] }
    expect(liveStrip(summary).tone).toBe('waiting')
  })
})
