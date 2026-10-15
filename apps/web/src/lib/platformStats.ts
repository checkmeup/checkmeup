// Checkmeup's platform-wide totals (GET /api/v1/stats).
export interface PlatformStats {
  monitors: number
  pings30d: number
  checks30d: number
  incidents30d: number
}

// Below these a number undersells the service more than it proves it, so
// it's left out rather than shown small.
const thresholds = {
  checks30d: 10_000,
  pings30d: 10_000,
  incidents30d: 100,
  monitors: 100,
}

// The line under the hero's live strip: only the totals big enough to show,
// or null when none are and the line should stay empty.
export function statsLine(stats: PlatformStats): string | null {
  const fmt = (n: number) => n.toLocaleString('en-US')
  const parts: string[] = []
  if (stats.checks30d >= thresholds.checks30d) parts.push(`${fmt(stats.checks30d)} checks`)
  if (stats.pings30d >= thresholds.pings30d) parts.push(`${fmt(stats.pings30d)} cron pings`)
  if (stats.incidents30d >= thresholds.incidents30d)
    parts.push(`${fmt(stats.incidents30d)} outages caught`)
  if (parts.length === 0) return null
  const line = `Across Checkmeup in the last 30 days: ${parts.join(' · ')}`
  return stats.monitors >= thresholds.monitors
    ? `${line} · ${fmt(stats.monitors)} monitors running now`
    : line
}
