import { describe, expect, it } from 'vitest'
import { statsLine } from './platformStats'

describe('statsLine', () => {
  it('shows only the totals above their thresholds', () => {
    expect(statsLine({ monitors: 9, pings30d: 0, checks30d: 28396, incidents30d: 7 })).toBe(
      'Across Checkmeup in the last 30 days: 28,396 checks',
    )
  })

  it('joins every total once they are all big enough', () => {
    expect(
      statsLine({ monitors: 1200, pings30d: 54000, checks30d: 2_100_000, incidents30d: 340 }),
    ).toBe(
      'Across Checkmeup in the last 30 days: 2,100,000 checks · 54,000 cron pings · 340 outages caught · 1,200 monitors running now',
    )
  })

  it('is null when nothing is big enough, even with many monitors', () => {
    expect(statsLine({ monitors: 500, pings30d: 10, checks30d: 9999, incidents30d: 99 })).toBeNull()
  })
})
