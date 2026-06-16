import type { BlogPost } from '../types'

export const post: BlogPost = {
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
}
