<script setup lang="ts">
import { RouterLink } from 'vue-router'
import LandingLayout from '@/layouts/LandingLayout.vue'

interface Plan {
  name: string
  price: number
  description: string
  highlight: boolean
  cta: string
  monitors: string
  statusPages: string
  checkInterval: string
}

const plans: Plan[] = [
  {
    name: 'Hobbyist',
    price: 0,
    description: 'Side projects and personal tools',
    highlight: false,
    cta: 'Get started free',
    monitors: '10',
    statusPages: '1',
    checkInterval: '5 min',
  },
  {
    name: 'Indie',
    price: 9,
    description: 'Solo builders shipping products',
    highlight: false,
    cta: 'Start Indie',
    monitors: '30',
    statusPages: '3',
    checkInterval: '1 min',
  },
  {
    name: 'Studio',
    price: 29,
    description: 'Small teams and agencies',
    highlight: true,
    cta: 'Start Studio',
    monitors: '100',
    statusPages: '10',
    checkInterval: '1 min',
  },
  {
    name: 'Agency',
    price: 79,
    description: 'Agencies with many clients',
    highlight: false,
    cta: 'Start Agency',
    monitors: 'Unlimited',
    statusPages: 'Unlimited',
    checkInterval: '1 min',
  },
]

interface TableRow {
  label: string
  values: string[]
}

const featureRows: TableRow[] = [
  { label: 'Monitors (all types combined)', values: ['10', '30', '100', 'Unlimited'] },
  { label: 'Uptime check interval', values: ['5 min', '1 min', '1 min', '1 min'] },
  { label: 'Status pages', values: ['1', '3', '10', 'Unlimited'] },
  { label: 'Cron job monitoring', values: ['✓', '✓', '✓', '✓'] },
  { label: 'Uptime monitoring', values: ['✓', '✓', '✓', '✓'] },
  { label: 'SSL expiry monitoring', values: ['✓', '✓', '✓', '✓'] },
  { label: 'Execution history & logs', values: ['✓', '✓', '✓', '✓'] },
  { label: 'Telegram alerts', values: ['✓', '✓', '✓', '✓'] },
  { label: 'White-label status pages', values: ['—', '✓', '✓', '✓'] },
]

const faqs = [
  {
    q: 'Do I need a credit card to start?',
    a: 'No. The Hobbyist plan is free forever with no credit card required.',
  },
  {
    q: 'What counts as a "monitor"?',
    a: 'Each cron job, uptime URL, or SSL certificate you track counts as one monitor. The limit applies to the total across all types.',
  },
  {
    q: 'Can I change plans later?',
    a: 'Yes — upgrade or downgrade at any time. Billing adjusts automatically on your next cycle.',
  },
  {
    q: 'What happens if I exceed my monitor limit?',
    a: "You'll see an error when creating a new monitor and can choose to upgrade. Existing monitors keep running — we never pause them mid-cycle.",
  },
  {
    q: 'Which payment methods do you accept?',
    a: 'All major credit and debit cards via LemonSqueezy, which handles global tax compliance so you pay the right tax wherever you are.',
  },
  {
    q: 'Is there a refund policy?',
    a: 'Yes. Contact us within 30 days of any charge and we\'ll issue a full refund, no questions asked.',
  },
]
</script>

