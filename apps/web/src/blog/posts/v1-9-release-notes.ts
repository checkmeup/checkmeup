import type { BlogPost } from '../types'

export const post: BlogPost = {
  slug: 'v1-9-release-notes',
  title: 'v1.9: Status Badges and Assertion Checks',
  date: 'June 23, 2026',
  readTime: '4 min read',
  excerpt:
    'Two additions this release: embeddable SVG status badges you can drop into a README or site footer to show live monitor status, and structured assertion checks for uptime monitors — JSON field assertions and a response-time threshold that actually fail the check, not just annotate it.',
  content: [
    {
      type: 'p',
      text: "v1.8 introduced domain expiry monitoring and closed the gap between what Checkmeup tracks and what status pages could show. This release ships two features flagged in the v1.8 post as 'next up': public status badges and assertion-based uptime checks. Neither adds a new monitor type — both deepen what the existing four types can do.",
    },
    {
      type: 'h3',
      text: 'Status badges',
    },
    {
      type: 'p',
      text: "Each status page now has two badge endpoints: one for the page's overall status, one per attached monitor. Both return hand-rendered SVGs — the same operational/degraded/outage wording as the page banner, no external badge service, no new infrastructure. A GET to `/status/:slug/badge.svg` returns the page-level badge; `/status/:slug/badge/:monitor_id.svg` returns a single monitor's badge. Responses are cached for 60 seconds so an embedded badge on a high-traffic docs page doesn't fan out into constant re-checks.",
    },
    {
      type: 'p',
      text: 'The status page editor gets a Badges section listing every badge available for that page — the page-level one at the top, one row per attached monitor below. Each row shows a live preview and two copy buttons: one for the Markdown embed (`![status](...)`), one for HTML (an img element inside an anchor linking back to the public page). Domain monitors, added in v1.8, appear here alongside the other three types.',
    },
    {
      type: 'h3',
      text: 'Domain monitors on status pages',
    },
    {
      type: 'p',
      text: "When domain monitoring shipped in v1.8 it wasn't wired into public status pages yet — you could track domain expiry internally but not surface it to visitors. That gap is now closed: domain monitors can be added to a status page the same way cron, uptime, and SSL monitors can, show the same status/days-remaining display on the public page, and are included in the page-level badge's overall status calculation.",
    },
    {
      type: 'h3',
      text: 'Assertion checks for uptime monitors',
    },
    {
      type: 'p',
      text: 'A 200 response can still be wrong. A health endpoint that always returns `{"status":"degraded"}` or a payment API that silently starts responding in 8 seconds are failures even when the HTTP status code looks clean. Assertion checks are the answer: structured conditions on the response beyond the status code.',
    },
    {
      type: 'p',
      text: 'Add one or more JSON assertions on a monitor. Each assertion has a path (dot-notation: `data.status` or `$.healthy`), a comparator (`equals`, `not_equals`, `contains`, `greater_than`, `less_than`), and an expected value. All assertions must pass — the first one that fails stops evaluation and records the exact reason. A response that isn\'t valid JSON when a JSON assertion is configured fails with "response is not valid JSON" rather than silently passing. The 512 KB body cap from the keyword check applies here too.',
    },
    {
      type: 'h3',
      text: 'Response-time threshold',
    },
    {
      type: 'p',
      text: 'Set a max response time in milliseconds. If the server responds but takes longer than the threshold, the check fails with "response time exceeded" — a distinct reason from a connection-level timeout, which already failed before this release. This is separate from the hard 10-second connection timeout: a 2 000 ms threshold fails a 3-second response that a clean status code would otherwise pass.',
    },
    {
      type: 'h3',
      text: 'How all the conditions fit together',
    },
    {
      type: 'p',
      text: 'Each check evaluates conditions in a fixed order: status code → keyword → JSON assertions → response-time threshold. The first failure stops the chain and is the recorded reason. Same two-consecutive-failures state machine and same alerting channels as before — no new monitoring model, just more things that can constitute a failure. A monitor with no assertions configured behaves exactly as before.',
    },
    {
      type: 'h3',
      text: "What's not in this release",
    },
    {
      type: 'p',
      text: "Multi-step (chained) API checks — verifying a login then calling an authenticated endpoint in one monitor — were scoped out. The implementation is a meaningful jump in complexity (request ordering, value templating between steps, new UI surface) and the remaining four assertion user stories didn't depend on it. It's a reasonable follow-on once the simpler assertion path has been in use for a while.",
    },
    {
      type: 'h2',
      text: 'Follow along',
    },
    {
      type: 'p',
      text: 'Multi-region checks and chained API requests are next on the radar. Releases land on this blog as they ship; the GitHub repo has the full commit history and architecture decision records if you want the why behind any of this.',
    },
    {
      type: 'signature',
      text: '— Andrew',
    },
  ],
}
