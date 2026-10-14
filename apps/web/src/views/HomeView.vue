<script setup lang="ts">
import { computed } from 'vue'
import { useHead } from '@unhead/vue'
import LandingLayout from '@/layouts/LandingLayout.vue'
import CtaBanner from '@/components/CtaBanner.vue'
import HeroSection from '@/components/HeroSection.vue'
import CheckStrip from '@/components/CheckStrip.vue'
import { useTheme } from '@/lib/theme'
import { useSeo } from '@/composables/useSeo'
import logoDark from '@/assets/logo-dark.svg'
import logoLight from '@/assets/logo-light.svg'
import { statusRows, plans, testimonials, monitorTypes } from '@/content/home'

const { theme } = useTheme()
const logo = computed(() => (theme.value === 'light' ? logoLight : logoDark))

// A 40-tick strip for the status page mockup: as many failed ticks as the
// row's uptime percentage leaves, spread through the run.
function ticksFor(pct: string) {
  const failures = Math.round(((100 - parseFloat(pct)) / 100) * 40)
  const every = failures > 0 ? Math.floor(40 / (failures + 1)) : 0
  return Array.from({ length: 40 }, (_, i) => ({
    ok: !(failures > 0 && i > 0 && i % every === 0 && i / every <= failures),
  }))
}

useSeo({
  title: 'Checkmeup — monitoring for freelancers & client sites',
  description:
    'Cron, uptime, SSL, domain expiry, port (TCP), and DNS record monitors with execution logs, Telegram alerts, and branded status pages — built for freelancers and solo devs managing client sites.',
  path: '/',
})

// Organization + SoftwareApplication structured data, so search results can
// show Checkmeup's identity (logo, GitHub) and pricing directly rather than
// just a title/description — same idea as the FAQPage/Article schema already
// on the FAQ and blog-post pages. Offers are derived from the real `plans`
// data so this can't silently drift from actual pricing.
const organizationSchema = {
  '@context': 'https://schema.org',
  '@type': 'Organization',
  name: 'Checkmeup',
  url: 'https://checkmeup.net',
  logo: 'https://checkmeup.net/img/checkmeup-og.png',
  sameAs: ['https://github.com/checkmeup/checkmeup'],
}

const softwareApplicationSchema = {
  '@context': 'https://schema.org',
  '@type': 'SoftwareApplication',
  name: 'Checkmeup',
  applicationCategory: 'BusinessApplication',
  operatingSystem: 'Web',
  url: 'https://checkmeup.net',
  description:
    'Cron, uptime, SSL, domain expiry, port (TCP), and DNS record monitors with execution logs, Telegram alerts, and branded status pages — built for freelancers and solo devs managing client sites.',
  offers: plans.map((plan) => ({
    '@type': 'Offer',
    name: plan.name,
    price: plan.price,
    priceCurrency: 'USD',
  })),
}

useHead({
  script: [
    { type: 'application/ld+json', innerHTML: JSON.stringify(organizationSchema) },
    { type: 'application/ld+json', innerHTML: JSON.stringify(softwareApplicationSchema) },
  ],
})
</script>

