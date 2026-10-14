import type { Tone } from '@/lib/monitorRows'

// The public summary of a status page (GET /status/:slug/summary.json): its
// headline and one combined 90-day strip, oldest day first.
export interface StatusSummary {
  title: string
  overall: string
  tone: 'up' | 'down' | 'warn' | 'paused' | 'none'
  monitors: number
  days: { date: string; tone: StatusSummary['tone'] }[]
}

// Turns a summary into what StatusLabel and CheckStrip draw: a degraded day
// is a slow tick, a day without data an empty one.
export function liveStrip(summary: StatusSummary): {
  tone: Tone
  label: string
  checks: { ok: boolean | null; slow?: boolean }[]
} {
  return {
    tone: summary.tone === 'none' ? 'waiting' : summary.tone,
    label: summary.overall,
    checks: summary.days.map((d) => {
      if (d.tone === 'down') return { ok: false }
      if (d.tone === 'warn') return { ok: true, slow: true }
      if (d.tone === 'up') return { ok: true }
      return { ok: null }
    }),
  }
}
