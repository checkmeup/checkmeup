export type ContentBlock =
  | { type: 'p'; text: string }
  | { type: 'h2'; text: string }
  | { type: 'h3'; text: string }
  | { type: 'code'; lang: string; text: string }
  | { type: 'ul'; items: string[] }
  | { type: 'blockquote'; text: string }
  | { type: 'divider' }
  | { type: 'signature'; text: string }

export interface BlogPost {
  slug: string
  title: string
  date: string
  readTime: string
  excerpt: string
  content: ContentBlock[]
}

export const posts: BlogPost[] = [
  {
    slug: 'three-days-to-mvp',
    title: 'Three Days to MVP: How We Built checkmeup',
    date: 'June 16, 2026',
    readTime: '3 min read',
    excerpt:
      "A cron monitor, uptime checks, SSL alerts, status pages, and billing — shipped in roughly 20 hours over three days. Here's what actually happened.",
    content: [
      {
        type: 'p',
        text: "This was supposed to be a weekend project. It became three very focused days. By the end of June 16, checkmeup was live with cron monitoring, uptime checks, SSL expiry alerts, status pages, billing, and a mobile UI. Here's the honest account of how that happened.",
      },
      {
        type: 'h2',
        text: 'Day 1 — Decisions before code',
      },
      {
        type: 'p',
        text: 'The first day was mostly architecture decisions, and getting them right matters more than writing fast. The temptation with monitoring tools is to reach for a job queue — Redis, BullMQ, something with workers. We decided against it. Each monitor gets its own goroutine with a time.Ticker. No broker, no consumer, no ops overhead. Simple, and it scales to tens of thousands of monitors on a single node before becoming a problem.',
      },
      {
        type: 'p',
        text: "Database access: sqlc, not an ORM. You write SQL, sqlc generates typed Go functions. Every query is auditable, every column mismatch is caught at compile time, and there's no ORM magic hiding an N+1 somewhere. The upfront cost is writing migrations and running sqlc generate — worth it.",
      },
      {
        type: 'p',
        text: "Auth: JWT in an httpOnly cookie, refresh tokens in the database. No Authorization header. The cookie is sent automatically by the browser, JavaScript can't read it, and SameSite=Strict handles CSRF. Access tokens expire in 15 minutes; refresh tokens last 7 days and rotate on use. By the end of day 1 we had sign-up, sign-in, and the Vue frontend skeleton running.",
      },
      {
        type: 'h2',
        text: 'Day 2 — Docker ruins everything (temporarily)',
      },
      {
        type: 'p',
        text: 'The deploy target is a Hetzner CX23 — €5.99/month, ARM64. The dev machine is also ARM64. The production server is amd64. This should be easy. It was not.',
      },
      {
        type: 'p',
        text: "The first Docker build attempt used QEMU emulation for the amd64 cross-compile. It was too slow and kept timing out. We burned through five consecutive fix commits in 36 minutes: trying --frozen-lockfile, removing postinstall scripts, simplifying the frontend stage, switching from vue-tsc to direct vite build, and finally landing on --platform=${BUILDPLATFORM} for both stages — Go cross-compiles natively to amd64 via GOOS/GOARCH, Vite output is platform-independent JS. That's the one that worked.",
      },
      {
        type: 'code',
        lang: 'dockerfile',
        text: '# Both stages run natively on the build machine\nFROM --platform=${BUILDPLATFORM} golang:1.24-alpine AS go-builder\n# Go cross-compiles cleanly\nENV GOOS=linux GOARCH=amd64',
      },
      {
        type: 'p',
        text: 'Once the build was unblocked, the rest of day 2 moved fast: password reset via Resend email, the silent token refresh interceptor, cron monitors, and the first Telegram alert. That last one — seeing a "missed" notification arrive on my phone for a test monitor — was the moment this stopped being abstract.',
      },
      {
        type: 'h2',
        text: 'Day 3 — Everything else',
      },
      {
        type: 'p',
        text: 'Day 3 started at 6am with rate limiting and Telegram webhook auth (security first), then alert debounce — by default, a monitor that stays down sends at most 3 Telegram alerts per incident, then goes silent until recovery. You always get the "back up" message regardless.',
      },
      {
        type: 'p',
        text: 'The afternoon was the feature sprint: uptime monitors, SSL expiry monitors, and status pages shipped in under four hours combined. Then billing.',
      },
      {
        type: 'p',
        text: "Billing was its own decision. The options were Stripe (you own the tax liability globally), Paddle (Merchant of Record, solid but more setup), and LemonSqueezy (also MoR, simpler for MVP stage). We went with LemonSqueezy. They collect and remit tax in every jurisdiction. We receive net payouts. For a bootstrapped product with one person, that's the right call.",
      },
      {
        type: 'p',
        text: 'We also did a competitor pricing review that afternoon. Our initial prices were higher than they needed to be. After looking at healthchecks.io and UptimeRobot, we landed on $9 / $29 / $79 — competitive for the value, sustainable for a bootstrapped product.',
      },
      {
        type: 'h2',
        text: 'Launch day — June 16',
      },
      {
        type: 'p',
        text: 'The last morning was mobile UI, cleanup, and the landing page. By the time this post was published, checkmeup was live. Total time logged: ~20 hours.',
      },
      {
        type: 'h2',
        text: 'What we actually learned',
      },
      {
        type: 'ul',
        items: [
          'Boring tech is fast tech. Go + PostgreSQL + goroutines meant we spent zero time fighting the runtime. The Docker cross-platform build was the only real infrastructure headache.',
          "Write the ADR before the code. Documenting the decision first clarifies it. We have 19 ADRs in the repo — they're faster to write than they are to undo.",
          'Shipping the billing integration early is the right call. It forces you to think about plan limits, upgrade flows, and webhooks while the codebase is still small.',
          'The first Telegram alert from your own monitor hits different. Build the demo path first.',
        ],
      },
      {
        type: 'divider',
      },
      {
        type: 'p',
        text: 'The code is on GitHub. All the architecture decisions are in the ADR log. If you have questions or want to talk through any of the choices, reach out at andrew@checkmeup.net.',
      },
      {
        type: 'p',
        text: "I drank a genuinely irresponsible amount of coffee to make this timeline work, and I'd do it again. Probably with slightly less Docker.",
      },
      {
        type: 'signature',
        text: '— Andrew',
      },
    ],
  },
  {
    slug: 'v1-release-changelog',
    title: "v1.0 — First Release: What Shipped and What's Next",
    date: 'June 16, 2026',
    readTime: '2 min read',
    excerpt:
      "checkmeup v1.0 is live. Here's the complete changelog, what we deliberately left out, and a foggy look at the road ahead.",
    content: [
      {
        type: 'p',
        text: 'checkmeup v1.0 shipped on June 16, 2026. This is the changelog.',
      },
      {
        type: 'h2',
        text: "What's in v1.0",
      },
      {
        type: 'h3',
        text: 'Cron job monitoring',
      },
      {
        type: 'p',
        text: 'Add a monitor, get a ping URL, call it at the end of your job. Miss a ping within the grace period and you get a Telegram alert. Recover and you get another. Execution history is stored so you can see exactly when jobs ran and for how long.',
      },
      {
        type: 'code',
        lang: 'bash',
        text: '# Add to the end of any cron job\ncurl -s https://checkmeup.net/ping/your-monitor-id',
      },
      {
        type: 'h3',
        text: 'Uptime monitoring',
      },
      {
        type: 'p',
        text: 'HTTP checks every minute. Alert on non-200 status or timeout. Response time tracked per check. Incident history with start/end times and duration.',
      },
      {
        type: 'h3',
        text: 'SSL expiry monitoring',
      },
      {
        type: 'p',
        text: 'Daily certificate checks. Alerts at 30, 14, and 7 days before expiry. Issuer and expiry date visible in the dashboard.',
      },
      {
        type: 'h3',
        text: 'Telegram alerts',
      },
      {
        type: 'p',
        text: 'Connect your Telegram account in settings. Alerts go to you directly — no shared channels, no noise from other users. Alert frequency is capped per incident (default: 3 alerts, then silence until recovery). Recovery alerts always send.',
      },
      {
        type: 'h3',
        text: 'Status pages',
      },
      {
        type: 'p',
        text: 'Public status pages at checkmeup.net/status/your-slug. Add any of your monitors to a page. Clients can bookmark it. No subdomain required — the same domain is intentional (avoids DNS setup overhead for every client).',
      },
      {
        type: 'h3',
        text: 'Plans',
      },
      {
        type: 'ul',
        items: [
          'Hobbyist — Free — 10 monitors, 5-min checks, 1 status page',
          'Indie — $9/mo — 30 monitors, 1-min checks, 3 status pages',
          'Studio — $29/mo — 100 monitors, 1-min checks, 10 status pages',
          'Agency — $79/mo — Unlimited monitors, 1-min checks, unlimited status pages',
        ],
      },
      {
        type: 'h2',
        text: 'What we deliberately left out',
      },
      {
        type: 'p',
        text: "Email alerts, Slack integration, webhooks, API keys, and multi-user organizations are all missing from v1.0. That's intentional. Each of those is a real feature that would have added real complexity — and delays. Better to ship something that does the core well than to ship everything half-done.",
      },
      {
        type: 'p',
        text: "If one of those is blocking you from using checkmeup, email andrew@checkmeup.net and say so. That's how the roadmap gets prioritized.",
      },
      {
        type: 'h2',
        text: 'The foggy roadmap',
      },
      {
        type: 'p',
        text: 'These are the things most likely to happen next, in rough priority order. None of it is committed until it ships.',
      },
      {
        type: 'ul',
        items: [
          "Email alerts — for when you're not on Telegram",
          'Webhook alerts — POST to any URL on state change',
          'Public API — programmatic monitor creation and ping ingestion',
          'Slack integration — channel notifications for teams',
          'Multi-user organizations — invite teammates, shared monitors',
          'Scheduled maintenance windows — suppress alerts during planned downtime',
          'Check locations — run uptime checks from multiple regions',
        ],
      },
      {
        type: 'blockquote',
        text: 'The order will shift based on what users actually ask for. If you have strong feelings about any of these, reach out.',
      },
      {
        type: 'h2',
        text: 'Follow along',
      },
      {
        type: 'p',
        text: 'Releases and product updates go here on the blog. For faster updates, the Telegram channel (@checkmeup) gets announcements as they happen. The GitHub repo has the full commit history and all 19 architecture decision records if you want to see exactly how things are built.',
      },
      {
        type: 'p',
        text: "Thanks for being early. It means a lot — you're trusting a tool built by one person over a weekend-that-became-three-days to tell you when your stuff is on fire. I take that seriously.",
      },
      {
        type: 'signature',
        text: '— Andrew',
      },
    ],
  },
]

export function getPost(slug: string): BlogPost | undefined {
  return posts.find((p) => p.slug === slug)
}
