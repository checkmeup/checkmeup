import type { BlogPost } from '../types'

export const post: BlogPost = {
  slug: 'v1-3-release-notes',
  title: 'v1.3: Email Alerts and Keyword Monitoring',
  date: 'June 19, 2026',
  readTime: '2 min read',
  excerpt:
    "A second alert channel alongside Telegram, and uptime monitors that can now look inside the response body. Here's what shipped in v1.3.",
  content: [
    {
      type: 'p',
      text: 'v1.3 is two features that have been on the roadmap since the MVP: email alerts and keyword monitoring.',
    },
    {
      type: 'h3',
      text: 'Email alerts',
    },
    {
      type: 'p',
      text: "Telegram has been the only alert channel until now. Email is a second one, fully independent — turn on Telegram, email, or both. Set an alert address in Settings (it defaults to your account email), send yourself a test email to confirm deliverability, then sit back. Down and recovery emails carry the same monitor name, type, reason, and timestamp as the Telegram message, and the per-incident alert cap counts across both channels together — enabling email doesn't double your notification volume for the same incident.",
    },
    {
      type: 'h3',
      text: 'Keyword monitoring',
    },
    {
      type: 'p',
      text: 'An uptime monitor returning 200 isn\'t always actually up — a maintenance page, a half-rendered error, a broken API response can all return 200. Keyword monitoring adds an optional check on the response body: require text to be present, or require it to be absent, case-sensitive or not. It runs as part of the existing check (no extra request), and a keyword failure is treated exactly like downtime — same 2-consecutive-failures rule, same alerting, same recovery handling, just with a clearer reason in the log ("keyword not found" instead of a generic failure).',
    },
    {
      type: 'p',
      text: "Keyword monitoring is gated to Solo and above — Hobby stays status-code-only. If you're on Hobby and already had a keyword set from a previous paid plan, it keeps working; you just can't change the text until you upgrade.",
    },
    {
      type: 'h2',
      text: "What's next",
    },
    {
      type: 'p',
      text: "Check the roadmap for what's queued after this. Releases land here on the blog, faster updates go out on the Telegram channel (@checkmeup), and the GitHub repo has the full commit history if you want the why behind any of it.",
    },
    {
      type: 'signature',
      text: '— Andrew',
    },
  ],
}
