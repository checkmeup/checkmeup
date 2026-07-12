import type { BlogPost } from '../types'

export const post: BlogPost = {
  slug: 'v1-29-release-notes',
  title: 'v1.29: Incidents Get an Expiry Date, and So Does Everything Else That Was Missing One',
  date: 'July 12, 2026',
  readTime: '4 min read',
  excerpt:
    "A follow-up to last release's incident management feature: resolved incidents now age out after 90 days, and every resource that could previously grow without limit — active incidents, incident updates, maintenance windows, API keys — now has a flat, plan-independent ceiling.",
  content: [
    {
      type: 'p',
      text: 'Manual incident declaration shipped last release without one thing every other creatable resource here already had: a limit. Monitors, status pages, and notification channels are all plan-limited; a newly-declared incident had nothing bounding it at all. This release closes that, and while auditing it, the same gap turned up in three other places that predate incidents entirely — maintenance windows and API keys never had a limit either.',
    },
    {
      type: 'h3',
      text: 'Resolved incidents age out after 90 days',
    },
    {
      type: 'p',
      text: 'Same window as check history (uptime and port checks), same daily cleanup pass, uniform across every plan — Hobby and Enterprise both get 90 days, nothing more, nothing less. The one rule that matters: only resolved incidents are ever touched. A still-active incident is exempt from this regardless of how old it gets, so a genuinely ongoing incident can never silently vanish off a status page on a timer. Only closing it starts the clock.',
    },
    {
      type: 'h3',
      text: "Retention alone doesn't stop an org that never resolves anything",
    },
    {
      type: 'p',
      text: "Time-based cleanup only works on things that eventually get marked done. An incident that's declared and never resolved sits there forever, exempt from the 90-day prune by design — so an org could still grow the incident table without bound just by never closing anything out. Declaring a new incident now checks how many the org already has open; past 100 at once, it's rejected until one gets resolved. Not a plan limit — every plan gets the same 100, and there's nothing to upgrade past it, only something to resolve.",
    },
    {
      type: 'p',
      text: "The same shape applies to updates on a single incident: up to 100 per incident, on every plan. This one mattered more than it might look — the public status page renders an active incident's entire update timeline on every visitor's page load, unauthenticated, no login. An incident with an unbounded number of updates would have meant an unbounded page load for every single visitor, not just something a paying customer's own dashboard had to deal with.",
    },
    {
      type: 'h3',
      text: 'Maintenance windows and API keys get the same treatment',
    },
    {
      type: 'p',
      text: 'Auditing incidents for this was the trigger to check whether anything else in the product had the same gap. Two did: maintenance windows and API keys could both be created without limit, on any plan, forever. Both now cap at 100 per org, same flat, plan-independent shape as the incident caps above. Delete an old maintenance window or revoke an old API key to make room for a new one.',
    },
    {
      type: 'h2',
      text: 'Also this release',
    },
    {
      type: 'ul',
      items: [
        'The private incidents list and the public status page’s active-incidents section were both unbounded queries with no result cap; both now stop at 200, matching the limit every other incident-list query in the product already used.',
        'A small spacing fix from the previous release — more breathing room above the "Past incidents" heading on the public status page.',
      ],
    },
    {
      type: 'h2',
      text: 'Follow along',
    },
    {
      type: 'p',
      text: 'Team management — inviting a teammate into your org — is next up. Releases land on this blog as they ship; the GitHub repo has the full commit history and architecture decision records if you want the reasoning behind any specific number in this post.',
    },
    {
      type: 'signature',
      text: '— Andrew',
    },
  ],
}
