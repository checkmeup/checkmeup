import type { BlogPost } from '../types'

export const post: BlogPost = {
  slug: 'v1-10-release-notes',
  title: 'v1.10: Slack Alerts',
  date: 'June 24, 2026',
  readTime: '3 min read',
  excerpt:
    'Slack joins Telegram, email, and webhooks as a native alert channel. Paste an Incoming Webhook URL, pick which monitors it covers, and get formatted Block Kit messages when something goes down and when it recovers.',
  content: [
    {
      type: 'p',
      text: "Alert channels have been growing steadily since v1.5: email in v1.5, generic webhooks in v1.7, Telegram from day one. This release adds Slack — the channel most teams already have open — using Slack's own Incoming Webhooks, which are free, don't need OAuth, and don't require App Store approval.",
    },
    {
      type: 'h3',
      text: 'How it works',
    },
    {
      type: 'p',
      text: "Create a Slack Incoming Webhook in your workspace (any channel, any app), copy the URL, and paste it into Settings → Notification channels. checkmeup validates the URL against the `hooks.slack.com` pattern and sends a test message so you can confirm it's wired up correctly before saving. You can add multiple Slack webhooks — one per channel, team, or severity tier — each is its own channel record.",
    },
    {
      type: 'h3',
      text: 'What the messages look like',
    },
    {
      type: 'p',
      text: "Down alerts arrive as Block Kit messages: monitor name, type, failure reason, and timestamp — readable in Slack without opening checkmeup. Recovery alerts add the downtime duration so the thread tells the whole story. Both use the same fire-and-forget delivery as the generic webhook channel: a non-2xx response (a removed or rotated webhook, for example) is logged in Settings, never blocks the check loop, and isn't retried automatically.",
    },
    {
      type: 'h3',
      text: 'Channel selection per monitor',
    },
    {
      type: 'p',
      text: "When creating or editing a monitor you choose which notification channels it uses — Slack webhooks, Telegram, email, or any mix. The per-incident alert cap introduced in v1.5 counts across all enabled channels together, so adding Slack to a monitor that already sends email doesn't double the alert volume for the same incident. Recovery alerts are always sent regardless of the cap.",
    },
    {
      type: 'h3',
      text: 'Failure handling',
    },
    {
      type: 'p',
      text: "Slack webhooks can be deleted or their URLs rotated. When a delivery fails (non-2xx or timeout), checkmeup logs the error against the channel in Settings so you can spot a broken webhook and update it. The worker's check loop is never blocked waiting on a Slack response.",
    },
    {
      type: 'h2',
      text: 'Follow along',
    },
    {
      type: 'p',
      text: 'Microsoft Teams alerts are next on the list, built on the same webhook delivery foundation. Releases land on this blog as they ship; the GitHub repo has the full commit history and architecture decision records for the notification channel model.',
    },
    {
      type: 'signature',
      text: '— Andrew',
    },
  ],
}
