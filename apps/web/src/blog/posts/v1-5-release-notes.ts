import type { BlogPost } from '../types'

export const post: BlogPost = {
  slug: 'v1-5-release-notes',
  title: 'v1.5: Cleanup, a Plan-Limit Bug, and More Tests',
  date: 'June 20, 2026',
  readTime: '2 min read',
  excerpt:
    'Another quiet release: one real bug in how uptime monitor intervals get enforced on edit, a dark-mode color fix, a data-fetching cleanup across the dashboard, and more backend test coverage.',
  content: [
    {
      type: 'p',
      text: 'v1.5 is in the same spirit as v1.4 — no new features, just the codebase getting steadier under the hood. The one item worth flagging is a real bug in plan enforcement; everything else is cleanup that should be invisible if it went well.',
    },
    {
      type: 'h3',
      text: 'The plan-limit bug',
    },
    {
      type: 'p',
      text: "Creating an uptime monitor already enforced your plan's minimum check interval correctly. Editing one didn't: the update handler floored any interval below 10 minutes straight to 10, ignoring the plan entirely. For paid plans with a 1-minute minimum, that meant you couldn't dial an existing monitor down below 10 minutes even though you were paying for faster checks. For Hobby's 5-minute minimum, a too-low request was silently rewritten to 10 instead of being rejected with a clear plan-limit error. Updates now run through the same clamp Create uses, so the behavior — and the error you see when you hit it — is consistent between creating and editing a monitor.",
    },
    {
      type: 'h3',
      text: 'Dark mode color fix',
    },
    {
      type: 'p',
      text: "A handful of status dots and accent-colored text were still hardcoded hex values instead of design tokens, left over from before the theme system existed. They've been swapped to the existing --status-* tokens, and a new --on-accent token now covers text and icons that sit on an accent-colored background. Mostly affects the landing page, pricing page, and status page badges in dark mode.",
    },
    {
      type: 'h3',
      text: 'Data-fetching cleanup',
    },
    {
      type: 'p',
      text: "Every dashboard view — cron, uptime, SSL monitors, maintenance windows, status pages, settings, billing — used to fetch and cache its own data ad hoc. They now all go through the same TanStack Query composables, which collapses a lot of repeated loading/error/refetch boilerplate into one consistent pattern. No behavior change you'd notice; the win is fewer places for that logic to drift out of sync.",
    },
    {
      type: 'h3',
      text: 'More backend tests',
    },
    {
      type: 'ul',
      items: [
        "Table-driven tests now pin every plan-limit boundary in the billing package — Hobby, Solo, Enterprise, both keyword and interval limits — so the plan-limit bug above can't reintroduce itself unnoticed.",
        'The config package — environment variable parsing, defaults, and required-vs-optional settings — now has unit test coverage for the first time.',
      ],
    },
    {
      type: 'h2',
      text: 'Follow along',
    },
    {
      type: 'p',
      text: 'Releases land here on the blog. The GitHub repo has the full commit history and architecture decision records if you want the why behind any of this.',
    },
    {
      type: 'signature',
      text: '— Andrew',
    },
  ],
}
