<script setup lang="ts">
import LandingLayout from '@/layouts/LandingLayout.vue'
import logoIcon from '@/assets/logo-icon.svg'

const statusRows = [
  { name: 'Hourly Cron Monitor', pct: '98%' },
  { name: 'Hourly Cron Monitor Ping', pct: '100%' },
  { name: 'checkmeup.net', pct: '99%' },
]

const plans = [
  {
    name: 'Hobby',
    price: 0,
    description: 'For personal projects',
    highlight: false,
    features: ['10 monitors', '5-min checks', 'Telegram, Slack & email alerts', '1 status page'],
  },
  {
    name: 'Solo',
    price: 9,
    description: 'For solo builders',
    highlight: false,
    features: [
      '30 monitors',
      '1-min checks',
      'Telegram, Slack, email & SMS alerts',
      '3 status pages',
    ],
  },
  {
    name: 'Startup',
    price: 29,
    description: 'For small agencies',
    highlight: true,
    features: [
      '100 monitors',
      '1-min checks',
      'Telegram, Slack, email & SMS alerts',
      '10 status pages',
    ],
  },
  {
    name: 'Enterprise',
    price: 99,
    description: 'For growing agencies',
    highlight: false,
    features: [
      '1000 monitors',
      '1-min checks',
      'Telegram, Slack, email & SMS alerts',
      '100 status pages',
    ],
  },
]

const customers = [
  {
    name: 'Alex M.',
    role: 'DevOps Engineer',
    avatar: 'img/customer1.png',
  },
  {
    name: 'Priya S.',
    role: 'Full-Stack Dev',
    avatar: 'img/customer2.png',
  },
  {
    name: 'James K.',
    role: 'Agency CTO',
    avatar: 'img/customer3.png',
  },
  {
    name: 'Layla H.',
    role: 'Startup Founder',
    avatar: 'img/customer4.png',
  },
  {
    name: 'Dan W.',
    role: 'Backend Engineer',
    avatar: 'img/customer5.png',
  },
  {
    name: 'Zoe C.',
    role: 'Platform Engineer',
    avatar: 'img/customer6.png',
  },
]

const testimonials = [
  {
    name: 'Sarah K.',
    role: 'Backend Engineer',
    avatar: 'img/avatar1.png',
    quote:
      'My nightly backup job silently failed for two weeks before I set up checkmeup. Now I get a Telegram ping within minutes if anything misses its window.',
  },
  {
    name: 'Marcus T.',
    role: 'Agency Owner',
    avatar: 'img/avatar2.png',
    quote:
      "We manage 40+ client sites. Having a white-label status page for each one changed how our clients perceive us. They feel taken care of, and we don't get 3am calls.",
  },
  {
    name: 'Gracy R.',
    role: 'Indie Developer',
    avatar: 'img/avatar3.png',
    quote:
      "Woke up to a Slack alert that my SSL cert was expiring in 7 days. I'd completely forgotten about it. Renewed it in 10 minutes. Crisis averted.",
  },
]
</script>

