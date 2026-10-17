<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import LandingLayout from '@/layouts/LandingLayout.vue'
import CtaBanner from '@/components/CtaBanner.vue'
import { findFaqCategory } from '@/faq/faqs'
import { useSeo } from '@/composables/useSeo'

useSeo({
  title: 'Pricing — Checkmeup',
  description:
    'Hobby $0, Solo $9, Startup $29, Enterprise $99 — cron, uptime, SSL, domain, port, and DNS monitoring plans for freelancers and agencies.',
  path: '/pricing',
})

const billingFaqs = findFaqCategory('billing')?.entries ?? []

// Annual price is always exactly 10x the monthly price — ~2 months free (PRD annual-billing).
const cycle = ref<'monthly' | 'annual'>('monthly')

interface Plan {
  name: string
  price: number
  annualPrice: number
  description: string
  highlight: boolean
  cta: string
  monitors: string
  statusPages: string
  checkInterval: string
  notificationChannels: string
  smsCredits: string
}

const plans: Plan[] = [
  {
    name: 'Hobby',
    price: 0,
    annualPrice: 0,
    description: 'Your own side projects',
    highlight: false,
    cta: 'Get started free',
    monitors: '10',
    statusPages: '1',
    checkInterval: '5 min',
    notificationChannels: '5',
    smsCredits: '0',
  },
  {
    name: 'Solo',
    price: 9,
    annualPrice: 90,
    description: 'A freelancer with a few clients',
    highlight: false,
    cta: 'Start Solo',
    monitors: '30',
    statusPages: '3',
    checkInterval: '1 min',
    notificationChannels: '20',
    smsCredits: '10',
  },
  {
    name: 'Startup',
    price: 29,
    annualPrice: 290,
    description: 'A small agency',
    highlight: true,
    cta: 'Start Startup',
    monitors: '100',
    statusPages: '10',
    checkInterval: '1 min',
    notificationChannels: '50',
    smsCredits: '30',
  },
  {
    name: 'Enterprise',
    price: 99,
    annualPrice: 990,
    description: 'An agency with a page per client',
    highlight: false,
    cta: 'Start Enterprise',
    monitors: '1000',
    statusPages: '100',
    checkInterval: '1 min',
    notificationChannels: '100',
    smsCredits: '100',
  },
]

function effectiveMonthly(plan: Plan): string {
  return (plan.annualPrice / 12).toFixed(2).replace(/\.00$/, '')
}

interface TableRow {
  label: string
  values: string[]
}

const featureRows: TableRow[] = [
  { label: 'Monitors (all types combined)', values: ['10', '30', '100', '1000'] },
  { label: 'Uptime check interval', values: ['5 min', '1 min', '1 min', '1 min'] },
  { label: 'Status pages', values: ['1', '3', '10', '100'] },
  { label: 'Notification channels', values: ['5', '20', '50', '100'] },
  { label: 'SMS credits / month', values: ['—', '10', '30', '100'] },
  { label: 'Cron job monitoring', values: ['✓', '✓', '✓', '✓'] },
  { label: 'Uptime monitoring', values: ['✓', '✓', '✓', '✓'] },
  { label: 'Keyword monitoring', values: ['✓', '✓', '✓', '✓'] },
  { label: 'SSL expiry monitoring', values: ['✓', '✓', '✓', '✓'] },
  { label: 'Domain expiry monitoring', values: ['✓', '✓', '✓', '✓'] },
  { label: 'Port (TCP) monitoring', values: ['✓', '✓', '✓', '✓'] },
  { label: 'DNS record monitoring', values: ['✓', '✓', '✓', '✓'] },
  { label: 'Execution history & logs', values: ['✓', '✓', '✓', '✓'] },
  { label: 'Alerts (Telegram, email, Slack, webhook)', values: ['✓', '✓', '✓', '✓'] },
  { label: 'Maintenance windows', values: ['✓', '✓', '✓', '✓'] },
  { label: 'Projects (group monitors by client)', values: ['✓', '✓', '✓', '✓'] },
  { label: 'Incident management', values: ['✓', '✓', '✓', '✓'] },
  { label: 'White-label status pages', values: ['—', '✓', '✓', '✓'] },
]
</script>

