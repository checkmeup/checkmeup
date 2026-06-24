<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import LandingLayout from '@/layouts/LandingLayout.vue'
import { findFaqCategory } from '@/faq/faqs'

const billingFaqs = findFaqCategory('billing')?.entries ?? []

// Annual price is always exactly 10x the monthly price — ~2 months free (EP-27).
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
}

const plans: Plan[] = [
  {
    name: 'Hobby',
    price: 0,
    annualPrice: 0,
    description: 'Side projects and personal tools',
    highlight: false,
    cta: 'Get started free',
    monitors: '10',
    statusPages: '1',
    checkInterval: '5 min',
  },
  {
    name: 'Solo',
    price: 9,
    annualPrice: 90,
    description: 'Solo builders shipping products',
    highlight: false,
    cta: 'Start Solo',
    monitors: '30',
    statusPages: '3',
    checkInterval: '1 min',
  },
  {
    name: 'Startup',
    price: 29,
    annualPrice: 290,
    description: 'Small teams and agencies',
    highlight: true,
    cta: 'Start Startup',
    monitors: '100',
    statusPages: '10',
    checkInterval: '1 min',
  },
  {
    name: 'Enterprise',
    price: 99,
    annualPrice: 990,
    description: 'Agencies with many clients',
    highlight: false,
    cta: 'Start Enterprise',
    monitors: '1000',
    statusPages: '100',
    checkInterval: '1 min',
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
  { label: 'Cron job monitoring', values: ['✓', '✓', '✓', '✓'] },
  { label: 'Uptime monitoring', values: ['✓', '✓', '✓', '✓'] },
  { label: 'Keyword monitoring', values: ['✓', '✓', '✓', '✓'] },
  { label: 'SSL expiry monitoring', values: ['✓', '✓', '✓', '✓'] },
  { label: 'Domain expiry monitoring', values: ['✓', '✓', '✓', '✓'] },
  { label: 'Execution history & logs', values: ['✓', '✓', '✓', '✓'] },
  { label: 'Alerts (Telegram, email, Slack)', values: ['✓', '✓', '✓', '✓'] },
  { label: 'Maintenance windows', values: ['✓', '✓', '✓', '✓'] },
  { label: 'White-label status pages', values: ['—', '✓', '✓', '✓'] },
]


</script>

<template>
  <LandingLayout>
    <!-- Hero -->
    <section class="max-w-4xl mx-auto px-4 sm:px-6 pt-16 pb-12 sm:pt-24 sm:pb-16 text-center">
      <h1
        class="text-4xl sm:text-5xl font-bold tracking-tight mb-4"
        style="color: var(--text-strong)"
      >
        Simple, honest pricing
      </h1>
      <p class="text-lg sm:text-xl max-w-xl mx-auto mb-8" style="color: var(--text-dim)">
        Start free. Pay only when you grow. No per-seat fees, no hidden costs.
      </p>

      <div class="inline-flex rounded-md border p-1" style="border-color: var(--border)">
        <button
          type="button"
          class="px-4 py-1.5 rounded text-sm transition-colors hover:cursor-pointer"
          :style="
            cycle === 'monthly'
              ? 'background-color: var(--surface-raised); color: var(--text-strong)'
              : 'color: var(--text-muted)'
          "
          @click="cycle = 'monthly'"
        >
          Monthly
        </button>
        <button
          type="button"
          class="px-4 py-1.5 rounded text-sm transition-colors hover:cursor-pointer"
          :style="
            cycle === 'annual'
              ? 'background-color: var(--surface-raised); color: var(--text-strong)'
              : 'color: var(--text-muted)'
          "
          @click="cycle = 'annual'"
        >
          Annual
          <span class="ml-1" style="color: var(--color-green-500)">— 2 months free</span>
        </button>
      </div>
    </section>

    <!-- Plan cards -->
    <section class="max-w-6xl mx-auto px-4 sm:px-6 pb-16 sm:pb-24">
      <div class="grid sm:grid-cols-2 lg:grid-cols-4 gap-5">
        <div
          v-for="plan in plans"
          :key="plan.name"
          class="rounded-2xl border p-7 flex flex-col relative"
          :style="{
            backgroundColor: plan.highlight ? 'var(--surface-raised)' : 'var(--surface)',
            borderColor: plan.highlight ? 'var(--color-green-500)' : 'var(--border)',
          }"
        >
          <!-- Popular badge -->
          <div
            v-if="plan.highlight"
            class="absolute -top-3.5 left-1/2 -translate-x-1/2 text-xs font-semibold px-3 py-1 rounded-full"
            style="background-color: var(--color-green-500); color: var(--on-accent)"
          >
            Most popular
          </div>

          <div class="mb-5">
            <div class="text-base font-semibold mb-1" style="color: var(--text-strong)">
              {{ plan.name }}
            </div>
            <p class="text-xs leading-relaxed" style="color: var(--text-muted)">
              {{ plan.description }}
            </p>
          </div>

          <div class="mb-6">
            <template v-if="plan.price === 0">
              <span class="text-4xl font-bold" style="color: var(--text-strong)">Free</span>
            </template>
            <template v-else-if="cycle === 'monthly'">
              <span class="text-4xl font-bold" style="color: var(--text-strong)">${{ plan.price }}</span>
              <span class="text-sm ml-1" style="color: var(--text-muted)">/month</span>
            </template>
            <template v-else>
              <span class="text-4xl font-bold" style="color: var(--text-strong)">${{ plan.annualPrice }}</span>
              <span class="text-sm ml-1" style="color: var(--text-muted)">/year</span>
              <p class="text-xs mt-1" style="color: var(--text-muted)">
                (${{ effectiveMonthly(plan) }}/mo, billed annually)
              </p>
            </template>
          </div>

          <ul class="space-y-3 mb-8 flex-1">
            <li class="flex items-start gap-2.5 text-sm" style="color: var(--text-dim)">
              <svg
                class="flex-shrink-0 mt-0.5"
                width="14"
                height="14"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                style="color: var(--color-green-500)"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              <span
                ><strong style="color: var(--text-strong)">{{ plan.monitors }}</strong>
                monitors</span
              >
            </li>
            <li class="flex items-start gap-2.5 text-sm" style="color: var(--text-dim)">
              <svg
                class="flex-shrink-0 mt-0.5"
                width="14"
                height="14"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                style="color: var(--color-green-500)"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              <span
                ><strong style="color: var(--text-strong)">{{ plan.checkInterval }}</strong> check
                interval</span
              >
            </li>
            <li class="flex items-start gap-2.5 text-sm" style="color: var(--text-dim)">
              <svg
                class="flex-shrink-0 mt-0.5"
                width="14"
                height="14"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                style="color: var(--color-green-500)"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              <span
                ><strong style="color: var(--text-strong)">{{ plan.statusPages }}</strong> status
                {{ plan.statusPages === '1' ? 'page' : 'pages' }}</span
              >
            </li>
            <li class="flex items-start gap-2.5 text-sm" style="color: var(--text-dim)">
              <svg
                class="flex-shrink-0 mt-0.5"
                width="14"
                height="14"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                style="color: var(--color-green-500)"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              Cron, uptime, SSL &amp; domain monitors
            </li>
            <li class="flex items-start gap-2.5 text-sm" style="color: var(--text-dim)">
              <svg
                class="flex-shrink-0 mt-0.5"
                width="14"
                height="14"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="3"
                style="color: var(--color-green-500)"
              >
                <polyline points="20 6 9 17 4 12" />
              </svg>
              Telegram, email &amp; Slack alerts
            </li>
          </ul>

          <RouterLink
            to="/sign-up"
            class="block text-sm font-semibold text-center px-4 py-2.5 rounded-md transition-colors"
            :style="{
              backgroundColor: plan.highlight ? 'var(--color-green-500)' : 'var(--surface-raised)',
              color: plan.highlight ? 'var(--on-accent)' : 'var(--text)',
              border: plan.highlight ? 'none' : '1px solid var(--border)',
            }"
          >
            {{ plan.cta }}
          </RouterLink>
        </div>
      </div>

      <p class="text-center text-xs mt-6" style="color: var(--text-muted)">
        All prices in USD. Billing handled by LemonSqueezy — taxes calculated at checkout based on
        your location.
      </p>
    </section>

    <!-- Feature comparison table -->
    <section class="max-w-6xl mx-auto px-4 sm:px-6 pb-20 sm:pb-28">
      <h2 class="text-xl font-bold mb-6 text-center" style="color: var(--text-strong)">
        Full feature comparison
      </h2>

      <div class="rounded-xl border overflow-hidden" style="border-color: var(--border)">
        <!-- Table header -->
        <div
          class="grid grid-cols-5 border-b"
          style="border-color: var(--border); background-color: var(--surface)"
        >
          <div class="col-span-1 px-4 py-3 text-xs font-medium" style="color: var(--text-muted)">
            Feature
          </div>
          <div
            v-for="plan in plans"
            :key="plan.name"
            class="px-4 py-3 text-xs font-semibold text-center"
            :style="{ color: plan.highlight ? 'var(--color-green-400)' : 'var(--text-dim)' }"
          >
            {{ plan.name }}
          </div>
        </div>

        <!-- Table rows -->
        <div
          v-for="(row, i) in featureRows"
          :key="row.label"
          class="grid grid-cols-5 border-b last:border-b-0"
          :style="{
            borderColor: 'var(--border)',
            backgroundColor: i % 2 === 0 ? 'var(--bg)' : 'var(--surface)',
          }"
        >
          <div class="col-span-1 px-4 py-3 text-xs" style="color: var(--text-dim)">
            {{ row.label }}
          </div>
          <div
            v-for="(val, j) in row.values"
            :key="j"
            class="px-4 py-3 text-xs text-center font-medium"
            :style="{
              color:
                val === '—'
                  ? 'var(--text-muted)'
                  : val === '✓'
                    ? 'var(--color-green-500)'
                    : 'var(--text-strong)',
            }"
          >
            {{ val }}
          </div>
        </div>
      </div>
    </section>

    <!-- FAQ -->
    <section id="billing" class="max-w-3xl mx-auto px-4 sm:px-6 pb-20 sm:pb-28">
      <h2 class="text-2xl font-bold mb-10 text-center" style="color: var(--text-strong)">
        Frequently asked questions
      </h2>

      <div class="space-y-0 divide-y" style="border-color: var(--border)">
        <div v-for="faq in billingFaqs" :key="faq.q" class="py-5" style="border-color: var(--border)">
          <h3 class="text-sm font-semibold mb-2" style="color: var(--text-strong)">{{ faq.q }}</h3>
          <p class="text-sm leading-relaxed" style="color: var(--text-dim)">{{ faq.a }}</p>
        </div>
      </div>

      <p class="text-center mt-8">
        <RouterLink to="/faq" class="text-sm transition-colors" style="color: var(--color-green-500)">
          See all FAQs →
        </RouterLink>
      </p>
    </section>

    <!-- CTA -->
    <section class="max-w-6xl mx-auto px-4 sm:px-6 pb-20 sm:pb-28">
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