<template>
  <LandingLayout>

    <!-- Hero -->
    <section class="max-w-4xl mx-auto px-4 sm:px-6 pt-16 pb-12 sm:pt-24 sm:pb-16 text-center">
      <h1 class="text-4xl sm:text-5xl font-bold tracking-tight mb-4" style="color: var(--text-strong)">
        Simple, honest pricing
      </h1>
      <p class="text-lg sm:text-xl max-w-xl mx-auto" style="color: var(--text-dim)">
        Start free. Pay only when you grow. No per-seat fees, no hidden costs.
      </p>
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
            style="background-color: var(--color-green-500); color: #fff"
          >
            Most popular
          </div>

          <div class="mb-5">
            <div class="text-base font-semibold mb-1" style="color: var(--text-strong)">{{ plan.name }}</div>
            <p class="text-xs leading-relaxed" style="color: var(--text-muted)">{{ plan.description }}</p>
          </div>

          <div class="mb-6">
            <span class="text-4xl font-bold" style="color: var(--text-strong)">
              {{ plan.price === 0 ? 'Free' : `$${plan.price}` }}
            </span>
            <span v-if="plan.price > 0" class="text-sm ml-1" style="color: var(--text-muted)">/month</span>
          </div>

          <ul class="space-y-3 mb-8 flex-1">
            <li class="flex items-start gap-2.5 text-sm" style="color: var(--text-dim)">
              <svg class="flex-shrink-0 mt-0.5" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" style="color: var(--color-green-500)"><polyline points="20 6 9 17 4 12"/></svg>
              <span><strong style="color: var(--text-strong)">{{ plan.monitors }}</strong> monitors</span>
            </li>
            <li class="flex items-start gap-2.5 text-sm" style="color: var(--text-dim)">
              <svg class="flex-shrink-0 mt-0.5" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" style="color: var(--color-green-500)"><polyline points="20 6 9 17 4 12"/></svg>
              <span><strong style="color: var(--text-strong)">{{ plan.checkInterval }}</strong> check interval</span>
            </li>
            <li class="flex items-start gap-2.5 text-sm" style="color: var(--text-dim)">
              <svg class="flex-shrink-0 mt-0.5" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" style="color: var(--color-green-500)"><polyline points="20 6 9 17 4 12"/></svg>
              <span><strong style="color: var(--text-strong)">{{ plan.statusPages }}</strong> status {{ plan.statusPages === '1' ? 'page' : 'pages' }}</span>
            </li>
            <li class="flex items-start gap-2.5 text-sm" style="color: var(--text-dim)">
              <svg class="flex-shrink-0 mt-0.5" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" style="color: var(--color-green-500)"><polyline points="20 6 9 17 4 12"/></svg>
              Cron, uptime &amp; SSL monitors
            </li>
            <li class="flex items-start gap-2.5 text-sm" style="color: var(--text-dim)">
              <svg class="flex-shrink-0 mt-0.5" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" style="color: var(--color-green-500)"><polyline points="20 6 9 17 4 12"/></svg>
              Telegram alerts
            </li>
          </ul>

          <RouterLink
            to="/sign-up"
            class="block text-sm font-semibold text-center px-4 py-2.5 rounded-md transition-colors"
            :style="{
              backgroundColor: plan.highlight ? 'var(--color-green-500)' : 'var(--surface-raised)',
              color: plan.highlight ? '#fff' : 'var(--text)',
              border: plan.highlight ? 'none' : '1px solid var(--border)',
            }"
          >
            {{ plan.cta }}
          </RouterLink>
        </div>
      </div>

      <p class="text-center text-xs mt-6" style="color: var(--text-muted)">
        All prices in USD. Billing handled by LemonSqueezy — taxes calculated at checkout based on your location.
      </p>
    </section>

    <!-- Feature comparison table -->
    <section class="max-w-6xl mx-auto px-4 sm:px-6 pb-20 sm:pb-28">
      <h2 class="text-xl font-bold mb-6 text-center" style="color: var(--text-strong)">Full feature comparison</h2>

      <div class="rounded-xl border overflow-hidden" style="border-color: var(--border)">
        <!-- Table header -->
        <div class="grid grid-cols-5 border-b" style="border-color: var(--border); background-color: var(--surface)">
          <div class="col-span-1 px-4 py-3 text-xs font-medium" style="color: var(--text-muted)">Feature</div>
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
          :style="{ borderColor: 'var(--border)', backgroundColor: i % 2 === 0 ? 'var(--bg)' : 'var(--surface)' }"
        >
          <div class="col-span-1 px-4 py-3 text-xs" style="color: var(--text-dim)">{{ row.label }}</div>
          <div
            v-for="(val, j) in row.values"
            :key="j"
            class="px-4 py-3 text-xs text-center font-medium"
            :style="{ color: val === '—' ? 'var(--text-muted)' : val === '✓' ? 'var(--color-green-500)' : 'var(--text-strong)' }"
          >
            {{ val }}
          </div>
        </div>
      </div>
    </section>

    <!-- FAQ -->
    <section class="max-w-3xl mx-auto px-4 sm:px-6 pb-20 sm:pb-28">
      <h2 class="text-2xl font-bold mb-10 text-center" style="color: var(--text-strong)">Frequently asked questions</h2>

      <div class="space-y-0 divide-y" style="border-color: var(--border)">
        <div
          v-for="faq in faqs"
          :key="faq.q"
          class="py-5"
          style="border-color: var(--border)"
        >
          <h3 class="text-sm font-semibold mb-2" style="color: var(--text-strong)">{{ faq.q }}</h3>
          <p class="text-sm leading-relaxed" style="color: var(--text-dim)">{{ faq.a }}</p>
        </div>
      </div>
    </section>

    <!-- CTA -->
    <section class="max-w-6xl mx-auto px-4 sm:px-6 pb-20 sm:pb-28">
      <div
        class="rounded-2xl text-center px-8 py-14 border"
        style="background: linear-gradient(135deg, var(--color-green-900) 0%, var(--surface) 100%); border-color: var(--color-green-700)"
      >
        <h2 class="text-2xl sm:text-3xl font-bold mb-3" style="color: var(--text-strong)">
          Start monitoring in 60 seconds.
        </h2>
        <p class="mb-8 text-base" style="color: var(--text-dim)">
          Free plan included. No credit card required.
        </p>
        <RouterLink
          to="/sign-up"
          class="inline-flex items-center gap-2 text-sm font-semibold px-7 py-3 rounded-md transition-colors"
          style="background-color: var(--color-green-500); color: #fff"
        >
          Create free account →
        </RouterLink>
      </div>
    </section>

  </LandingLayout>
</template>
