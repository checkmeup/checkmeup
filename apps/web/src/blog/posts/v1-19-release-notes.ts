import type { BlogPost } from '../types'

export const post: BlogPost = {
  slug: 'v1-19-release-notes',
  title: 'v1.19: SMS Alerts, and a Downgrade That Actually Downgrades',
  date: 'July 4, 2026',
  readTime: '5 min read',
  excerpt:
    "checkmeup can text you now — an eighth alert channel via Twilio, with a monthly credit quota so a flapping monitor can't turn into a surprise bill. And a smaller but overdue fix: downgrading your plan used to leave every over-limit monitor running forever. Now it actually enforces the new limit — pausing the newest monitors first and keeping your oldest ones active.",
  content: [
    {
      type: 'p',
      text: "Two things shipped today that don't have much to do with each other except that they both touch the plan-limits system: SMS alerts, and a fix to how downgrading a plan actually behaves. The SMS one is the feature people will notice. The downgrade one is the one that should have worked this way from the start.",
    },
    {
      type: 'h3',
      text: 'SMS alerts, priced like everything else here',
    },
    {
      type: 'p',
      text: 'Telegram, email, Slack, and webhooks are all free to send — checkmeup eats the cost of an API call either way. SMS is the first channel where that\'s not true: every message has a real, variable, per-destination cost through Twilio. That changes the design question from "how do we send a text" to "how do we stop a flapping monitor from generating a bill nobody agreed to."',
    },
    {
      type: 'p',
      text: 'The answer is the same one checkmeup already uses for monitors and status pages: a flat quota bundled into the plan price, not metered pass-through billing. Solo gets 10 SMS credits a month, Startup 30, Enterprise 100 — Hobby gets none, same as every competitor I looked at gates SMS to paid plans. Run out mid-month and that one channel just goes quiet for the rest of it — every other channel attached to the monitor still fires normally, and if SMS was the only one you had configured, the alert falls back to your account email instead of silently disappearing. A monitor going down should never fail to tell you just because one channel ran dry.',
    },
    {
      type: 'p',
      text: 'Setup is a phone number plus an explicit consent checkbox — not just providing a number, an actual "I agree to receive automated texts" checkbox with a timestamp stamped server-side at the moment you check it. That\'s not friction for its own sake: TCPA-style rules in the US (and equivalents elsewhere) require real opt-in before sending automated texts, even purely informational ones like a downtime alert. Down and recovery messages are kept to a single 160-character segment where possible, so one alert never quietly becomes two credits.',
    },
    {
      type: 'h3',
      text: 'One thing this shipped without, on purpose',
    },
    {
      type: 'p',
      text: "The original design called for destination-weighted credits — a US text costs one credit, an expensive-to-deliver destination costs three, so the quota tracks actual cost exposure instead of just message count. That needs a hand-maintained table of which countries land in which cost band, built from Twilio's own pricing pages. That's data entry, not code, and it was the one piece deliberately left for a follow-up rather than blocking the rest of the feature on it. Every credit costs 1 for now, regardless of where it's going. The plumbing (a per-send cost parameter, not a hardcoded 1 baked into the query) is already there for whenever that table gets built.",
    },
    {
      type: 'h3',
      text: "Downgrading used to be a promise checkmeup didn't keep",
    },
    {
      type: 'p',
      text: 'Before today, downgrading from Startup to Hobby with 40 active monitors did... nothing to those 40 monitors. They kept running exactly as before. The only enforcement was on creating new ones — try to add a 41st monitor on a 10-monitor plan and you\'d get blocked, but the 40 that predated the downgrade never felt it. That was a deliberate MVP-era choice (never delete anything on downgrade), but "never delete" had quietly become "never actually enforce," which is a different and much more generous promise than intended.',
    },
    {
      type: 'p',
      text: 'The fix keeps the "nothing is ever deleted" part and drops the "forever" part. Downgrade past your monitor or notification-channel limit now, and the newest-created ones beyond the new limit get paused or disabled automatically — not deleted, just switched off, exactly like manually pausing a monitor yourself. The oldest ones stay active, on the theory that whatever you set up first is usually what you care about most. Try to resume a paused monitor or re-enable a disabled channel past the limit and you\'ll get the same inline upgrade prompt as trying to create a new one over the limit — the enforcement is symmetric now, not just one-directional.',
    },
    {
      type: 'p',
      text: "Status pages are the one thing this doesn't touch — there's no pause primitive for a status page today, so they stay fully grandfathered on downgrade, same as before. That's a narrower gap than \"every resource type,\" and a smaller one to close later if it turns out to matter.",
    },
    {
      type: 'h2',
      text: 'Also this release',
    },
    {
      type: 'ul',
      items: [
        'Dashboard has a new "SMS sent" card and the Billing page shows an SMS-credit usage bar, same style as the existing monitor/status-page/channel usage.',
        '"Send test SMS" is capped independently of the monthly credit quota — 10/minute, 10/hour, 20/day per org — so verifying a channel works can\'t eat into the budget you\'re paying for.',
        'The SMS consent timestamp is always stamped server-side now, full stop — an earlier pass at this let a client supply its own value for it, which a same-day security review caught before it shipped anywhere near production.',
      ],
    },
    {
      type: 'h2',
      text: 'Follow along',
    },
    {
      type: 'p',
      text: 'Team management is next on the board — inviting a teammate into your org. Releases land on this blog as they ship; the GitHub repo has the full commit history and architecture decision records if you want the why behind any of this.',
    },
    {
      type: 'signature',
      text: '— Andrew',
    },
  ],
}
