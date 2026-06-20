import type { BlogPost } from '../types'

export const post: BlogPost = {
  slug: 'v1-4-release-notes',
  title: 'v1.4: A Testing Pass (and the Bugs It Found)',
  date: 'June 19, 2026',
  readTime: '2 min read',
  excerpt:
    'No new features this time. Just a deliberate pass at backend test coverage — and the handful of real bugs, including one security issue, that it turned up along the way.',
  content: [
    {
      type: 'p',
      text: "v1.4 doesn't add anything you can click on. It's a testing pass over the Go backend — the part of checkmeup that handles auth, billing, and your monitor data — and it's worth a post because of what it found, not what it adds.",
    },
    {
      type: 'h3',
      text: 'What got tested',
    },
    {
      type: 'p',
      text: "Roughly 190 test cases went in across auth, billing, maintenance windows, cron and SSL monitors, the public ping endpoint, settings, and status pages — all running against a real Postgres instance, not mocks. That's the bulk of the API's HTTP handlers now covered, up from essentially nothing beyond the auth middleware.",
    },
    {
      type: 'h3',
      text: 'The security fix',
    },
    {
      type: 'p',
      text: "The one worth being upfront about: the endpoint that attaches a monitor to a status page checked that the monitor type and ID were well-formed, but never checked that the monitor actually belonged to your account. In theory, if you knew or guessed another account's monitor ID, you could have attached it to your own public status page and had their monitor's live status show up on it. I found this while writing tests, not from a report, and have no evidence it was ever used that way. It's fixed — status pages now verify ownership before attaching a monitor — and a test pins that behavior so it can't quietly regress.",
    },
    {
      type: 'h3',
      text: 'Other fixes along the way',
    },
    {
      type: 'ul',
      items: [
        'Sign-up now rolls back cleanly if account creation fails partway through, instead of leaving an orphaned record behind.',
        'The LemonSqueezy billing webhook is more defensive: it no longer silently drops a failed plan update, and it rejects requests instead of accepting them unsigned if the webhook secret is ever misconfigured.',
        'A cron monitor with alerts turned off now correctly clears its incident history on recovery, instead of looking permanently "down" in the data even after it checked back in.',
        'The "upgrade to add more" prompt on plan limits now shows up consistently on edit screens, not just when creating something new.',
      ],
    },
    {
      type: 'h2',
      text: 'Why blog about tests',
    },
    {
      type: 'p',
      text: "checkmeup is a one-person product handling other people's monitoring and billing data, so a pass like this is exactly the unglamorous work that keeps it trustworthy. I'd rather post about it plainly than only ever talk about new features.",
    },
    {
      type: 'h2',
      text: 'Follow along',
    },
    {
      type: 'p',
      text: 'Releases land here on the blog. The GitHub repo has the full commit history and architecture decision records if you want the why behind any of this.',
    },
    {
      type: 'signature',
      text: '— Andrew',
    },
  ],
}
