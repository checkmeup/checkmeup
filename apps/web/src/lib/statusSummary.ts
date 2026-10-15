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

// What a day's tooltip says it was, the same words as the status page.
const dayWords: Record<StatusSummary['tone'], string> = {
  up: 'Operational',
  warn: 'Degraded',
  down: 'Outage',
  paused: 'Paused',
  none: 'No data',
}

// Turns a summary into what StatusLabel and CheckStrip draw: a degraded day
// is a slow tick, a day without data an empty one, and each tick's label
// names its date.
export function liveStrip(summary: StatusSummary): {
  tone: Tone
  label: string
  checks: { ok: boolean | null; slow?: boolean; label: string }[]
} {
  return {
    tone: summary.tone === 'none' ? 'waiting' : summary.tone,
    label: summary.overall,
    checks: summary.days.map((d) => {
      const label = `${d.date} · ${dayWords[d.tone]}`
      if (d.tone === 'down') return { ok: false, label }
      if (d.tone === 'warn') return { ok: true, slow: true, label }
      if (d.tone === 'up') return { ok: true, label }
      return { ok: null, label }
    }),
  }
}
