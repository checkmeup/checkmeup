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
    features: [
      '1000 monitors',
      '1-min checks',
      'Telegram, Slack & email alerts',
      '100 status pages',
    ],
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
        Cron job monitoring, uptime checks, SSL and domain expiry alerts — all in one place. Get
        alerted on Telegram, Slack, or email the moment something breaks.
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

      <!-- Hero screenshot -->
      <div
        class="max-w-4xl mx-auto rounded-xl border overflow-hidden"
        style="
          border-color: var(--border);
          box-shadow:
            0 0 80px rgba(29, 158, 117, 0.15),
            0 40px 80px rgba(0, 0, 0, 0.4);
        "
      >
        <img
          src="/img/uptime-monitor.png"
          alt="Uptime monitor dashboard"
          class="w-full h-auto block"
        />
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
              Telegram, Slack &amp; email alerts
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
        </div>

        <!-- Status page screenshot -->
        <div
          class="w-full lg:w-80 rounded-xl border overflow-hidden flex-shrink-0"
          style="border-color: var(--border); box-shadow: 0 20px 60px rgba(0, 0, 0, 0.35)"
        >
          <img src="/img/status-page.png" alt="Status page example" class="w-full h-auto block" />
        </div>
      </div>
      <p class="text-base pt-10 mx-auto text-center" style="color: var(--text-dim)">
        Embed a live status badge anywhere — links straight to your status page
      </p>
      <div class="flex justify-center pt-4">
        <a href="https://checkmeup.net/status/checkmeup-net"
          ><img
            src="https://checkmeup.net/status/checkmeup-net/badge.svg"
            alt="When you run a successful product, your monitors run quietly"
        /></a>
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