<template>
  <LandingLayout>
    <!-- Hero -->
    <section
      class="relative max-w-6xl mx-auto px-4 sm:px-6 pt-20 pb-16 sm:pt-28 sm:pb-24 text-center overflow-hidden"
    >
      <!-- Background orbs -->
      <div
        class="absolute pointer-events-none"
        style="
          top: -300px;
          left: 50%;
          transform: translateX(-50%);
          width: 900px;
          height: 900px;
          background: radial-gradient(circle, rgba(29, 158, 117, 0.14) 0%, transparent 65%);
          filter: blur(60px);
          z-index: -1;
        "
      ></div>
      <div
        class="absolute pointer-events-none"
        style="
          top: 100px;
          right: -100px;
          width: 500px;
          height: 500px;
          background: radial-gradient(circle, rgba(77, 201, 160, 0.07) 0%, transparent 70%);
          filter: blur(50px);
          z-index: -1;
        "
      ></div>

      <!-- Badge -->
      <div
        class="inline-flex items-center gap-2 text-xs font-medium px-3 py-1.5 rounded-full mb-8"
        style="
          background: rgba(29, 158, 117, 0.12);
          border: 1px solid rgba(29, 158, 117, 0.35);
          color: var(--color-green-300);
          backdrop-filter: blur(8px);
        "
      >
        <span
          class="w-1.5 h-1.5 rounded-full"
          style="background-color: var(--color-green-500)"
        ></span>
        Monitoring for developers and agencies
      </div>

      <h1
        class="text-4xl sm:text-5xl lg:text-6xl font-bold tracking-tight mb-6 leading-tight"
        style="
          background: linear-gradient(135deg, var(--text-strong) 30%, var(--color-green-300) 100%);
          -webkit-background-clip: text;
          -webkit-text-fill-color: transparent;
          background-clip: text;
        "
      >
        Know before<br class="hidden sm:block" />
        your users do.
      </h1>

      <p class="text-lg sm:text-xl max-w-2xl mx-auto mb-10" style="color: var(--text-dim)">
        Cron job monitoring, uptime checks, SSL and domain expiry alerts, and raw TCP port checks —
        all in one place. Get alerted on Telegram, Slack, or email the moment something breaks.
      </p>

      <div class="flex flex-col sm:flex-row items-center justify-center gap-3 mb-16">
        <RouterLink
          to="/sign-up"
          class="w-full sm:w-auto text-sm font-semibold px-6 py-3 rounded-md transition-all bg-[var(--color-green-500)] text-white hover:bg-[var(--color-green-700)]"
          style="box-shadow: 0 0 28px rgba(29, 158, 117, 0.45)"
        >
          Start free — no credit card
        </RouterLink>
        <RouterLink
          to="/sign-in"
          class="w-full sm:w-auto text-sm px-6 py-3 rounded-md border transition-colors hover:bg-[var(--surface-raised)]"
          style="color: var(--text-dim); border-color: var(--border)"
        >
          Sign in →
        </RouterLink>
      </div>

      <!-- Hero mockup -->
      <div
        class="max-w-4xl mx-auto rounded-xl border overflow-hidden text-left"
        style="
          border-color: var(--border);
          box-shadow:
            0 0 80px rgba(29, 158, 117, 0.15),
            0 40px 80px rgba(0, 0, 0, 0.4);
        "
      >
        <!-- Fake browser chrome -->
        <div
          class="flex items-center gap-1.5 px-4 py-3 border-b"
          style="border-color: var(--border); background-color: var(--surface-raised)"
        >
          <span class="w-2.5 h-2.5 rounded-full" style="background-color: #ff5f57"></span>
          <span class="w-2.5 h-2.5 rounded-full" style="background-color: #febc2e"></span>
          <span class="w-2.5 h-2.5 rounded-full" style="background-color: #28c840"></span>
          <span
            class="flex-1 text-center font-mono text-xs"
            style="color: var(--text-muted); margin-right: 76px"
          >
            checkmeup.net — Monitor
          </span>
        </div>

        <div class="p-5 sm:p-6" style="background-color: var(--surface)">
          <div class="flex items-center gap-3 mb-4">
            <span class="font-mono text-xs" style="color: var(--text-muted)">← Back</span>
            <span class="text-base font-semibold" style="color: var(--text-strong)"
              >checkmeup.net</span
            >
          </div>

          <div
            class="flex items-start justify-between pb-4 mb-4 border-b"
            style="border-color: var(--border)"
          >
            <div>
              <div class="flex items-center gap-2 mb-1">
                <span class="relative inline-block w-2.5 h-2.5">
                  <span
                    class="absolute inset-0 rounded-full animate-ping"
                    style="background-color: var(--status-up); opacity: 0.55"
                  ></span>
                  <span
                    class="absolute inset-0 rounded-full"
                    style="background-color: var(--status-up)"
                  ></span>
                </span>
                <span class="text-sm font-semibold" style="color: var(--status-up)">Up</span>
              </div>
              <div class="font-mono text-xs" style="color: var(--text-muted)">
                https://checkmeup.net
              </div>
              <div class="font-mono text-xs mt-0.5" style="color: var(--text-muted)">
                Every 10 min · Last checked 7m ago
              </div>
            </div>
            <div class="flex gap-2">
              <button
                class="px-2.5 py-1.5 rounded-md border text-xs"
                style="
                  border-color: var(--border);
                  background-color: var(--surface-raised);
                  color: var(--text-dim);
                "
              >
                Edit
              </button>
              <button
                class="px-2.5 py-1.5 rounded-md border text-xs"
                style="
                  border-color: var(--border);
                  background-color: var(--surface-raised);
                  color: var(--text-dim);
                "
              >
                Pause
              </button>
              <button
                class="px-2.5 py-1.5 rounded-md border text-xs"
                style="
                  border-color: rgba(239, 68, 68, 0.3);
                  background-color: rgba(239, 68, 68, 0.1);
                  color: var(--status-down);
                "
              >
                Delete
              </button>
            </div>
          </div>

          <div class="grid grid-cols-3 mb-4">
            <div class="text-center py-3 border-r" style="border-color: var(--border)">
              <div class="font-mono text-xl font-medium" style="color: var(--text-strong)">
                100.00%
              </div>
              <div class="text-xs mt-1" style="color: var(--text-muted)">Uptime 24h</div>
            </div>
            <div class="text-center py-3 border-r" style="border-color: var(--border)">
              <div class="font-mono text-xl font-medium" style="color: var(--text-strong)">
                100.00%
              </div>
              <div class="text-xs mt-1" style="color: var(--text-muted)">Uptime 7d</div>
            </div>
            <div class="text-center py-3">
              <div class="font-mono text-xl font-medium" style="color: var(--text-strong)">
                99.91%
              </div>
              <div class="text-xs mt-1" style="color: var(--text-muted)">Uptime 30d</div>
            </div>
          </div>

          <div class="pt-4 border-t" style="border-color: var(--border)">
            <div class="text-xs font-medium mb-2.5" style="color: var(--text-dim)">
              Response time — last 24h
            </div>
            <svg viewBox="0 0 640 48" fill="none" class="w-full h-12 block">
              <path
                d="M0 34 C10 32 18 27 32 30 C46 33 54 24 68 27 C82 30 92 37 106 34 C120 31 128 22 142 25 C156 28 166 34 180 31 C194 28 202 20 216 23 C230 26 240 33 254 29 C268 25 278 17 292 21 C306 25 316 33 330 29 C344 25 352 17 366 21 C380 25 392 33 406 29 C420 25 428 17 442 21 C456 25 468 33 482 29 C496 25 504 17 518 21 C532 25 544 33 558 29 C572 25 580 17 594 21 C608 25 620 33 634 29 C638 27 640 26 640 27"
                style="stroke: var(--status-up); stroke-width: 1.5px"
                fill="none"
              />
            </svg>
            <div class="font-mono text-xs mt-1.5" style="color: var(--text-muted)">
              Red dots = failed checks. Y axis = response time (max 77ms).
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Our Customers -->
    <section class="border-y" style="border-color: var(--border)">
      <div class="max-w-6xl mx-auto px-4 sm:px-6 py-12">
        <p
          class="text-center text-xs font-semibold uppercase tracking-widest mb-8"
          style="color: var(--text-muted)"
        >
          Trusted by developers worldwide
        </p>

        <div class="flex flex-col sm:flex-row items-center justify-center gap-6 sm:gap-10">
          <!-- Overlapping avatars -->
          <div class="flex -space-x-3">
            <img
              v-for="c in customers"
              :key="c.name"
              :src="c.avatar"
              :alt="c.name"
              :title="`${c.name} · ${c.role}`"
              class="w-10 h-10 rounded-full object-cover"
              style="box-shadow: 0 0 0 2px var(--bg)"
            />
          </div>

          <!-- Count + label -->
          <div class="text-center sm:text-left">
            <span class="text-2xl font-bold" style="color: var(--text-strong)">200+</span>
            <span class="text-sm ml-1.5" style="color: var(--text-dim)"
              >developers monitoring with checkmeup</span
            >
          </div>
        </div>

        <!-- Customer type pills -->
        <div class="flex flex-wrap justify-center gap-2 mt-8">
          <span
            v-for="tag in [
              'Solo developers',
              'Web agencies',
              'DevOps teams',
              'SaaS startups',
              'Freelancers',
            ]"
            :key="tag"
            class="text-xs px-3 py-1 rounded-full border"
            style="
              background-color: var(--surface);
              border-color: var(--border);
              color: var(--text-dim);
            "
          >
            {{ tag }}
          </span>
        </div>
      </div>
    </section>

    <!-- Features -->
    <section class="max-w-6xl mx-auto px-4 sm:px-6 py-16 sm:py-24">
      <div class="text-center mb-12">
        <h2 class="text-2xl sm:text-3xl font-bold mb-3" style="color: var(--text-strong)">
          Everything you need to stay on top
        </h2>
        <p class="text-base" style="color: var(--text-dim)">
          Five monitor types, one dashboard, zero blind spots.
        </p>
      </div>

      <div class="grid sm:grid-cols-2 lg:grid-cols-3 gap-6">
        <!-- Cron -->
        <div
          class="rounded-xl border p-6 transition-colors hover:border-[var(--color-green-700)]"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div
            class="w-10 h-10 rounded-lg flex items-center justify-center mb-4"
            style="
              background: linear-gradient(
                135deg,
                var(--color-green-700) 0%,
                var(--color-green-500) 100%
              );
            "
          >
            <svg
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="white"
              stroke-width="2"
            >
              <circle cx="12" cy="12" r="10" />
              <polyline points="12 6 12 12 16 14" />
            </svg>
          </div>
          <h3 class="font-semibold text-base mb-2" style="color: var(--text-strong)">
            Cron job monitoring
          </h3>
          <p class="text-sm leading-relaxed" style="color: var(--text-dim)">
            Ping a URL after each run. Miss a ping and we alert you immediately — before your
            backups rot or invoices stop sending.
          </p>
          <ul class="mt-4 space-y-2">
            <li class="flex items-center gap-2 text-xs" style="color: var(--text-muted)">
              <svg
                width="12"
                height="12"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                style="color: var(--color-green-500)"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              Custom grace periods
            </li>
            <li class="flex items-center gap-2 text-xs" style="color: var(--text-muted)">
              <svg
                width="12"
                height="12"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                style="color: var(--color-green-500)"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              Execution history &amp; logs
            </li>
            <li class="flex items-center gap-2 text-xs" style="color: var(--text-muted)">
              <svg
                width="12"
                height="12"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                style="color: var(--color-green-500)"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              Telegram, Slack, email &amp; SMS alerts
            </li>
          </ul>
        </div>

        <!-- Uptime -->
        <div
          class="rounded-xl border p-6 transition-colors hover:border-[var(--color-green-700)]"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div
            class="w-10 h-10 rounded-lg flex items-center justify-center mb-4"
            style="
              background: linear-gradient(
                135deg,
                var(--color-green-700) 0%,
                var(--color-green-500) 100%
              );
            "
          >
            <svg
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="white"
              stroke-width="2"
            >
              <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
            </svg>
          </div>
          <h3 class="font-semibold text-base mb-2" style="color: var(--text-strong)">
            Uptime monitoring
          </h3>
          <p class="text-sm leading-relaxed" style="color: var(--text-dim)">
            We poll your URLs as often as every 1 minute and alert you the moment your site goes
            down or returns an unexpected status code.
          </p>
          <ul class="mt-4 space-y-2">
            <li class="flex items-center gap-2 text-xs" style="color: var(--text-muted)">
              <svg
                width="12"
                height="12"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                style="color: var(--color-green-500)"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              1-minute check intervals
            </li>
            <li class="flex items-center gap-2 text-xs" style="color: var(--text-muted)">
              <svg
                width="12"
                height="12"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                style="color: var(--color-green-500)"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              Response time tracking
            </li>
            <li class="flex items-center gap-2 text-xs" style="color: var(--text-muted)">
              <svg
                width="12"
                height="12"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                style="color: var(--color-green-500)"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              Incident history
            </li>
          </ul>
        </div>

        <!-- SSL -->
        <div
          class="rounded-xl border p-6 transition-colors hover:border-[var(--color-green-700)]"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div
            class="w-10 h-10 rounded-lg flex items-center justify-center mb-4"
            style="
              background: linear-gradient(
                135deg,
                var(--color-green-700) 0%,
                var(--color-green-500) 100%
              );
            "
          >
            <svg
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="white"
              stroke-width="2"
            >
              <rect x="3" y="11" width="18" height="11" rx="2" />
              <path d="M7 11V7a5 5 0 0 1 10 0v4" />
            </svg>
          </div>
          <h3 class="font-semibold text-base mb-2" style="color: var(--text-strong)">
            SSL expiry monitoring
          </h3>
          <p class="text-sm leading-relaxed" style="color: var(--text-dim)">
            Get ahead of certificate expiry with early warnings at 30, 14, and 7 days. Never let a
            forgotten cert take your site offline.
          </p>
          <ul class="mt-4 space-y-2">
            <li class="flex items-center gap-2 text-xs" style="color: var(--text-muted)">
              <svg
                width="12"
                height="12"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                style="color: var(--color-green-500)"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              Multi-threshold alerts
            </li>
            <li class="flex items-center gap-2 text-xs" style="color: var(--text-muted)">
              <svg
                width="12"
                height="12"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                style="color: var(--color-green-500)"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              Issuer &amp; expiry details
            </li>
            <li class="flex items-center gap-2 text-xs" style="color: var(--text-muted)">
              <svg
                width="12"
                height="12"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                style="color: var(--color-green-500)"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              Daily checks
            </li>
          </ul>
        </div>

        <!-- Domain -->
        <div
          class="rounded-xl border p-6 transition-colors hover:border-[var(--color-green-700)]"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div
            class="w-10 h-10 rounded-lg flex items-center justify-center mb-4"
            style="
              background: linear-gradient(
                135deg,
                var(--color-green-700) 0%,
                var(--color-green-500) 100%
              );
            "
          >
            <svg
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="white"
              stroke-width="2"
            >
              <circle cx="12" cy="12" r="10" />
              <line x1="2" y1="12" x2="22" y2="12" />
              <path
                d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"
              />
            </svg>
          </div>
          <h3 class="font-semibold text-base mb-2" style="color: var(--text-strong)">
            Domain expiry monitoring
          </h3>
          <p class="text-sm leading-relaxed" style="color: var(--text-dim)">
            Track your domain's registration separately from its SSL certificate — rarer to lapse,
            but far more catastrophic when it does.
          </p>
          <ul class="mt-4 space-y-2">
            <li class="flex items-center gap-2 text-xs" style="color: var(--text-muted)">
              <svg
                width="12"
                height="12"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                style="color: var(--color-green-500)"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              Multi-threshold alerts
            </li>
            <li class="flex items-center gap-2 text-xs" style="color: var(--text-muted)">
              <svg
                width="12"
                height="12"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                style="color: var(--color-green-500)"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              Registrar &amp; expiry details
            </li>
            <li class="flex items-center gap-2 text-xs" style="color: var(--text-muted)">
              <svg
                width="12"
                height="12"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                style="color: var(--color-green-500)"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              Daily checks
            </li>
          </ul>
        </div>

        <!-- Port -->
        <div
          class="rounded-xl border p-6 transition-colors hover:border-[var(--color-green-700)]"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div
            class="w-10 h-10 rounded-lg flex items-center justify-center mb-4"
            style="
              background: linear-gradient(
                135deg,
                var(--color-green-700) 0%,
                var(--color-green-500) 100%
              );
            "
          >
            <svg
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="white"
              stroke-width="2"
            >
              <rect x="2" y="9" width="6" height="6" rx="1" />
              <rect x="16" y="9" width="6" height="6" rx="1" />
              <line x1="8" y1="12" x2="16" y2="12" />
            </svg>
          </div>
          <h3 class="font-semibold text-base mb-2" style="color: var(--text-strong)">
            Port (TCP) monitoring
          </h3>
          <p class="text-sm leading-relaxed" style="color: var(--text-dim)">
            Raw TCP connect checks for mail servers, databases, and any other non-HTTP service —
            plus a security mode that alerts if a port that should be closed becomes reachable.
          </p>
          <ul class="mt-4 space-y-2">
            <li class="flex items-center gap-2 text-xs" style="color: var(--text-muted)">
              <svg
                width="12"
                height="12"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                style="color: var(--color-green-500)"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              Any host and port
            </li>
            <li class="flex items-center gap-2 text-xs" style="color: var(--text-muted)">
              <svg
                width="12"
                height="12"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                style="color: var(--color-green-500)"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              Open or closed expected state
            </li>
            <li class="flex items-center gap-2 text-xs" style="color: var(--text-muted)">
              <svg
                width="12"
                height="12"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                style="color: var(--color-green-500)"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              Connect-time history
            </li>
          </ul>
        </div>

        <!-- Status pages -->
        <div
          class="rounded-xl border p-6 transition-colors hover:border-[var(--color-green-700)]"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div
            class="w-10 h-10 rounded-lg flex items-center justify-center mb-4"
            style="
              background: linear-gradient(
                135deg,
                var(--color-green-700) 0%,
                var(--color-green-500) 100%
              );
            "
          >
            <svg
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="white"
              stroke-width="2"
            >
              <rect x="3" y="4" width="18" height="14" rx="2" />
              <line x1="3" y1="8" x2="21" y2="8" />
              <circle cx="17" cy="15" r="2" />
            </svg>
          </div>
          <h3 class="font-semibold text-base mb-2" style="color: var(--text-strong)">
            Public status pages
          </h3>
          <p class="text-sm leading-relaxed" style="color: var(--text-dim)">
            White-label status pages your users or clients can bookmark — uptime, incidents, and
            maintenance windows, no login required to view.
          </p>
          <ul class="mt-4 space-y-2">
            <li class="flex items-center gap-2 text-xs" style="color: var(--text-muted)">
              <svg
                width="12"
                height="12"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                style="color: var(--color-green-500)"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              White-label branding
            </li>
            <li class="flex items-center gap-2 text-xs" style="color: var(--text-muted)">
              <svg
                width="12"
                height="12"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                style="color: var(--color-green-500)"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              No DNS setup required
            </li>
            <li class="flex items-center gap-2 text-xs" style="color: var(--text-muted)">
              <svg
                width="12"
                height="12"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                style="color: var(--color-green-500)"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              Embeddable status badges
            </li>
          </ul>
        </div>
      </div>
    </section>

    <!-- Status pages highlight -->
    <section class="max-w-6xl mx-auto px-4 sm:px-6 py-16 sm:py-20">
      <div
        class="rounded-2xl border px-8 py-12 sm:px-12 sm:py-16 flex flex-col lg:flex-row items-start lg:items-center gap-10"
        style="background-color: var(--surface); border-color: var(--border)"
      >
        <div class="flex-1">
          <div
            class="inline-flex items-center gap-1.5 text-xs font-medium px-2.5 py-1 rounded-full mb-4"
            style="
              background: rgba(29, 158, 117, 0.12);
              border: 1px solid rgba(29, 158, 117, 0.3);
              color: var(--color-green-300);
            "
          >
            <svg
              width="10"
              height="10"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
            >
              <circle cx="12" cy="12" r="10" />
              <line x1="2" y1="12" x2="22" y2="12" />
              <path
                d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"
              />
            </svg>
            Status pages
          </div>
          <h2 class="text-2xl sm:text-3xl font-bold mb-4" style="color: var(--text-strong)">
            Give your clients a page they can bookmark.
          </h2>
          <p class="text-base leading-relaxed mb-6" style="color: var(--text-dim)">
            Create white-label status pages for every client. They see uptime, incidents, and
            maintenance windows — you stay in control.
          </p>
          <RouterLink
            to="/sign-up"
            class="inline-flex items-center gap-2 text-sm font-medium px-5 py-2.5 rounded-md transition-all bg-[var(--color-green-500)] text-white hover:bg-[var(--color-green-700)]"
            style="box-shadow: 0 0 20px rgba(29, 158, 117, 0.35)"
          >
            Create a status page →
          </RouterLink>

          <div
            class="mt-6 p-4 rounded-lg border"
            style="border-color: var(--border); background-color: var(--surface-raised)"
          >
            <p class="text-sm mb-2.5" style="color: var(--text-dim)">
              Embed a live status badge anywhere — links straight to your status page
            </p>
            <a href="https://checkmeup.net/status/checkmeup-net">
              <img
                src="https://checkmeup.net/status/checkmeup-net/badge.svg"
                alt="When you run a successful product, your monitors run quietly"
              />
            </a>
          </div>
        </div>

        <!-- Status page mockup -->
        <div
          class="w-full lg:w-80 rounded-xl border overflow-hidden flex-shrink-0 text-sm"
          style="border-color: var(--border); box-shadow: 0 20px 60px rgba(0, 0, 0, 0.35)"
        >
          <div
            class="flex items-center gap-2 px-4 py-3 border-b"
            style="border-color: var(--border); background-color: var(--surface)"
          >
            <img :src="logoIcon" alt="" class="w-4 h-4" />
            <span class="font-semibold" style="color: var(--text-strong)">checkmeup</span>
          </div>
          <div
            class="px-4 py-2.5 border-b"
            style="border-color: var(--border); background-color: rgba(29, 158, 117, 0.08)"
          >
            <span class="text-xs font-semibold" style="color: var(--color-green-500)"
              >● All systems operational</span
            >
          </div>
          <div
            v-for="row in statusRows"
            :key="row.name"
            class="px-4 py-3 border-b last:border-0"
            style="border-color: var(--border); background-color: var(--surface)"
          >
            <div class="flex items-center justify-between mb-1.5">
              <span style="color: var(--text-dim)">{{ row.name }}</span>
              <span
                class="text-xs font-semibold px-2 py-0.5 rounded-full"
                style="background-color: rgba(29, 158, 117, 0.1); color: var(--color-green-500)"
              >
                Operational
              </span>
            </div>
            <div
              class="h-1.5 rounded overflow-hidden"
              style="background-color: rgba(29, 158, 117, 0.1)"
            >
              <div
                class="h-full rounded"
                :style="{ width: row.pct, backgroundColor: 'var(--color-green-500)' }"
              ></div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Pricing teaser -->
    <section class="max-w-6xl mx-auto px-4 sm:px-6 py-16 sm:py-20 text-center">
      <h2 class="text-2xl sm:text-3xl font-bold mb-3" style="color: var(--text-strong)">
        Simple pricing
      </h2>
      <p class="mb-10" style="color: var(--text-dim)">Start free. Scale as you grow.</p>

      <div class="grid sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div
          v-for="plan in plans"
          :key="plan.name"
          class="rounded-xl border p-6 text-left relative transition-colors"
          :style="{
            backgroundColor: 'var(--surface)',
            borderColor: plan.highlight ? 'var(--color-green-500)' : 'var(--border)',
            boxShadow: plan.highlight ? '0 0 30px rgba(29, 158, 117, 0.2)' : 'none',
          }"
        >
          <div
            v-if="plan.highlight"
            class="absolute -top-3 left-1/2 -translate-x-1/2 text-xs font-semibold px-3 py-0.5 rounded-full"
            style="background-color: var(--color-green-500); color: var(--on-accent)"
          >
            Popular
          </div>
          <div class="text-sm font-semibold mb-1" style="color: var(--text-strong)">
            {{ plan.name }}
          </div>
          <div class="text-2xl font-bold mb-1" style="color: var(--text-strong)">
            {{ plan.price === 0 ? 'Free' : `$${plan.price}` }}
            <span v-if="plan.price > 0" class="text-sm font-normal" style="color: var(--text-muted)"
              >/mo</span
            >
          </div>
          <p class="text-xs mb-4" style="color: var(--text-muted)">{{ plan.description }}</p>
          <ul class="space-y-1.5">
            <li
              v-for="f in plan.features"
              :key="f"
              class="flex items-center gap-2 text-xs"
              style="color: var(--text-dim)"
            >
              <svg
                width="11"
                height="11"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                style="color: var(--color-green-500)"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              {{ f }}
            </li>
          </ul>
        </div>
      </div>

      <RouterLink
        to="/pricing"
        class="inline-flex items-center gap-1 mt-8 text-sm transition-colors"
        style="color: var(--color-green-500)"
      >
        See full pricing details →
      </RouterLink>
    </section>

    <!-- Testimonials -->
    <section class="max-w-6xl mx-auto px-4 sm:px-6 py-16 sm:py-20">
      <div class="text-center mb-12">
        <h2 class="text-2xl sm:text-3xl font-bold mb-3" style="color: var(--text-strong)">
          Loved by developers
        </h2>
        <p class="text-base" style="color: var(--text-dim)">
          Monitoring that gets out of your way.
        </p>
      </div>

      <div class="grid sm:grid-cols-3 gap-6">
        <div
          v-for="t in testimonials"
          :key="t.name"
          class="rounded-xl border p-6 flex flex-col gap-4"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div class="flex items-center gap-3">
            <img
              :src="t.avatar"
              :alt="t.name"
              class="w-10 h-10 rounded-full object-cover flex-shrink-0"
            />
            <div>
              <div class="text-sm font-semibold" style="color: var(--text-strong)">
                {{ t.name }}
              </div>
              <div class="text-xs" style="color: var(--text-muted)">{{ t.role }}</div>
            </div>
          </div>
          <!-- Stars -->
          <div class="flex gap-0.5">
            <svg
              v-for="i in 5"
              :key="i"
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="currentColor"
              style="color: var(--color-green-500)"
            >
              <polygon
                points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"
              />
            </svg>
          </div>
          <p class="text-sm leading-relaxed" style="color: var(--text-dim)">"{{ t.quote }}"</p>
        </div>
      </div>
    </section>

    <!-- CTA banner -->
    <section class="max-w-6xl mx-auto px-4 sm:px-6 py-16 sm:py-20">
      <div
        class="rounded-2xl text-center px-8 py-14 border"
        style="
          background: linear-gradient(
            135deg,
            var(--cta-gradient-start) 0%,
            var(--cta-gradient-end) 100%
          );
          border-color: var(--cta-border);
        "
      >
        <h2 class="text-2xl sm:text-3xl font-bold mb-3" style="color: var(--cta-text)">
          Start monitoring in 60 seconds.
        </h2>
        <p class="mb-8 text-base" style="color: var(--cta-text-dim)">
          Free plan included. No credit card required.
        </p>
        <RouterLink
          to="/sign-up"
          class="inline-flex items-center gap-2 text-sm font-semibold px-7 py-3 rounded-md transition-all bg-[var(--color-green-500)] text-white hover:bg-[var(--color-green-700)]"
          style="box-shadow: 0 0 28px rgba(29, 158, 117, 0.5)"
        >
          Create free account →
        </RouterLink>
      </div>
    </section>
  </LandingLayout>
</template>
