<script setup lang="ts">
import { computed } from 'vue'
import { useHead } from '@unhead/vue'
import LandingLayout from '@/layouts/LandingLayout.vue'
import CtaBanner from '@/components/CtaBanner.vue'
import HeroSection from '@/components/HeroSection.vue'
import FeatureCard from '@/components/FeatureCard.vue'
import CheckStrip from '@/components/CheckStrip.vue'
import { useTheme } from '@/lib/theme'
import { useSeo } from '@/composables/useSeo'
import logoDark from '@/assets/logo-dark.svg'
import logoLight from '@/assets/logo-light.svg'
import { statusRows, plans, customers, testimonials } from '@/content/home'

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

    <!-- Our Customers -->
    <section class="border-y border-[var(--border)]">
      <div
        class="mx-auto flex max-w-[1200px] flex-col items-start gap-4 px-4 py-8 sm:flex-row sm:items-center sm:justify-between sm:px-6"
      >
        <div class="flex items-center gap-4">
          <div class="flex -space-x-2.5">
            <img
              v-for="c in customers"
              :key="c.name"
              :src="c.avatar"
              :alt="c.name"
              :title="`${c.name} · ${c.role}`"
              class="h-7 w-7 rounded-full object-cover ring-2 ring-[var(--bg)]"
            />
          </div>
          <p class="text-sm text-[var(--text-dim)]">
            <span class="font-semibold text-[var(--text-strong)]">200+</span>
            freelancers monitoring client sites with Checkmeup
          </p>
        </div>
        <p class="text-[13px] text-[var(--text-muted)]">
          Freelancers · Solo agency operators · Web agencies · Solo developers · SaaS startups
        </p>
      </div>
    </section>

    <!-- Features -->
    <section class="max-w-[1200px] mx-auto px-4 sm:px-6 py-16 sm:py-24">
      <div class="mb-12 max-w-[640px]">
        <h2
          class="text-[30px] leading-[1.1] sm:text-[38px] font-semibold mb-3 tracking-[-0.035em] text-[var(--text-strong)]"
        >
          Everything you need to stay on top.
        </h2>
        <p class="text-base text-[var(--text-dim)]">
          Six monitor types, one dashboard, zero blind spots.
        </p>
      </div>

      <div class="grid gap-x-10 gap-y-12 sm:grid-cols-2 lg:grid-cols-3">
        <FeatureCard
          title="Cron job monitoring"
          description="Ping a URL after each run. Miss a ping and we alert you immediately — before your backups rot or invoices stop sending."
          :bullets="['Custom grace periods', 'Execution history & logs', 'Telegram, Slack, email & SMS alerts']"
        >
          <template #icon>
            <svg
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="var(--accent)"
              stroke-width="1.75"
            >
              <circle cx="12" cy="12" r="10" />
              <polyline points="12 6 12 12 16 14" />
            </svg>
          </template>
        </FeatureCard>

        <FeatureCard
          title="Uptime monitoring"
          description="We poll your URLs as often as every 1 minute and alert you the moment your site goes down or returns an unexpected status code."
          :bullets="['1-minute check intervals', 'Response time tracking', 'Incident history']"
        >
          <template #icon>
            <svg
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="var(--accent)"
              stroke-width="1.75"
            >
              <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
            </svg>
          </template>
        </FeatureCard>

        <FeatureCard
          title="SSL expiry monitoring"
          description="Get ahead of certificate expiry with early warnings at 30, 14, and 7 days. Never let a forgotten cert take your site offline."
          :bullets="['Multi-threshold alerts', 'Issuer & expiry details', 'Daily checks']"
        >
          <template #icon>
            <svg
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="var(--accent)"
              stroke-width="1.75"
            >
              <rect x="3" y="11" width="18" height="11" rx="2" />
              <path d="M7 11V7a5 5 0 0 1 10 0v4" />
            </svg>
          </template>
        </FeatureCard>

        <FeatureCard
          title="Domain expiry monitoring"
          description="Track your domain's registration separately from its SSL certificate — rarer to lapse, but far more catastrophic when it does."
          :bullets="['Multi-threshold alerts', 'Registrar & expiry details', 'Daily checks']"
        >
          <template #icon>
            <svg
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="var(--accent)"
              stroke-width="1.75"
            >
              <circle cx="12" cy="12" r="10" />
              <line x1="2" y1="12" x2="22" y2="12" />
              <path
                d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"
              />
            </svg>
          </template>
        </FeatureCard>

        <FeatureCard
          title="Port (TCP) monitoring"
          description="Raw TCP connect checks for mail servers, databases, and any other non-HTTP service — plus a security mode that alerts if a port that should be closed becomes reachable."
          :bullets="['Any host and port', 'Open or closed expected state', 'Connect-time history']"
        >
          <template #icon>
            <svg
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="var(--accent)"
              stroke-width="1.75"
            >
              <rect x="2" y="9" width="6" height="6" rx="1" />
              <rect x="16" y="9" width="6" height="6" rx="1" />
              <line x1="8" y1="12" x2="16" y2="12" />
            </svg>
          </template>
        </FeatureCard>

        <FeatureCard
          title="DNS record monitoring"
          description="Watch a hostname's A, AAAA, CNAME, MX, TXT, or NS record and get alerted the moment it changes or stops resolving — a hijack or a migration gone sideways doesn't fail the way a downed server does."
          :bullets="['Pin a value or auto-capture a baseline', 'Old value → new value in every alert', 'Lookup errors read differently than a change']"
        >
          <template #icon>
            <svg
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="var(--accent)"
              stroke-width="1.75"
            >
              <circle cx="11" cy="11" r="8" />
              <line x1="21" y1="21" x2="16.65" y2="16.65" />
            </svg>
          </template>
        </FeatureCard>

        <FeatureCard
          title="Public status pages"
          description="A branded status page for every client site — your logo, uptime, incidents, and maintenance windows, no login required to view."
          :bullets="['Your logo, your branding', 'No DNS setup required', 'Embeddable status badges']"
        >
          <template #icon>
            <svg
              width="18"
              height="18"
              viewBox="0 0 24 24"
              fill="none"
              stroke="var(--accent)"
              stroke-width="1.75"
            >
              <rect x="3" y="4" width="18" height="14" rx="2" />
              <line x1="3" y1="8" x2="21" y2="8" />
              <circle cx="17" cy="15" r="2" />
            </svg>
          </template>
        </FeatureCard>
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
            <div class="mb-6 text-[32px] font-semibold tracking-[-0.035em] text-[var(--text-strong)]">
              {{ plan.price === 0 ? 'Free' : `$${plan.price}` }}
              <span v-if="plan.price > 0" class="text-sm font-normal tracking-normal text-[var(--text-dim)]"
                >/mo</span
              >
            </div>

            <ul class="mb-6 flex flex-col gap-2.5 text-[13px] text-[var(--text-dim)]">
              <li v-for="st in plan.stats" :key="st.label" class="flex items-center gap-2">
                <svg aria-hidden="true" width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="var(--accent)" stroke-width="3">
                  <polyline points="20 6 9 17 4 12" />
                </svg>
                <span><strong class="font-semibold text-[var(--text-strong)]">{{ st.value }}</strong> {{ st.label }}</span>
              </li>
              <li v-for="e in plan.extras" :key="e" class="flex items-start gap-2 leading-relaxed">
                <svg aria-hidden="true" width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="var(--accent)" stroke-width="3" class="mt-[3px] flex-shrink-0">
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
              <img :src="t.avatar" :alt="t.name" class="h-9 w-9 flex-shrink-0 rounded-full object-cover" />
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
