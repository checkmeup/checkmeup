import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "Checkmeup is built for freelancers and small agencies, and their monitors come in bundles: one client has an uptime check, an SSL certificate, a domain and a nightly backup job; the next client has the same four again. Until now all of them sat in one flat list. With a dozen clients that's fifty rows, and when an alert arrived saying \"backup missed its ping\", you had to open the app to find out whose backup it was. v1.52 adds projects.",
  },
  {
    type: 'h2',
    text: 'One project per monitor',
  },
  {
    type: 'p',
    text: 'A project is a name, usually a client\'s, and each monitor of any type can be in one. You set it on the monitor\'s form, where "New project…" creates one without leaving the page, or on the new Projects page, where you pick a project\'s monitors from one list. Deleting a project never deletes monitors: they keep running, just without a project.',
  },
  {
    type: 'p',
    text: "I went back and forth on tags. Tags are more flexible: a monitor could be both \"Acme\" and \"production\". But a monitor with three tags has no single place in a grouped list, and once a project can carry settings of its own (more on that below), a monitor in two groups needs a rule for which group's settings win. One project per monitor keeps every answer simple. If cross-cutting groups like prod and staging turn out to matter, tags can come later on top of projects.",
  },
  {
    type: 'p',
    text: 'It\'s called "Project" rather than "Client" because not everyone monitors clients. A developer grouping their own side projects is just as much the user here.',
  },
  {
    type: 'h2',
    text: 'The monitor list, grouped',
  },
  {
    type: 'p',
    text: 'Once you have a project, /monitors shows one section per project, each with how many of its monitors need attention, and "No project" last. Sections collapse. A project filter sits next to the search box, and it lives in the URL, so ?project= gives each client\'s monitors a link of their own. If you never create a project, nothing changes: the list stays flat, exactly as before. On the overview, each monitor under "Needs attention" now shows its project above its name.',
  },
  {
    type: 'h2',
    text: 'Alerts that say whose they are',
  },
  {
    type: 'p',
    text: 'This was the part I wanted most. Every alert for a monitor in a project now starts with the project in brackets, on every channel: "[Acme] nightly-backup missed its ping" in Telegram, Slack and SMS, and in both the email subject and body. You know whose problem it is from the lock screen. Webhook events gain projectId and projectName fields, so your own automation can route by client.',
  },
  {
    type: 'p',
    text: "An alert for a monitor without a project is unchanged, byte for byte, and the webhook leaves the two new fields out rather than sending them empty. If you parse those payloads, nothing breaks.",
  },
  {
    type: 'h2',
    text: 'No plan limit',
  },
  {
    type: 'p',
    text: "Projects aren't a paid feature and don't count against anything. They cost nothing to keep, and putting a grouping feature behind an upgrade would only make the free plan more annoying to use. There's a flat cap of 1,000 per account, the same on every plan, which exists only so a script gone wrong can't create millions of them.",
  },
  {
    type: 'h2',
    text: "What's not in this release",
  },
  {
    type: 'ul',
    items: [
      'Per-project settings: default notification channels for a project\'s new monitors, a maintenance window for a whole project, a status page made from a project. That\'s the obvious next step, and it\'s why the one-project rule mattered.',
      'Per-project access for team members. Team accounts aren\'t built yet, but projects are the natural place to hang that on when they are.',
      'Projects in the public API. The API-key endpoints still report on single monitors.',
    ],
  },
  {
    type: 'h2',
    text: 'Also this release',
  },
  {
    type: 'ul',
    items: [
      "The project of a monitor is checked by the database itself: the link from a monitor to its project includes the account, so a monitor can't end up in another account's project even through a bug in the code above it.",
      'The production image now installs its frontend dependencies from the lockfile, exactly as tested, instead of resolving every version range again at build time.',
      "The landing page's live status strip dropped its \"average response\" from the header. A single average over 90 days said little; per-day response times are still in each day's tooltip.",
      'Behind the scenes: a test that failed now and then, because two clock reads a millisecond apart could round a 10-minute run down to 9, now reads the clock once.',
    ],
  },
  {
    type: 'h2',
    text: 'Follow along',
  },
  {
    type: 'p',
    text: "The per-project settings above are where projects go next. Default alert channels per project is the one I'd expect first, so a new Acme monitor alerts Acme's Slack without picking it each time. The full commit history and every architecture decision record, including the one for projects, are on GitHub.",
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
