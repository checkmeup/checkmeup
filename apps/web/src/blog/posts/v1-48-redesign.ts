import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "v1.48 is the biggest visual change since Checkmeup launched: every screen, in both light and dark, from the dashboard to the public status page. The direction is calm and light-first, with dense data tables where you need them, and very little decoration everywhere else.",
  },
  {
    type: 'h2',
    text: 'A dashboard that starts with a sentence',
  },
  {
    type: 'p',
    text: "The sidebar is gone in favor of a top navigation bar, which gives the content the full width of the screen. The dashboard now opens with a plain sentence, like \"3 monitors need you.\", and explains each problem in words instead of status codes. Every monitor is listed below it with problems first, plus tabs and search.",
  },
  {
    type: 'p',
    text: "There's now one /monitors list for all six monitor types, filtered by type, instead of six separate lists. The old per-type URLs still work and redirect there, so bookmarks don't break.",
  },
  {
    type: 'h2',
    text: 'Check strips',
  },
  {
    type: 'p',
    text: "Every uptime, port, DNS and cron monitor in a list now shows a strip of its last 48 hours, one tick per hour. A row of green with one red tick tells you more at a glance than a status label ever did. The strip comes from a new check-history endpoint, scoped to your organization like every other query.",
  },
  {
    type: 'h2',
    text: 'One layout for six monitor pages',
  },
  {
    type: 'p',
    text: "All six monitor detail pages are rebuilt on the same parts: a header, a stat grid, the check strip, a response chart and a log table. Pause, resume and delete now come from one implementation instead of six copies. That's not just tidier: when deleting a cron monitor failed, it used to show no error at all, and now it does, like the others.",
  },
  {
    type: 'h2',
    text: 'Status pages, badges and colors',
  },
  {
    type: 'p',
    text: "The public status page your clients see got the same treatment: the app's fonts and colors, an overall-status headline, and hairline monitor rows with the 90-day bar. Status colors now switch per theme, and every one of them is readable as text at WCAG AA contrast, in light and in dark. The embeddable status badges use the same colors, drawn flat. The fonts are Geist and Geist Mono.",
  },
  {
    type: 'p',
    text: "What I deliberately didn't touch: the logo and the favicon. A redesign is a good moment to fiddle with a logo, which is exactly why I didn't.",
  },
  {
    type: 'h2',
    text: 'Also this release',
  },
  {
    type: 'ul',
    items: [
      'Timestamps from the API are now always sent in UTC. On a server in another time zone, every time used to shift by that zone\'s offset.',
      'Incident times on public status pages are printed in UTC, and the grid layout titles each monitor.',
      'Paused monitors no longer count toward the dashboard\'s "all fine" count or its uptime average, and a cron monitor now says when its ping was due.',
      'Settings and billing are laid out as labeled sections. Forms, incidents and maintenance windows dropped their boxes for hairlines and tables.',
      'Under the hood, about 700 inline style attributes became regular CSS classes, which is what made a change this size manageable.',
    ],
  },
  {
    type: 'h2',
    text: 'Follow along',
  },
  {
    type: 'p',
    text: "Next up is the landing page, which got the redesign's colors here but not yet a rethink of what it actually says. Full commit history and every architecture decision record are on GitHub if you want the unfiltered version.",
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
