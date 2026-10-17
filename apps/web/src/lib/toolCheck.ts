import type { Tone } from './monitorRows'

// One run of a free checker (PRD free-tools): the API's result, or a message
// to show in its place.
export type ToolOutcome<T> = { ok: true; data: T } | { ok: false; message: string }

// Asks a /api/v1/tools endpoint about value. Plain fetch: the tools need no
// session, and a failure is shown on the page rather than thrown.
export async function runToolCheck<T>(
  endpoint: string,
  param: string,
  value: string,
): Promise<ToolOutcome<T>> {
  try {
    const res = await fetch(`${endpoint}?${param}=${encodeURIComponent(value.trim())}`)
    if (res.status === 429) {
      return { ok: false, message: 'Too many checks right now. Try again in a minute.' }
    }
    const body = await res.json()
    if (!res.ok) {
      return { ok: false, message: body?.error || "The check couldn't run. Try again." }
    }
    return { ok: true, data: body as T }
  } catch {
    return { ok: false, message: "The check couldn't run. Try again." }
  }
}

// How an expiry date reads: expiring within 30 days is a warning, matching
// when SSL and domain monitors turn "expiring soon".
export function expiryStatus(daysLeft: number): { tone: Tone; label: string } {
  if (daysLeft < 0) return { tone: 'down', label: 'Expired' }
  if (daysLeft === 0) return { tone: 'warn', label: 'Expires today' }
  if (daysLeft <= 30)
    return { tone: 'warn', label: `Expires in ${daysLeft} ${daysLeft === 1 ? 'day' : 'days'}` }
  return { tone: 'up', label: `Valid for ${daysLeft} days` }
}
