import type { Tone } from '@/lib/monitorRows'

// The public summary of a status page (GET /status/:slug/summary.json): its
// headline and one combined 90-day strip, oldest day first.
export interface StatusSummary {
  title: string
  overall: string
  tone: 'up' | 'down' | 'warn' | 'paused' | 'none'
  monitors: number
  // uptime (% of checks passed) and latencyMs (average uptime response time)
  // are absent on a day none of the page's check-based monitors ran.
  days: { date: string; tone: StatusSummary['tone']; uptime?: number; latencyMs?: number }[]
}

// What a day's tooltip says it was, the same words as the status page.
const dayWords: Record<StatusSummary['tone'], string> = {
  up: 'Operational',
  warn: 'Degraded',
  down: 'Outage',
  paused: 'Paused',
  none: 'No data',
}

// A day's tooltip: its date and status word, then its uptime and latency
// when the page's checks measured them.
function dayLabel(d: StatusSummary['days'][number]): string {
  const parts = [d.date, dayWords[d.tone]]
  if (d.uptime !== undefined) parts.push(`${d.uptime}% uptime`)
  if (d.latencyMs !== undefined) parts.push(`${d.latencyMs} ms`)
  return parts.join(' · ')
}

// Turns a summary into what StatusLabel and CheckStrip draw: a degraded day
// is a slow tick, a day without data an empty one, and each tick's label
// describes its day.
export function liveStrip(summary: StatusSummary): {
  tone: Tone
  label: string
  checks: { ok: boolean | null; slow?: boolean; label: string }[]
} {
  return {
    tone: summary.tone === 'none' ? 'waiting' : summary.tone,
    label: summary.overall,
    checks: summary.days.map((d) => {
      const label = dayLabel(d)
      if (d.tone === 'down') return { ok: false, label }
      if (d.tone === 'warn') return { ok: true, slow: true, label }
      if (d.tone === 'up') return { ok: true, label }
      return { ok: null, label }
    }),
  }
}

// The strip's headline numbers over its last `lastDays` days: uptime as the
// mean of the days' uptime, truncated (not rounded) to two decimals so it
// never shows better than it was, and the mean of the days' average
// response times. Each is absent when no day in the range measured it.
export function stripTotals(
  summary: StatusSummary,
  lastDays: number,
): { uptime?: number; latencyMs?: number } {
  const days = summary.days.slice(-lastDays)
  const uptimes = days.flatMap((d) => (d.uptime === undefined ? [] : [d.uptime]))
  const latencies = days.flatMap((d) => (d.latencyMs === undefined ? [] : [d.latencyMs]))
  const mean = (xs: number[]) => xs.reduce((a, b) => a + b, 0) / xs.length
  return {
    // The epsilon keeps float noise (99.99 × 100 = 9998.999…) from
    // truncating a whole hundredth away.
    uptime: uptimes.length ? Math.floor(mean(uptimes) * 100 + 1e-9) / 100 : undefined,
    latencyMs: latencies.length ? Math.round(mean(latencies)) : undefined,
  }
}
