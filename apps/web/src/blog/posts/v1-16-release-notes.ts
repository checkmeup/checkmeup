import type { BlogPost } from '../types'

export const post: BlogPost = {
  slug: 'v1-16-release-notes',
  title: 'v1.16: Change Plan Without Leaving the Page',
  date: 'July 2, 2026',
  readTime: '3 min read',
  excerpt:
    'Upgrading, downgrading, and cancelling an existing subscription now happens directly on the Billing page — no re-running checkout to switch tiers. Paddle rejections surface with the actual reason instead of a generic error, and a scheduled cancellation is tracked properly instead of just hoping the UI catches up.',
  content: [
    {
      type: 'p',
      text: 'v1.15 moved billing onto Paddle, but it only carried over what LemonSqueezy already did: Hobby to a first paid plan via checkout. Moving between paid tiers, or cancelling back down to Hobby, had no path of its own yet. This release adds one.',
    },
    {
      type: 'h3',
      text: 'Upgrade, downgrade, cancel — in place',
    },
    {
      type: 'p',
      text: "The Billing page now lists every plan you could move to directly, split into upgrade and downgrade options relative to your current tier, each a single click. Moving up or down between paid plans calls Paddle's subscription-update API against your existing subscription — no new checkout, no re-entering payment details. Only the very first move off Hobby still goes through the checkout overlay from v1.15, because there's no subscription yet to modify.",
    },
    {
      type: 'p',
      text: "Cancelling works the same way but takes effect at the end of the current billing period, same as before — you keep your plan's access until then rather than getting cut off mid-cycle. The difference is how the UI now knows that's happened: it used to just show a static confirmation message and hope for the best. Now it polls the billing query until the account's subscription status actually flips to `cancel_scheduled`, so the page reflects the real state — including hiding the upgrade/downgrade/cancel actions entirely once a cancellation is locked in, since Paddle rejects any further change to a subscription in that state as a conflict until it actually lapses.",
    },
    {
      type: 'h3',
      text: 'Errors that say what Paddle actually said',
    },
    {
      type: 'p',
      text: 'Paddle rejecting a plan change (wrong subscription state, a price ID that no longer exists, that kind of thing) used to come back as a flat "failed to change plan" — same message regardless of cause. It now distinguishes a 4xx from Paddle, which means the request itself was rejected for a business reason, from a genuine infrastructure failure (network error, Paddle 5xx, an unparseable response). The former surfaces as a 409 with Paddle\'s own error code and detail text attached, so the actual reason is visible without digging through server logs; the latter still falls back to a generic 500.',
    },
    {
      type: 'h2',
      text: 'Also this release',
    },
    {
      type: 'p',
      text: "The subscription-cancellation webhook handler picked up the same treatment — Paddle's `subscription.updated`/`.canceled` events now drive the `cancel_scheduled` status transition the Billing page polls for. And a Codacy pass on the new `ChangePlan` handler split its request decoding and validation into a separate function, bringing it back under the cyclomatic-complexity and length thresholds without changing behavior.",
    },
    {
      type: 'h2',
      text: 'Follow along',
    },
    {
      type: 'p',
      text: 'Microsoft Teams alerts are still next on the board. Releases land on this blog as they ship; the GitHub repo has the full commit history and architecture decision records if you want the why behind any of this.',
    },
    {
      type: 'signature',
      text: '— Andrew',
    },
  ],
}
