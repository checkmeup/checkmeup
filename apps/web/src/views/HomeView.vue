<script setup lang="ts">
import LandingLayout from '@/layouts/LandingLayout.vue'

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
    features: ['30 monitors', '1-min checks', 'Telegram, Slack & email alerts', '3 status pages'],
  },
  {
    name: 'Startup',
    price: 29,
    description: 'For small agencies',
    highlight: true,
    features: ['100 monitors', '1-min checks', 'Telegram, Slack & email alerts', '10 status pages'],
  },
  {
    name: 'Enterprise',
    price: 99,
    description: 'For growing agencies',
    highlight: false,
    features: ['1000 monitors', '1-min checks', 'Telegram, Slack & email alerts', '100 status pages'],
  },
]
</script>

<template>
  <LandingLayout>
    <!-- Hero -->
    <section class="max-w-6xl mx-auto px-4 sm:px-6 pt-20 pb-16 sm:pt-28 sm:pb-24 text-center">
      <div
        class="inline-flex items-center gap-2 text-xs font-medium px-3 py-1 rounded-full mb-6 border"
        style="
          background-color: var(--surface);
          border-color: var(--border);
          color: var(--color-green-300);
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
        style="color: var(--text-strong)"
      >
        Know before<br class="hidden sm:block" />
        your users do.
      </h1>

      <p class="text-lg sm:text-xl max-w-2xl mx-auto mb-10" style="color: var(--text-dim)">
        Cron job monitoring, uptime checks, SSL and domain expiry alerts — all in one place. Get
        alerted on Telegram, Slack, or email the moment something breaks.
      </p>

      <div class="flex flex-col sm:flex-row items-center justify-center gap-3 mb-16">
        <RouterLink
          to="/sign-up"
          class="w-full sm:w-auto text-sm font-semibold px-6 py-3 rounded-md transition-colors"
          style="background-color: var(--color-green-500); color: var(--on-accent)"
        >
          Start free — no credit card
        </RouterLink>
        <RouterLink
          to="/sign-in"
          class="w-full sm:w-auto text-sm px-6 py-3 rounded-md border transition-colors"
          style="color: var(--text-dim); border-color: var(--border)"
        >
          Sign in →
        </RouterLink>
      </div>

      <!-- Terminal mockup -->
      <div
        class="max-w-2xl mx-auto rounded-xl border text-left overflow-hidden"
        style="background-color: var(--surface); border-color: var(--border)"
      >
        <div
          class="flex items-center gap-1.5 px-4 py-3 border-b"
          style="border-color: var(--border)"
        >
          <span class="w-3 h-3 rounded-full" style="background-color: var(--status-down); opacity: 0.6"></span>
          <span class="w-3 h-3 rounded-full" style="background-color: var(--status-degraded); opacity: 0.6"></span>
          <span class="w-3 h-3 rounded-full" style="background-color: var(--status-up); opacity: 0.6"></span>
          <span class="ml-3 text-xs" style="color: var(--text-muted)">bash</span>
        </div>
        <div
          class="px-5 py-5 font-mono text-xs sm:text-sm space-y-1.5"
          style="color: var(--text-dim)"
        >
          <div style="color: var(--text-muted)"># 1. Add a cron monitor, copy the ping URL</div>
          <div>
            <span style="color: var(--color-green-300)">$</span>
            <span class="ml-2"
              >curl https://checkmeup.net/ping/<span style="color: var(--color-green-500)"
                >abc-123</span
              ></span
            >
          </div>
          <div style="color: var(--color-green-500)">OK</div>
          <div class="pt-1" style="color: var(--text-muted)">
            # 2. Miss a ping — get alerted instantly
          </div>
          <div class="pt-1 flex items-center gap-2">
            <span
              class="inline-block w-2 h-2 rounded-full flex-shrink-0"
              style="background-color: var(--status-up)"
            ></span>
            <span
              >backup-db · <span style="color: var(--status-up)">healthy</span> · last ping 4 min
              ago</span
            >
          </div>
          <div class="flex items-center gap-2">
            <span
              class="inline-block w-2 h-2 rounded-full flex-shrink-0"
              style="background-color: var(--status-down)"
            ></span>
            <span
              >send-invoices · <span style="color: var(--status-down)">missed</span> · 3 hours
              late</span
            >
          </div>
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
          Four monitor types, one dashboard, zero blind spots.
        </p>
      </div>

      <div class="grid sm:grid-cols-2 lg:grid-cols-4 gap-6">
        <!-- Cron -->
        <div
          class="rounded-xl border p-6"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div
            class="w-10 h-10 rounded-lg flex items-center justify-center mb-4"
            style="background-color: var(--surface-raised)"
          >
            <svg
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              style="color: var(--color-green-500)"
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
              Telegram alerts
            </li>
          </ul>
        </div>

        <!-- Uptime -->
        <div
          class="rounded-xl border p-6"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div
            class="w-10 h-10 rounded-lg flex items-center justify-center mb-4"
            style="background-color: var(--surface-raised)"
          >
            <svg
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              style="color: var(--color-green-500)"
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
          class="rounded-xl border p-6"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div
            class="w-10 h-10 rounded-lg flex items-center justify-center mb-4"
            style="background-color: var(--surface-raised)"
          >
            <svg
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              style="color: var(--color-green-500)"
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
          class="rounded-xl border p-6"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div
            class="w-10 h-10 rounded-lg flex items-center justify-center mb-4"
            style="background-color: var(--surface-raised)"
          >
            <svg
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              style="color: var(--color-green-500)"
            >
              <circle cx="12" cy="12" r="10" />
              <line x1="2" y1="12" x2="22" y2="12" />
              <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" />
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
            style="background-color: var(--surface-raised); color: var(--color-green-300)"
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
            class="inline-flex items-center gap-2 text-sm font-medium px-5 py-2.5 rounded-md transition-colors"
            style="background-color: var(--color-green-500); color: var(--on-accent)"
          >
            Create a status page →
          </RouterLink>
        </div>

        <!-- Status page mockup -->
        <div
          class="w-full lg:w-80 rounded-xl border overflow-hidden flex-shrink-0"
          style="background-color: var(--bg); border-color: var(--border)"
        >
          <div
            class="px-5 py-4 border-b flex items-center gap-2"
            style="border-color: var(--border)"
          >
            <span
              class="w-2 h-2 rounded-full"
              style="background-color: var(--color-green-500)"
            ></span>
            <span class="text-xs font-semibold" style="color: var(--text-strong)"
              >status.acme.com</span
            >
          </div>
          <div class="px-5 py-5 space-y-3">
            <div
              class="text-xs font-medium px-2 py-1 rounded"
              style="background-color: var(--surface); color: var(--color-green-500)"
            >
              All systems operational
            </div>
            <div class="space-y-2 pt-1">
              <div class="flex items-center justify-between text-xs" style="color: var(--text-dim)">
                <span>API</span>
                <span style="color: var(--status-up)">Operational</span>
              </div>
              <div class="flex items-center justify-between text-xs" style="color: var(--text-dim)">
                <span>Dashboard</span>
                <span style="color: var(--status-up)">Operational</span>
              </div>
              <div class="flex items-center justify-between text-xs" style="color: var(--text-dim)">
                <span>Webhooks</span>
                <span style="color: var(--status-degraded)">Degraded</span>
              </div>
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
          class="rounded-xl border p-6 text-left relative"
          :style="{
            backgroundColor: 'var(--surface)',
            borderColor: plan.highlight ? 'var(--color-green-500)' : 'var(--border)',
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

    <!-- CTA banner -->
    <section class="max-w-6xl mx-auto px-4 sm:px-6 py-16 sm:py-20">
      <div
        class="rounded-2xl text-center px-8 py-14 border"
        style="
          background: linear-gradient(135deg, var(--cta-gradient-start) 0%, var(--cta-gradient-end) 100%);
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
          class="inline-flex items-center gap-2 text-sm font-semibold px-7 py-3 rounded-md transition-colors"
          style="background-color: var(--color-green-500); color: var(--on-accent)"
        >
          Create free account →
        </RouterLink>
      </div>
    </section>
  </LandingLayout>
</template>
