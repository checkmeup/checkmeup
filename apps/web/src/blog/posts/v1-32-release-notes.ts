import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "An uptime monitor has always meant one specific thing here: a GET request, a 10-second timeout, and exactly HTTP 200 counted as healthy. That was a deliberate MVP simplification (ADR-014 called it out at the time, and even flagged the strict-200 rule as something to relax if it ever caused real false alarms) — and it did, eventually. A 201 from a create endpoint or a 204 from a delete endpoint is a perfectly healthy response that was quietly getting alerted on as down. v1.32 makes all three of those hardcoded assumptions configurable per monitor, defaulting to exactly today's behavior so nothing changes for a monitor that doesn't touch the new fields.",
  },
  {
    type: 'h3',
    text: 'Method, timeout, and accepted status codes',
  },
  {
    type: 'p',
    text: 'Each uptime monitor now has three settings tucked under an "Advanced check settings" section on create/edit, collapsed by default so they don\'t compete with name/URL/interval for attention: the HTTP method (GET, HEAD, or POST), a request timeout, and a multiselect of which status codes count as up. An API that legitimately returns 201 or 204 on success can now say so, instead of the monitor grading it down every single check.',
  },
  {
    type: 'h3',
    text: "The timeout field isn't new — it was hiding as something else",
  },
  {
    type: 'p',
    text: 'checkmeup already had a "max response time" field, added back with assertion checks — an optional, per-monitor SLA threshold. But it only ever ran after a successful response: the request was allowed to complete against a fixed, shared 10-second timeout, and only then would elapsed time get compared to the threshold and fail the check retroactively. That\'s two knobs doing overlapping jobs. A monitor with a 3-second SLA still let the request run the full 10 seconds before deciding it was too slow. v1.32 collapses them into one — that field is now required, defaults to 10 seconds (today\'s exact behavior), and is used directly as the actual connection timeout. A slow response now aborts at the configured ceiling instead of finishing and then being judged after the fact.',
  },
  {
    type: 'h3',
    text: 'Validated, not just accepted',
  },
  {
    type: 'p',
    text: 'The method is restricted to a fixed whitelist — anything else is rejected with a clear error, not silently coerced to GET. The timeout is bounded to 1–30 seconds, and every accepted status code has to fall in the valid HTTP range. All of it is enforced server-side as the authoritative check, with the same bounds mirrored on the form so a bad value shows up as an inline error instead of a raw API failure.',
  },
  {
    type: 'h2',
    text: 'Follow along',
  },
  {
    type: 'p',
    text: 'Releases land on this blog as they ship; the GitHub repo has the full commit history and architecture decision records if you want the why behind any of this, including the original 200-only tradeoff this release revisits.',
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
