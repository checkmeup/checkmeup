import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "v1.12 shipped port monitoring, a new monitor type. This release doesn't add a sixth — it's two smaller, unrelated changes that landed the same day: a homepage redesign, and a license change worth explaining plainly rather than burying in a diff.",
  },
  {
    type: 'h3',
    text: 'A new homepage',
  },
  {
    type: 'p',
    text: 'The hero section\'s static screenshot is now a live-looking HTML mockup instead — a fake browser window with a pulsing status dot, the same three-stat uptime grid the real monitor detail page shows, and a response-time sparkline. It never goes stale the way a screenshot does, and it reads as more "alive" than a flat image. The status-pages section got the same treatment: a mini status page with a real "all systems operational" banner and per-monitor progress bars, and the embeddable status-badge callout moved into that card instead of sitting in its own orphaned block below it.',
  },
  {
    type: 'p',
    text: "The feature grid also got a sixth card for public status pages — previously status pages were only pitched in their own dedicated section further down the page, not alongside the five monitor types up top. And the footer was reorganized into three grouped link columns (Docs/FAQ/Pricing/Blog, About/Sign in/Sign up, Terms/Privacy) instead of one long wrapping row, matching the rest of the site's information density.",
  },
  {
    type: 'h3',
    text: 'Business Source License',
  },
  {
    type: 'p',
    text: "Checkmeup's code was MIT-licensed. It's now BUSL 1.1 (Business Source License), the same model MariaDB, Sentry, and CockroachDB use. In practice, almost nothing changes for almost everyone: you can still read the code, self-host it, modify it, and use it in production for your own monitoring — including monitoring your clients' systems. The one thing BUSL restricts is reselling Checkmeup itself, unmodified or modified, as a competing hosted monitoring service. On July 1, 2030, the license automatically converts to Apache License 2.0 and the restriction disappears entirely.",
  },
  {
    type: 'p',
    text: 'Why bother, for a small bootstrapped product? Because "built in the open" and "anyone can legally clone this and undercut you on hosting" are two different promises, and conflating them isn\'t honest. BUSL keeps the first promise — the source stays public, the architecture decisions stay documented — without handing a competitor a free SaaS business. The About page now says so explicitly, with a link to the full license text.',
  },
  {
    type: 'h2',
    text: 'Also this release',
  },
  {
    type: 'p',
    text: 'A few smaller things: a Codacy pass on the port monitoring code from v1.12 knocked two handler functions and three Vue submit()/formatting functions down from above-threshold cyclomatic complexity to well within it, by extracting the validation and plan-limit logic into named helpers — same behavior, easier to read. The FAQ\'s billing category had a stale "what counts as a monitor" answer that listed cron/uptime/SSL/domain but not port, missed in the doc pass that updated the identical question elsewhere on the same page — now consistent.',
  },
  {
    type: 'h2',
    text: 'Follow along',
  },
  {
    type: 'p',
    text: 'Microsoft Teams alerts are still next on the board. Releases land on this blog as they ship; the GitHub repo has the full commit history and architecture decision records if you want the why behind any of this.',
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
