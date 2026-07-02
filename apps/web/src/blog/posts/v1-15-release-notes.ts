import type { BlogPost } from '../types'

export const post: BlogPost = {
  slug: 'v1-15-release-notes',
  title: 'v1.15: Billing Moved to Paddle',
  date: 'July 2, 2026',
  readTime: '3 min read',
  excerpt:
    "checkmeup's payment provider changed from LemonSqueezy to Paddle — full replacement, not a dual-provider setup. Nothing changes for existing subscribers except where the invoice comes from; the trigger was a payment processor's domain-verification checklist that surfaced Paddle as the better long-term fit before real revenue depended on either provider.",
  content: [
    {
      type: 'p',
      text: "LemonSqueezy has handled checkmeup's billing since launch — Merchant of Record, so it collects and remits tax globally and I never have to think about VAT in forty countries. That arrangement worked fine. The switch to Paddle wasn't a LemonSqueezy problem; it came from an unrelated direction, the same payment-processor domain-verification checklist that led to the Terms of Service and Privacy Policy overhaul in v1.14. Working through it surfaced Paddle as the stronger fit going forward, and better to switch now, before real revenue depends on either provider, than after.",
    },
    {
      type: 'h3',
      text: 'What actually changed',
    },
    {
      type: 'p',
      text: 'Same deal as before in every way that matters to a subscriber: Paddle is the legal seller on every transaction, invoices come from Paddle not checkmeup, and Paddle handles tax collection and remittance worldwide. checkmeup still operates as an Israeli sole proprietor (עוסק פטור) with no registered company, so a Merchant of Record staying in the picture at all was non-negotiable — see ADR-026, which supersedes the original LemonSqueezy decision (ADR-018).',
    },
    {
      type: 'p',
      text: "The technical shape of the integration changed more than the business terms did. LemonSqueezy's checkout was a server-generated hosted URL — the backend made a request, got a link back, and the browser navigated to it. Paddle's checkout is client-side: the backend creates a transaction and hands the frontend a transaction ID, and a `@paddle/paddle-js` overlay opens right there in the page — no navigating away from checkmeup at all. Webhooks changed shape too (different signature scheme, explicit event types instead of generic status branching), and the customer-portal link is now a per-customer session generated on demand server-side rather than one static URL for everyone.",
    },
    {
      type: 'p',
      text: "The org_id trust boundary carried over exactly as it worked before: the backend reads which org a checkout belongs to from the authenticated session, never from anything the client sends, so there's no way to tamper with a request and attribute a purchase to someone else's account.",
    },
    {
      type: 'h2',
      text: 'Also this release',
    },
    {
      type: 'p',
      text: "LemonSqueezy's integration is removed outright rather than kept around as a fallback — the DB columns, env vars, and API client are all Paddle-shaped now. `docs/billing-setup.md` has the updated Paddle dashboard checklist for anyone standing up their own fork.",
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
