import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "v1.0 shipped June 16. Two days later, here's v1.1.",
  },
  {
    type: 'h3',
    text: 'Light & dark theme',
  },
  {
    type: 'p',
    text: "Checkmeup was dark-only at launch. Now there's a light theme too — toggle it from Settings or the quick switch in the app shell. It applies instantly, remembers your choice on this device, and defaults to your OS-level light/dark preference the first time you show up.",
  },
  {
    type: 'h3',
    text: 'Terms of Service & Privacy Policy',
  },
  {
    type: 'p',
    text: "Embarrassing in hindsight: a product handling emails and payment data had no published legal docs. That's fixed — /terms and /privacy are live, sign-up now requires accepting both, and if either ever changes materially you'll get a blocking re-accept screen on your next sign-in instead of being silently bound to new terms.",
  },
  {
    type: 'h3',
    text: 'FAQ page',
  },
  {
    type: 'p',
    text: 'A proper /faq page now exists, organized into Getting started, Billing & plans, Monitors & alerts, Status pages, and Privacy & security. The pricing page FAQ used to be its own disconnected copy — it now pulls from the same source as everything else.',
  },
  {
    type: 'h3',
    text: 'Suggest a feature',
  },
  {
    type: 'p',
    text: "There's still no support ticket queue — that's deliberate, not a gap. But there's now an in-app form in Settings so you don't have to switch to email or GitHub to tell me what's missing. It goes straight to my inbox, same as before, just easier to reach.",
  },
  {
    type: 'h2',
    text: "What's next",
  },
  {
    type: 'p',
    text: "v1.2 is annual billing and real upgrade buttons on the Billing page — that one's already out too, separate post.",
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
