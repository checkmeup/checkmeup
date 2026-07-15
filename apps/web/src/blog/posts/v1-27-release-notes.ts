import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "Not every release is a new feature. This one is entirely under the hood — the kind of work that doesn't show up as a button anywhere, but is the reason checks keep running cleanly as the number of monitors grows.",
  },
  {
    type: 'h3',
    text: 'Monitor checks reuse connections instead of rebuilding them',
  },
  {
    type: 'p',
    text: 'Uptime, SSL, and port checks each used to build a fresh HTTP client (or raw dialer) on every single check — every 30 seconds to every few minutes, per monitor, forever. That works fine at small scale, but it means constantly paying the cost of a new connection setup instead of reusing one, and it adds socket churn that only grows with how many monitors an account has. All three check types now share one client per process, built once, with connection reuse enabled for repeat checks against the same host. Nothing about what gets checked or how results are reported changed — this is purely about how the engine gets there.',
  },
  {
    type: 'h3',
    text: 'The database pool now has an explicit ceiling',
  },
  {
    type: 'p',
    text: "Every check writes its result through the same connection pool that also serves the dashboard, the public API, and every status page — and until this release, that pool had no explicit size, just whatever the driver's default happens to be. As the number of monitors and customers grows, that default was the most likely actual ceiling on how much load the system could take, well ahead of anything the check scheduler itself does. It's set explicitly now, with real headroom above current usage.",
  },
  {
    type: 'h2',
    text: 'Also this release',
  },
  {
    type: 'ul',
    items: [
      'Alert delivery — Telegram, Slack, webhook, and SMS — shares more of its underlying code now, instead of each channel hand-rolling the same send-and-check-the-response logic slightly differently.',
      'SSL and domain expiry alerts (the "expires in N days" / "has expired" messages) now share the same underlying code — they were doing the exact same thing with different field names, just written out twice.',
      "The notification-channel settings form got a cleanup pass: the parts that never actually change (channel labels, icons, validation rules) moved out of the reactive form logic and into their own place, so the channel list doesn't have to load the whole edit form just to show an icon.",
      'Founder-facing internal notifications (the in-app feedback form) are now handled separately from the email code that actually reaches customers — different audiences, now different code paths.',
    ],
  },
  {
    type: 'h2',
    text: 'Follow along',
  },
  {
    type: 'p',
    text: 'Next up is back to user-facing work — this release was about making sure the engine underneath can keep scaling quietly rather than becoming the next thing that needs fixing under pressure. Full commit history and the architecture decisions behind Checkmeup are public on GitHub if you want the details behind any of this.',
  },
  { type: 'signature', text: '— Andrew' },
]
