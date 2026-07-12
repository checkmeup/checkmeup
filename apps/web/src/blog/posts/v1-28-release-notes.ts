import type { BlogPost } from '../types'

export const post: BlogPost = {
  slug: 'v1-28-release-notes',
  title: 'v1.28: Manual Incidents, and a Real "Powered by" Toggle',
  date: 'July 12, 2026',
  readTime: '4 min read',
  excerpt:
    'Two things landed on status pages today: you can now declare and narrate an incident yourself instead of only ever showing raw monitor up/down, and paid plans can finally turn off the "Powered by Checkmeup" footer — a feature the pricing page has quietly promised for a while.',
  content: [
    {
      type: 'p',
      text: "Status pages have always been a straight read-out of monitor state: green means the last check passed, red means it didn't. That's accurate, but it's not always what actually happened. A slow API isn't \"down.\" A known degradation you're already fixing isn't a mystery to your visitors if you tell them so. Today's release adds a way to say that — plus a smaller, overdue fix to something the pricing table has said since before it was true.",
    },
    {
      type: 'h3',
      text: 'Declare an incident, not just a monitor state',
    },
    {
      type: 'p',
      text: "There's a new Incidents page now, separate from monitors. Declare one with a title, an initial message, a severity (Minor / Major / Critical), and any combination of affected monitors — cron, uptime, SSL, domain, or port, mixed freely, since a single incident is often not one monitor's problem. It shows up on the public status page immediately, above the monitor list, and a Major or Critical incident escalates the page's overall status banner the same way a real monitor-down already does.",
    },
    {
      type: 'p',
      text: 'From there it\'s a running log: post updates as you work it, and the status moves forward through Investigating → Identified → Monitoring → Resolved with each one. Visitors see the full timeline, newest first. Mark it Resolved and it moves off the active list into a paginated incident history further down the page — the "Past incidents" section that already existed for automatic monitor-down incidents now carries these too.',
    },
    {
      type: 'p',
      text: "The one thing this deliberately doesn't do is touch monitor state. Declaring, updating, or resolving an incident never flips a monitor's own up/down status, and it never triggers an alert on any channel — this is a status-page narration tool for MVP, not a new way to get paged. If you're also running a maintenance window on the same monitor, declaring an incident against it warns you first instead of silently layering two different \"something's going on\" states on top of each other; you can still confirm and go ahead if that's genuinely what's happening.",
    },
    {
      type: 'h3',
      text: 'A "Powered by Checkmeup" you can actually turn off',
    },
    {
      type: 'p',
      text: 'The pricing page has listed "white-label status pages" as a paid-plan feature for a while. Until today that wasn\'t quite true — every status page rendered the same "Powered by Checkmeup" footer with FAQ/Terms/Privacy links underneath it, regardless of plan, because there was no toggle to turn it off. It\'s real now: each status page has a hide-branding checkbox in its edit settings, available on Solo and above. Flip it and the public footer drops down to just the "Last updated" line — no Checkmeup branding, no outbound links.',
    },
    {
      type: 'p',
      text: "It's a per-page setting, not account-wide, on purpose — an org can run more than one status page, and there's no reason a client-facing page and an internal one need the same branding decision. Hobby stays as it was: the checkbox is visible but disabled, with a link to upgrade, and the API rejects a direct attempt to set it with the same 402 plan_limit_reached response every other plan-gated feature here uses. Downgrade back to Hobby later and it's cleared automatically, not left in a stale \"paid-only setting on a free plan\" state.",
    },
    {
      type: 'h2',
      text: 'Also this release',
    },
    {
      type: 'ul',
      items: [
        'A bit more breathing room above the "Past incidents" heading on the public status page — it was sitting right on top of the monitor list with almost no separation from the section above it.',
        "Full test coverage added for the incident-management backend handler and all three new frontend views, which shipped without much of either in the original PR — caught by this repo's Codacy coverage-variation check before merge, not after.",
        "CreateIncident, the one handler doing the most at once (resolving monitors, checking for a maintenance-window overlap, and writing three related rows), got split into four smaller single-purpose functions to bring it back under this repo's complexity thresholds.",
      ],
    },
    {
      type: 'h2',
      text: 'Follow along',
    },
    {
      type: 'p',
      text: "Team management — inviting a teammate into your org — is next up. Releases land on this blog as they ship; the GitHub repo has the full commit history and architecture decision records if you want the why behind any of this, including ADR-035's reasoning for gating branding removal specifically to paid plans.",
    },
    {
      type: 'signature',
      text: '— Andrew',
    },
  ],
}
