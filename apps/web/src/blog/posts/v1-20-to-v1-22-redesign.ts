import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "Three releases, one thread running through all of them: take the visual redesign that had so far only touched the marketing pages and bring it into the product itself. Not a rebrand — same tokens, same greens, same logic — just applied to screens that hadn't caught up yet.",
  },
  {
    type: 'h3',
    text: 'v1.20 — the dashboard gets a hero',
  },
  {
    type: 'p',
    text: "The dashboard used to open with a grid of per-type count cards — how many cron monitors, how many uptime monitors, and so on. Useful the first week, forgettable after that, because a count by itself doesn't tell you whether anything is wrong. It's replaced now with hero stats that actually answer that question: monitors healthy, average uptime, how many need attention right now, and SMS credits remaining. Below that, a needs-attention banner surfaces what's actually down or expiring instead of making you scan a table for it, a filterable table lists monitors across all five types in one place, and an upcoming-expirations panel gives SSL/domain renewals a dedicated spot instead of burying them in per-type pages.",
  },
  {
    type: 'p',
    text: 'All of it is backed by the existing monitor and billing queries — the redesign mockup this was built from used placeholder numbers, but nothing on the shipped dashboard is fake.',
  },
  {
    type: 'h3',
    text: 'v1.21 — the public status page catches up',
  },
  {
    type: 'p',
    text: "Checkmeup's public /status/:slug pages hadn't been touched visually since the MVP. This release gives them a token-driven header with a logo (or an initials avatar if you haven't uploaded one), a theme toggle that shares the same localStorage key as the rest of the app so light/dark stays consistent if someone bounces between your dashboard and your public page, a redesigned overall-status banner, and restyled monitor cards.",
  },
  {
    type: 'p',
    text: "One thing from the original mockup didn't make it: an active-incident timeline with a running history of past incidents. That needs a manual incident-authoring feature — writing a note, timestamping updates as you investigate — that Checkmeup doesn't have yet. What ships today is still fully automatic: down and resolved timestamps per monitor, no editorializing required, just not a full incident log with commentary.",
  },
  {
    type: 'p',
    text: "A same-day follow-up swapped the theme toggle's plain sun/moon text characters for the same SVG icons the sidebar already uses, and fixed a local-dev-only bug where hitting the API directly on :8080 404'd on the status page's own assets — the kind of thing that only shows up once you're testing outside the normal Vite dev proxy.",
  },
  {
    type: 'h3',
    text: 'v1.22 — Settings, Status Pages Admin, Monitors, and Maintenance get the once-over',
  },
  {
    type: 'p',
    text: "The last piece: Settings' notification-channel picker moved from a plain dropdown to an icon grid, with a matching icon badge on each connected channel in the list. It's still scoped to the five channel types that actually exist — Telegram, email, webhook, Slack, SMS. The redesign mockup this came from showed nine, including WhatsApp, Signal, Viber, and Teams, but none of those are built yet (they're separate, mostly-blocked epics), so the picker doesn't pretend they're one click away.",
  },
  {
    type: 'p',
    text: "Status Pages Admin and the five monitor types (Cron, Uptime, SSL, Domain, Port) turned out to be a smaller job than expected — every feature in their redesign mockups, including the monitor-picker/reorder UI and the status-badge Copy Markdown/HTML buttons, was already built and tested. What shipped is the visual delta: raised, uppercase-label table headers to match the new design system, and each monitor's status pill moved inline next to its name in the detail header instead of sitting on its own line below.",
  },
  {
    type: 'p',
    text: "The monitors mockup also proposed something bigger that didn't ship: collapsing Cron/Uptime/SSL/Domain/Port from five separate routes (twenty views total, each independently tested) into one unified page with tab-chips and shared list/detail/form components. Every field that page would need already exists on the backend — nothing about it is blocked technically — but rewriting twenty mature, working views into a handful of shared ones is a real architecture project, not a coat of paint. That one's staying on the shelf for now rather than getting bundled into a redesign pass.",
  },
  {
    type: 'p',
    text: "Maintenance windows got the same treatment as Status Pages Admin: the feature — scheduling a window, picking monitors across all five types, ending one early, the alert suppression that actually excludes covered monitors from checks while a window is active — was already built and tested end to end. The list view's table header picked up the same raised, uppercase-label treatment as everywhere else in this pass; nothing else needed to change.",
  },
  {
    type: 'h2',
    text: 'Also this release',
  },
  {
    type: 'ul',
    items: [
      'Resolved a CodeQL alert on the dashboard by extracting its row-building logic into a dedicated helper module, and cleaned up a few dead undefined-checks Codacy flagged along the way.',
      'Corrected an ambiguous line in the v1.19 release notes: the excerpt said the new downgrade enforcement pauses monitors "oldest-first," which reads backwards from what actually happens — newest-created monitors get paused, oldest stay active. The body text already had this right; only the excerpt needed the fix.',
    ],
  },
  {
    type: 'h2',
    text: 'Follow along',
  },
  {
    type: 'p',
    text: 'Team management — inviting a teammate into your org — is still next on the board. Releases land on this blog as they ship; the GitHub repo has the full commit history and architecture decision records if you want the why behind any of this.',
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