<template>
  <LandingLayout>
    <HeroSection />

    <!-- Monitor types: a spec sheet, one row per type -->
    <section class="border-t border-[var(--border)]">
      <div class="mx-auto max-w-[1200px] px-4 py-16 sm:px-6 sm:py-24">
        <div class="mb-10 max-w-[640px]">
          <h2
            class="mb-3 text-[30px] leading-[1.1] font-semibold tracking-[-0.035em] text-[var(--text-strong)] sm:text-[38px]"
          >
            Six monitor types, one dashboard.
          </h2>
          <p class="text-base text-[var(--text-dim)]">
            What each one watches, how often it looks, and when it wakes you up.
          </p>
        </div>

        <div class="overflow-x-auto rounded-[14px] border border-[var(--border)]">
          <table class="w-full min-w-[760px] border-collapse text-left text-sm">
            <thead>
              <tr class="bg-[var(--surface)] text-[13px] text-[var(--text-dim)]">
                <th scope="col" class="px-5 py-3 font-medium">Monitor</th>
                <th scope="col" class="px-5 py-3 font-medium">Watches</th>
                <th scope="col" class="px-5 py-3 font-medium">How often</th>
                <th scope="col" class="px-5 py-3 font-medium">Alerts you when</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="m in monitorTypes"
                :key="m.name"
                class="border-t border-[var(--border)] align-top"
              >
                <th scope="row" class="px-5 py-4 font-semibold text-[var(--text-strong)]">
                  {{ m.name }}
                </th>
                <td class="px-5 py-4 text-[var(--text-dim)]">{{ m.watches }}</td>
                <td class="px-5 py-4 whitespace-nowrap">
                  <span class="font-mono text-[13px] text-[var(--text)]">{{ m.often }}</span>
                  <span v-if="m.oftenNote" class="block text-xs text-[var(--text-muted)]">{{
                    m.oftenNote
                  }}</span>
                </td>
                <td class="px-5 py-4 text-[var(--text-dim)]">{{ m.alerts }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <p class="mt-4 text-[13px] text-[var(--text-muted)]">
          Alerts go to Telegram, email, Slack, a webhook, or SMS on paid plans.
          <RouterLink to="/docs" class="text-[var(--accent)] hover:underline"
            >How each one works →</RouterLink
          >
        </p>
      </div>
    </section>

    <!-- Status pages highlight -->
    <section class="border-t border-[var(--border)]">
      <div
        class="mx-auto grid max-w-[1200px] items-center gap-12 px-4 py-20 sm:px-6 lg:grid-cols-[1fr_1.1fr] lg:gap-16"
      >
        <div class="flex flex-col gap-5">
          <p class="text-[13px] font-semibold text-[var(--accent)]">Status pages</p>
          <h2
            class="text-[30px] leading-[1.1] font-semibold tracking-[-0.035em] text-[var(--text-strong)] sm:text-[38px]"
          >
            Give your clients a page they can bookmark.
          </h2>
          <p class="text-[17px] leading-relaxed text-[var(--text-dim)]">
            Create a branded status page for every client site. They see uptime, incidents, and
            maintenance windows — you stay in control, and it looks like your work.
          </p>
          <RouterLink
            to="/sign-up"
            class="self-start rounded-[10px] bg-[var(--accent)] px-5 py-2.5 text-sm font-semibold text-[var(--on-accent)] transition-colors hover:bg-[var(--accent-hover)]"
          >
            Create a status page →
          </RouterLink>

          <div class="mt-2 flex flex-col gap-2.5 border-t border-[var(--border)] pt-5">
            <p class="text-sm text-[var(--text-dim)]">
              Embed a live status badge anywhere — links straight to your status page
            </p>
            <a href="https://checkmeup.net/status/checkmeup-net" class="self-start">
              <img
                src="https://checkmeup.net/status/checkmeup-net/badge.svg"
                alt="When you run a successful product, your monitors run quietly"
              />
            </a>
          </div>
        </div>

        <!-- Status page mockup: mirrors the real public page's layout -->
        <div
          class="flex flex-col gap-6 rounded-[18px] border border-[var(--border)] bg-[var(--bg)] p-6 shadow-[0_24px_60px_rgb(0_0_0/8%)] sm:p-8"
        >
          <img :src="logo" alt="Checkmeup" class="h-[18px] w-auto self-start" />
          <div class="flex items-center gap-3">
            <span
              class="h-3 w-3 rounded-full bg-[var(--status-up)] shadow-[0_0_0_5px_var(--accent-wash)]"
            ></span>
            <span class="text-[24px] font-semibold tracking-[-0.03em] text-[var(--text-strong)]"
              >All systems operational</span
            >
          </div>
          <div class="flex flex-col">
            <div
              v-for="row in statusRows"
              :key="row.name"
              class="flex flex-col gap-2.5 border-t border-[var(--border)] py-4"
            >
              <div class="flex items-baseline justify-between gap-3">
                <span class="text-sm font-medium text-[var(--text-strong)]">{{ row.name }}</span>
                <span class="inline-flex items-center gap-2 text-[13px] text-[var(--text-dim)]">
                  <span class="h-2 w-2 rounded-full bg-[var(--status-up)]"></span>
                  Operational
                </span>
              </div>
              <CheckStrip size="sm" :checks="ticksFor(row.pct)" />
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Pricing teaser -->
    <section class="border-t border-[var(--border)]">
      <div class="mx-auto max-w-[1200px] px-4 py-20 sm:px-6">
        <div class="mb-10 flex flex-wrap items-end justify-between gap-4">
          <div class="flex flex-col gap-3">
            <h2
              class="text-[30px] leading-[1.1] font-semibold tracking-[-0.035em] text-[var(--text-strong)] sm:text-[38px]"
            >
              Simple pricing.
            </h2>
            <p class="text-[17px] text-[var(--text-dim)]">Start free. Scale as you grow.</p>
          </div>
          <RouterLink
            to="/pricing"
            class="text-sm font-semibold text-[var(--accent)] hover:underline"
          >
            See full pricing details →
          </RouterLink>
        </div>

        <div
          class="grid gap-px overflow-hidden rounded-[18px] border border-[var(--border)] bg-[var(--border)] sm:grid-cols-2 lg:grid-cols-4"
        >
          <div
            v-for="plan in plans"
            :key="plan.name"
            class="flex flex-col p-6"
            :class="
              plan.highlight
                ? 'bg-[color-mix(in_srgb,var(--accent)_6%,var(--bg))]'
                : 'bg-[var(--bg)]'
            "
          >
            <div class="mb-1 flex items-center justify-between gap-2">
              <span class="font-semibold text-[var(--text-strong)]">{{ plan.name }}</span>
              <span v-if="plan.highlight" class="text-xs font-semibold text-[var(--accent)]"
                >Most popular</span
              >
            </div>
            <p class="mb-5 text-[13px] text-[var(--text-dim)]">{{ plan.description }}</p>
            <div
              class="mb-6 text-[32px] font-semibold tracking-[-0.035em] text-[var(--text-strong)]"
            >
              {{ plan.price === 0 ? 'Free' : `$${plan.price}` }}
              <span
                v-if="plan.price > 0"
                class="text-sm font-normal tracking-normal text-[var(--text-dim)]"
                >/mo</span
              >
            </div>

            <ul class="mb-6 flex flex-col gap-2.5 text-[13px] text-[var(--text-dim)]">
              <li v-for="st in plan.stats" :key="st.label" class="flex items-center gap-2">
                <svg
                  aria-hidden="true"
                  width="11"
                  height="11"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="var(--accent)"
                  stroke-width="3"
                >
                  <polyline points="20 6 9 17 4 12" />
                </svg>
                <span
                  ><strong class="font-semibold text-[var(--text-strong)]">{{ st.value }}</strong>
                  {{ st.label }}</span
                >
              </li>
              <li v-for="e in plan.extras" :key="e" class="flex items-start gap-2 leading-relaxed">
                <svg
                  aria-hidden="true"
                  width="11"
                  height="11"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="var(--accent)"
                  stroke-width="3"
                  class="mt-[3px] flex-shrink-0"
                >
                  <polyline points="20 6 9 17 4 12" />
                </svg>
                {{ e }}
              </li>
            </ul>

            <RouterLink
              to="/sign-up"
              class="mt-auto rounded-[10px] px-4 py-2.5 text-center text-sm font-semibold transition-colors"
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
      </div>
    </section>

    <!-- Testimonials -->
    <section class="border-t border-[var(--border)]">
      <div class="mx-auto max-w-[1200px] px-4 py-20 sm:px-6">
        <div class="mb-12 flex max-w-[640px] flex-col gap-3">
          <h2
            class="text-[30px] leading-[1.1] font-semibold tracking-[-0.035em] text-[var(--text-strong)] sm:text-[38px]"
          >
            Loved by freelancers and solo devs.
          </h2>
          <p class="text-[17px] text-[var(--text-dim)]">
            Monitoring that gets out of your way — so you can focus on client work.
          </p>
        </div>

        <div class="grid gap-x-10 gap-y-12 sm:grid-cols-3">
          <figure
            v-for="t in testimonials"
            :key="t.name"
            class="flex flex-col gap-5 border-t border-[var(--border)] pt-6"
          >
            <blockquote class="text-[17px] leading-relaxed text-[var(--text)]">
              “{{ t.quote }}”
            </blockquote>
            <figcaption class="mt-auto flex items-center gap-3">
              <img
                :src="t.avatar"
                :alt="t.name"
                class="h-9 w-9 flex-shrink-0 rounded-full object-cover"
              />
              <span class="flex flex-col">
                <span class="text-sm font-semibold text-[var(--text-strong)]">{{ t.name }}</span>
                <span class="text-[13px] text-[var(--text-dim)]">{{ t.role }}</span>
              </span>
            </figcaption>
          </figure>
        </div>
      </div>
    </section>

    <CtaBanner
      title="Start monitoring in 60 seconds."
      subtitle="Free plan included. No credit card required."
    />
  </LandingLayout>
</template>
