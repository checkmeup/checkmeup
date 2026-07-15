import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "Not every release is a new feature. Sometimes it's paying down the quiet cost of a codebase that's grown for a few months straight — five monitor types, eight alert channels, a public API, billing, status pages — without anyone stopping to check whether any single file had quietly become the place everything gets bolted onto.",
  },
  {
    type: 'h3',
    text: "What you'll actually notice",
  },
  {
    type: 'p',
    text: "The dashboard sidebar and the site footer now show which version of Checkmeup you're looking at. Small, but genuinely useful — if something looks off and you want to report it, or you're just curious whether a fix has actually reached production yet, it's right there instead of a mystery.",
  },
  {
    type: 'h3',
    text: 'What happened behind the scenes',
  },
  {
    type: 'p',
    text: "The bigger piece of this week was building a small internal tool that audits Checkmeup's own source for a specific smell: a function or a file quietly taking on more than one job. Not a style linter — those already run — but something that flags a Go function branching in fifteen different directions, or a Vue page that's grown to over a thousand lines because six unrelated features all live in the same file.",
  },
  {
    type: 'p',
    text: 'Run against the real codebase, it found six of them: two Go handler functions with more branching than they should carry, a background worker file that had grown to cover cron checks, uptime checks, SSL checks, domain checks, and port checks all in one place, and four Vue pages/components that had each accumulated more responsibility than they should — the dashboard, the docs page, the marketing homepage, and the notification-channel settings form.',
  },
  {
    type: 'p',
    text: "Every one got split along its actual seams rather than an arbitrary line count. The worker file became five files, one per monitor type, sharing the alert-dispatch code that's genuinely common to all of them. The docs page — sixteen unrelated topics that happened to live in one file because nobody had separated them yet — became sixteen small components, one per topic, which is the same seam the sidebar navigation already implied. The notification-channel form's few hundred lines of state and validation logic moved into a reusable composable, separate from the markup that renders it.",
  },
  {
    type: 'p',
    text: "None of it changes what Checkmeup does. Every one of these was verified behavior-for-behavior against the existing test suite before merging — the point isn't new functionality, it's that the next feature that touches any of these files starts from something smaller and easier to reason about, instead of one more thing bolted onto an already-overloaded file.",
  },
  {
    type: 'h2',
    text: 'Also this release',
  },
  {
    type: 'ul',
    items: [
      'The version shown is the short tag (e.g. v1.23) even when the underlying build string carries extra commit metadata — nobody needs to see a commit hash to know what shipped.',
      "The audit tool itself is now a permanent fixture, not a one-off script — it'll get run periodically going forward rather than waiting for the codebase to get uncomfortable again.",
    ],
  },
  {
    type: 'h2',
    text: 'Follow along',
  },
  {
    type: 'p',
    text: 'Next up is back to feature work — the decision backlog has a few things queued (zombie-job and overlap detection for cron monitors among them). Full commit history and the architecture decisions behind Checkmeup are public on GitHub, including the audit tool from this release, if you want to see exactly what it flags and why.',
  },
  { type: 'signature', text: '— Andrew' },
]
