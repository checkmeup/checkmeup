import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "v1.6's centerpiece is a rework of how alerts get delivered: notification channels. It's the kind of change that's invisible if you only ever used one Telegram chat and one email address — your monitors keep alerting exactly the same way after upgrading — but it unlocks something the old model couldn't do at all.",
  },
  {
    type: 'h3',
    text: 'Notification channels',
  },
  {
    type: 'p',
    text: "Until now, an org had exactly one Telegram chat ID and one alert email, set globally in Settings, and every monitor alerted through both. That's fine if you're a single person watching everything in one place — not so fine if you want a critical payment monitor to page a Slack channel while a low-priority one just emails you, or if you want two separate Telegram chats for two different teams.",
  },
  {
    type: 'p',
    text: 'Settings now has a notification channels list instead of two fixed fields. Add as many channels as you want — multiple Telegram chats, multiple alert emails, more types coming — give each one a name, and test it before saving, same as the old Telegram/email test buttons worked. Every monitor (cron, uptime, SSL) gets a channel picker alongside its existing alert mute toggle, so you choose per monitor which channels fire. A new monitor defaults to all of your org\'s enabled channels attached, matching the old implicit "alerts go everywhere" behavior, so nothing changes unless you deliberately narrow it down.',
  },
  {
    type: 'p',
    text: "One deliberate behavior change: a monitor with zero channels attached no longer goes silent. It falls back to emailing every user in your org instead. The old model couldn't go silent either — Telegram and email were always both configured or both off — so this fallback keeps that guarantee intact under the new, more flexible setup. The migration also backfills your existing Telegram chat and alert email straight into the new channel model and attaches them to every monitor that had alerts on, so upgrading doesn't require touching anything.",
  },
  {
    type: 'p',
    text: "This is foundational, not the finish line — it's exactly what unblocks webhook, Slack, Teams, and the other channel types queued up next. They each just add a new channel type on top of this model instead of inventing their own org-level field.",
  },
  {
    type: 'h3',
    text: 'Keyword monitoring is free on every plan',
  },
  {
    type: 'p',
    text: 'Keyword monitoring on uptime checks — alerting on "down" or "maintenance" page text even when the HTTP status looks fine — used to be gated behind a plan check. Every plan had quietly returned "allowed" for a while, so the gate was dead weight pretending to be a limit. It\'s gone now: keyword monitoring is available on Hobby same as everywhere else, and the in-app docs and pricing copy have been corrected to match.',
  },
  {
    type: 'h3',
    text: 'Small fixes',
  },
  {
    type: 'ul',
    items: [
      "The channel picker now warns you if every channel you've selected for a monitor is individually disabled — previously it let you save a monitor that looked configured but couldn't actually alert anyone.",
      'The Docker image now runs as a dedicated non-root user instead of root.',
      "The SSL monitor's own TLS check now enforces TLS 1.2 as a floor when connecting to the certificate it's checking — closing the kind of gap a tool watching for *your* SSL hygiene shouldn't have in its own.",
    ],
  },
  {
    type: 'h3',
    text: 'Test coverage',
  },
  {
    type: 'p',
    text: 'The email-sending package and the frontend bootstrap file both got their first tests this release — small, focused additions, but two more corners of the codebase that no longer rely on manual checking alone.',
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
]
