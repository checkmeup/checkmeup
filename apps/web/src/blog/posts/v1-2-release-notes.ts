import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "Following yesterday's v1.1, here's v1.2 — billing-focused.",
  },
  {
    type: 'h3',
    text: 'Annual billing',
  },
  {
    type: 'p',
    text: 'Every paid plan now has an annual option at roughly two months free: Solo $90/yr, Startup $290/yr, Enterprise $990/yr. Toggle between monthly and annual on the pricing page or right on the Billing page before you upgrade.',
  },
  {
    type: 'h3',
    text: 'Real upgrade buttons',
  },
  {
    type: 'p',
    text: 'The Billing page said "coming soon" since the day v1.0 launched. It now has actual upgrade buttons that take you to checkout. And if you hit a plan limit while creating a monitor or status page, you get an inline "upgrade to add more" prompt instead of a dead-end red error message.',
  },
  {
    type: 'h2',
    text: "Why you can't pay me yet",
  },
  {
    type: 'p',
    text: "All of the billing code above is done and tested. What it's waiting on is LemonSqueezy's identity verification review on my account — standard process, typically 1-2 business days, currently in progress. The store is in test mode in the meantime, so clicking \"Upgrade\" won't charge anyone real money. I'll post here the moment it's live.",
  },
  {
    type: 'h2',
    text: 'Follow along',
  },
  {
    type: 'p',
    text: 'Releases land here on the blog. The GitHub repo has full commit history and 22 architecture decision records if you want the why behind any of this.',
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