<template>
  <LandingLayout>
    <!-- Hero -->
    <section
      class="mx-auto flex max-w-[1200px] flex-col items-start gap-6 px-4 pt-16 pb-12 sm:px-6 sm:pt-24"
    >
      <h1
        class="text-[44px] leading-[1.02] font-semibold tracking-[-0.045em] text-[var(--text-strong)] sm:text-[60px]"
      >
        Priced by monitors, not seats.
      </h1>
      <p class="max-w-xl text-[17px] leading-relaxed text-[var(--text-dim)] sm:text-lg">
        Hobby is free for 10 monitors with 5-minute checks, no card. Paid plans check every minute
        and add status pages, channels and SMS — billed monthly or yearly.
      </p>

      <div class="inline-flex gap-0.5 rounded-[10px] bg-[var(--surface-raised)] p-[3px]">
        <button
          type="button"
          class="min-h-[34px] rounded-lg px-4 text-sm font-medium transition-colors hover:cursor-pointer"
          :class="
            cycle === 'monthly'
              ? 'bg-[var(--bg)] text-[var(--text-strong)] shadow-[0_1px_2px_rgb(0_0_0/12%)]'
              : 'text-[var(--text-dim)] hover:text-[var(--text)]'
          "
          @click="cycle = 'monthly'"
        >
          Monthly
        </button>
        <button
          type="button"
          class="min-h-[34px] rounded-lg px-4 text-sm font-medium transition-colors hover:cursor-pointer"
          :class="
            cycle === 'annual'
              ? 'bg-[var(--bg)] text-[var(--text-strong)] shadow-[0_1px_2px_rgb(0_0_0/12%)]'
              : 'text-[var(--text-dim)] hover:text-[var(--text)]'
          "
          @click="cycle = 'annual'"
        >
          Annual
          <span class="ml-1 text-[var(--accent)]">— 2 months free</span>
        </button>
      </div>
    </section>

    <!-- Plan cards -->
    <section class="mx-auto max-w-[1200px] px-4 pb-16 sm:px-6 sm:pb-24">
      <div
        class="grid gap-px overflow-hidden rounded-[18px] border border-[var(--border)] bg-[var(--border)] sm:grid-cols-2 lg:grid-cols-4"
      >
        <div
          v-for="plan in plans"
          :key="plan.name"
          class="flex flex-col p-7"
          :class="
            plan.highlight ? 'bg-[color-mix(in_srgb,var(--accent)_6%,var(--bg))]' : 'bg-[var(--bg)]'
          "
        >
          <div class="mb-5">
            <div class="mb-1 flex items-center justify-between gap-2">
              <span class="text-base font-semibold text-[var(--text-strong)]">{{ plan.name }}</span>
              <span v-if="plan.highlight" class="text-xs font-semibold text-[var(--accent)]"
                >Most popular</span
              >
            </div>
            <p class="text-xs leading-relaxed text-[var(--text-muted)]">
              {{ plan.description }}
            </p>
          </div>

          <div class="mb-6">
            <template v-if="plan.price === 0">
              <span class="text-4xl font-semibold tracking-[-0.035em] text-[var(--text-strong)]"
                >Free</span
              >
            </template>
            <template v-else-if="cycle === 'monthly'">
              <span class="text-4xl font-semibold tracking-[-0.035em] text-[var(--text-strong)]"
                >${{ plan.price }}</span
              >
              <span class="text-sm ml-1 text-[var(--text-muted)]">/month</span>
            </template>
            <template v-else>
              <span class="text-4xl font-semibold tracking-[-0.035em] text-[var(--text-strong)]"
                >${{ plan.annualPrice }}</span
              >
              <span class="text-sm ml-1 text-[var(--text-muted)]">/year</span>
              <p class="text-xs mt-1 text-[var(--text-muted)]">
                (${{ effectiveMonthly(plan) }}/mo, billed annually)
              </p>
            </template>
          </div>

          <ul class="space-y-3 mb-8 flex-1">
            <li class="flex items-start gap-2.5 text-sm text-[var(--text-dim)]">
              <svg
                class="flex-shrink-0 mt-0.5 text-[var(--accent)]"
                width="14"
                height="14"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              <span
                ><strong class="text-[var(--text-strong)]">{{ plan.monitors }}</strong>
                monitors</span
              >
            </li>
            <li class="flex items-start gap-2.5 text-sm text-[var(--text-dim)]">
              <svg
                class="flex-shrink-0 mt-0.5 text-[var(--accent)]"
                width="14"
                height="14"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              <span
                ><strong class="text-[var(--text-strong)]">{{ plan.checkInterval }}</strong> check
                interval</span
              >
            </li>
            <li class="flex items-start gap-2.5 text-sm text-[var(--text-dim)]">
              <svg
                class="flex-shrink-0 mt-0.5 text-[var(--accent)]"
                width="14"
                height="14"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              <span
                ><strong class="text-[var(--text-strong)]">{{ plan.statusPages }}</strong> status
                {{ plan.statusPages === '1' ? 'page' : 'pages' }}</span
              >
            </li>
            <li class="flex items-start gap-2.5 text-sm text-[var(--text-dim)]">
              <svg
                class="flex-shrink-0 mt-0.5 text-[var(--accent)]"
                width="14"
                height="14"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              <span
                ><strong class="text-[var(--text-strong)]">{{ plan.notificationChannels }}</strong>
                notification channels</span
              >
            </li>
            <li class="flex items-start gap-2.5 text-sm text-[var(--text-dim)]">
              <svg
                class="flex-shrink-0 mt-0.5 text-[var(--accent)]"
                width="14"
                height="14"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              Cron, uptime, SSL, domain, port &amp; DNS monitors
            </li>
            <li class="flex items-start gap-2.5 text-sm text-[var(--text-dim)]">
              <svg
                class="flex-shrink-0 mt-0.5 text-[var(--accent)]"
                width="14"
                height="14"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              Telegram, email, Slack &amp; webhook alerts
            </li>
            <li
              v-if="plan.smsCredits !== '0'"
              class="flex items-start gap-2.5 text-sm text-[var(--text-dim)]"
            >
              <svg
                class="flex-shrink-0 mt-0.5 text-[var(--accent)]"
                width="14"
                height="14"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              <span
                ><strong class="text-[var(--text-strong)]">{{ plan.smsCredits }}</strong> SMS
                credits / month</span
              >
            </li>
          </ul>

          <RouterLink
            to="/sign-up"
            class="block rounded-[10px] px-4 py-2.5 text-center text-sm font-semibold transition-colors"
            :class="
              plan.highlight
                ? 'bg-[var(--accent)] text-[var(--on-accent)] hover:bg-[var(--accent-hover)]'
                : 'border border-[var(--border-strong)] text-[var(--text-strong)] hover:bg-[var(--surface-raised)]'
            "
          >
            {{ plan.cta }}
          </RouterLink>
        </div>
      </div>

      <p class="mt-6 text-[13px] text-[var(--text-dim)]">
        All prices in USD. Billing handled by Paddle — taxes calculated at checkout based on your
        location.
      </p>
    </section>

    <!-- Feature comparison table -->
    <section class="mx-auto max-w-[1200px] px-4 pb-20 sm:px-6 sm:pb-28">
      <h2 class="mb-6 text-[22px] font-semibold tracking-[-0.02em] text-[var(--text-strong)]">
        Full feature comparison
      </h2>

      <div class="overflow-x-auto rounded-[14px] border border-[var(--border)]">
        <!-- Table header -->
        <div class="grid min-w-[640px] grid-cols-5 bg-[var(--surface)]">
          <div class="col-span-1 px-4 py-2.5 text-[12.5px] font-medium text-[var(--text-dim)]">
            Feature
          </div>
          <div
            v-for="plan in plans"
            :key="plan.name"
            class="px-4 py-2.5 text-center text-[12.5px] font-semibold"
            :class="plan.highlight ? 'text-[var(--accent)]' : 'text-[var(--text-dim)]'"
          >
            {{ plan.name }}
          </div>
        </div>

        <!-- Table rows -->
        <div
          v-for="row in featureRows"
          :key="row.label"
          class="grid min-w-[640px] grid-cols-5 border-t border-[var(--border)]"
        >
          <div class="col-span-1 px-4 py-3 text-[13px] text-[var(--text)]">
            {{ row.label }}
          </div>
          <div
            v-for="(val, j) in row.values"
            :key="j"
            class="px-4 py-3 text-center font-mono text-[13px]"
            :style="{ color: val === '—' ? 'var(--text-muted)' : 'var(--text-strong)' }"
          >
            <!-- An icon, not the ✓ glyph: at 13px the glyph was too small to scan. -->
            <svg
              v-if="val === '✓'"
              role="img"
              aria-label="Included"
              class="mx-auto text-[var(--accent)]"
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="3"
            >
              <polyline points="20 6 9 17 4 12" />
            </svg>
            <template v-else>{{ val }}</template>
          </div>
        </div>
      </div>
    </section>

    <!-- FAQ -->
    <section id="billing" class="mx-auto max-w-[1200px] px-4 pb-20 sm:px-6 sm:pb-28">
      <h2 class="mb-6 text-[22px] font-semibold tracking-[-0.02em] text-[var(--text-strong)]">
        Frequently asked questions
      </h2>

      <div class="max-w-[760px]">
        <div v-for="faq in billingFaqs" :key="faq.q" class="border-t border-[var(--border)] py-5">
          <h3 class="mb-2 font-semibold text-[var(--text-strong)]">{{ faq.q }}</h3>
          <p class="text-[15px] leading-relaxed text-[var(--text-dim)]">{{ faq.a }}</p>
        </div>
      </div>

      <p class="mt-6">
        <RouterLink to="/faq" class="text-sm font-semibold text-[var(--accent)] hover:underline">
          See all FAQs →
        </RouterLink>
      </p>
    </section>

    <CtaBanner
      title="Ten monitors free. Pay when you need more."
      subtitle="No credit card to start. Change plan any time from Billing."
    />
  </LandingLayout>
</template>
