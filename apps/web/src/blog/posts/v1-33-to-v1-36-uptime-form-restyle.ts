import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "Four versions since the last release notes, and honestly, it's the quietest stretch checkmeup has had — one real product change, two new posts on the blog, and a chunk of time that went into how I actually work on this codebase day to day rather than what it does for you. Here's all of it.",
  },
  {
    type: 'h2',
    text: 'The uptime monitor form catches up to the redesign',
  },
  {
    type: 'p',
    text: "The homepage, dashboard, and status pages got the \"Claude Design\" treatment a few releases back; the uptime monitor's create/edit form hadn't caught up yet. It has now. Accepted status codes used to be a stack of checkboxes — one row per code — and now they're toggleable pill chips, the same accent/wash/border treatment the monitor table's own type filters already use, so nothing about the palette is new, just more consistent. The advanced check settings section swapped a native <details>/<summary> disclosure (inconsistent triangle glyph, no animation, looked out of place next to everything else) for a custom one with a rotating chevron and a raised card background that actually matches the rest of the form.",
  },
  {
    type: 'p',
    text: "Notification channel selection on that same form is untouched for now — it's next, not forgotten.",
  },
  {
    type: 'h2',
    text: 'Also this cycle',
  },
  {
    type: 'ul',
    items: [
      '"The 3 AM SSL Cert That Cost Us a Client" — a disaster-story post about a two-month-silent renewal cron job and the six hours of bounced checkout traffic that followed. If you run cron-driven cert renewal anywhere, it\'s worth the five minutes.',
      '"Port (TCP) Monitoring: Beyond HTTP Checks" — a guide to what a raw TCP connect check actually proves (and doesn\'t) for anything that isn\'t speaking HTTP, closing out the guide series alongside the cron and uptime ones.',
      "A production Dockerfile fix: the base image's ca-certificates pin had drifted, which is the kind of thing that's invisible until a dependency somewhere needs a cert it no longer trusts.",
      "A pass on how I work with Claude Code on this repo — deterministic guardrails (never commit on main, never run a deploy command, never force-push main, a secrets scan before every push) instead of hoping written instructions get followed, plus fixing a couple of the audit scripts that check this codebase's own claims about itself after they drifted from refactors.",
    ],
  },
  {
    type: 'h2',
    text: 'Follow along',
  },
  {
    type: 'p',
    text: "Nothing on the roadmap changed because of a quiet cycle — the notification-channel section of that same uptime form is next, and there's more of the design rollout to finish elsewhere. Full commit history and every ADR are on GitHub if you want the unfiltered version.",
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
