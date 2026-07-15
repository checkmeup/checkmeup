<script setup lang="ts">
import { computed } from 'vue'
import { useHead } from '@unhead/vue'
import LandingLayout from '@/layouts/LandingLayout.vue'
import HeroSection from '@/components/HeroSection.vue'
import FeatureCard from '@/components/FeatureCard.vue'
import { useTheme } from '@/lib/theme'
import { useSeo } from '@/composables/useSeo'
import logoDark from '@/assets/logo-dark.svg'
import logoLight from '@/assets/logo-light.svg'
import { statusRows, plans, customers, testimonials } from '@/content/home'

const { theme } = useTheme()
const logo = computed(() => (theme.value === 'light' ? logoLight : logoDark))

useSeo({
  title: 'Checkmeup — monitoring for freelancers & client sites',
  description:
    'Cron, uptime, SSL, domain expiry, and port (TCP) monitors with execution logs, Telegram alerts, and branded status pages — built for freelancers and solo devs managing client sites.',
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
    'Cron, uptime, SSL, domain expiry, and port (TCP) monitors with execution logs, Telegram alerts, and branded status pages — built for freelancers and solo devs managing client sites.',
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
    <section class="border-y" style="border-color: var(--border)">
      <div class="max-w-[1100px] mx-auto px-4 sm:px-6 py-10">
        <p
          class="text-center text-xs uppercase tracking-widest mb-6"
          style="color: var(--text-muted); font-family: var(--font-mono)"
        >
          Trusted by freelancers and solo devs worldwide
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
              class="w-7 h-7 rounded-full object-cover"
              style="box-shadow: 0 0 0 2px var(--bg)"
            />
          </div>

          <!-- Count + label -->
          <div class="text-center sm:text-left">
            <span class="text-sm font-semibold" style="color: var(--text-strong)">200+</span>
            <span class="text-sm ml-1" style="color: var(--text-dim)"
              >freelancers monitoring client sites with Checkmeup</span
            >
          </div>
        </div>

        <!-- Customer type pills -->
        <div class="flex flex-wrap justify-center gap-2 mt-6">
          <span
            v-for="tag in [
              'Freelancers',
              'Solo agency operators',
              'Web agencies',
              'Solo developers',
              'SaaS startups',
            ]"
            :key="tag"
            class="text-xs px-3 py-1 rounded-full border"
            style="border-color: var(--border); color: var(--text-muted)"
          >
            {{ tag }}
          </span>
        </div>
      </div>
    </section>

    <!-- Features -->
    <section class="max-w-[1100px] mx-auto px-4 sm:px-6 py-16 sm:py-24">
      <div class="text-center mb-12">
        <h2
          class="text-2xl sm:text-3xl font-extrabold mb-3 tracking-tight"
          style="color: var(--text-strong)"
        >
          Everything you need to stay on top.
        </h2>
        <p class="text-base" style="color: var(--text-dim)">
          Five monitor types, one dashboard, zero blind spots.
        </p>
      </div>

      <div class="grid sm:grid-cols-2 lg:grid-cols-3 gap-3">
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
    <section class="max-w-[1100px] mx-auto px-4 sm:px-6 py-16 sm:py-20">
      <div
        class="rounded-[20px] border p-8 sm:p-12 grid lg:grid-cols-[1fr_1.1fr] gap-10 lg:gap-[52px] items-center"
        style="background-color: var(--card); border-color: var(--border)"
      >
        <div class="flex-1">
          <div
            class="inline-flex items-center gap-1.5 text-xs font-medium px-2.5 py-1 rounded-full mb-4"
            style="
              border: 1px solid var(--border);
              background-color: var(--surface);
              color: var(--accent-emphasis);
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
          <h2
            class="text-2xl sm:text-3xl font-extrabold mb-4 tracking-tight"
            style="color: var(--text-strong)"
          >
            Give your clients a page they can bookmark.
          </h2>
          <p class="text-base leading-relaxed mb-6" style="color: var(--text-muted)">
            Create a branded status page for every client site. They see uptime, incidents, and
            maintenance windows — you stay in control, and it looks like your work.
          </p>
          <RouterLink
            to="/sign-up"
            class="inline-flex items-center gap-2 text-sm font-medium px-5 py-2.5 rounded-md transition-opacity hover:opacity-90"
            style="background-color: var(--accent); color: var(--on-accent)"
          >
            Create a status page →
          </RouterLink>

          <div
            class="mt-6 p-4 rounded-lg border"
            style="border-color: var(--border); background-color: var(--surface)"
          >
            <p class="text-sm mb-2.5" style="color: var(--text-muted)">
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
          class="w-full rounded-xl border overflow-hidden text-sm"
          style="border-color: var(--border)"
        >
          <div
            class="flex items-center px-4 py-3 border-b"
            style="border-color: var(--border); background-color: var(--surface)"
          >
            <img :src="logo" alt="Checkmeup" class="h-[15px] w-auto" />
          </div>
          <div
            class="px-4 py-2.5 border-b"
            style="border-color: var(--border); background-color: var(--accent-wash-dim)"
          >
            <span class="text-xs font-semibold" style="color: var(--accent)"
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
                style="background-color: var(--accent-wash); color: var(--accent)"
              >
                Operational
              </span>
            </div>
            <div class="h-1.5 rounded overflow-hidden" style="background-color: var(--accent-wash)">
              <div
                class="h-full rounded"
                :style="{ width: row.pct, backgroundColor: 'var(--accent)' }"
              ></div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Pricing teaser -->
    <section class="border-t" style="border-color: var(--border)">
      <div class="max-w-[1100px] mx-auto px-4 sm:px-6 py-16 sm:py-20 text-center">
        <h2
          class="text-2xl sm:text-3xl font-extrabold mb-3 tracking-tight"
          style="color: var(--text-strong)"
        >
          Simple pricing.
        </h2>
        <p class="mb-10" style="color: var(--text-dim)">Start free. Scale as you grow.</p>

        <div class="grid sm:grid-cols-2 lg:grid-cols-4 gap-3">
          <div
            v-for="plan in plans"
            :key="plan.name"
            class="rounded-2xl border p-6 text-left relative flex flex-col"
            :style="{
              backgroundColor: plan.highlight ? 'var(--accent-wash-dim)' : 'var(--card)',
              borderColor: plan.highlight ? 'var(--accent)' : 'var(--border)',
            }"
          >
            <div
              v-if="plan.highlight"
              class="absolute -top-3 left-1/2 -translate-x-1/2 text-xs font-semibold px-3 py-0.5 rounded-full"
              style="background-color: var(--accent); color: var(--on-accent)"
            >
              Most popular
            </div>
            <div class="text-sm font-semibold mb-1" style="color: var(--text-strong)">
              {{ plan.name }}
            </div>
            <p class="text-xs mb-4" style="color: var(--text-muted)">{{ plan.description }}</p>
            <div class="text-2xl font-extrabold mb-5" style="color: var(--text-strong)">
              {{ plan.price === 0 ? 'Free' : `$${plan.price}` }}
              <span
                v-if="plan.price > 0"
                class="text-sm font-normal"
                style="color: var(--text-muted)"
                >/mo</span
              >
            </div>

            <div
              class="pt-4 mb-5 border-t flex flex-col gap-2.5"
              style="border-color: var(--border)"
            >
              <div
                v-for="s in plan.stats"
                :key="s.label"
                class="flex items-center gap-2 text-xs"
                style="color: var(--text-dim)"
              >
                <svg
                  width="11"
                  height="11"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="var(--accent)"
                  stroke-width="3"
                >
                  <polyline points="20 6 9 17 4 12" />
                </svg>
                <strong style="color: var(--text-strong)">{{ s.value }}</strong
                >&nbsp;{{ s.label }}
              </div>
              <div
                v-for="e in plan.extras"
                :key="e"
                class="flex items-start gap-2 text-xs leading-relaxed"
                style="color: var(--text-dim)"
              >
                <svg
                  width="11"
                  height="11"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="var(--accent)"
                  stroke-width="3"
                  style="margin-top: 3px; flex-shrink: 0"
                >
                  <polyline points="20 6 9 17 4 12" />
                </svg>
                {{ e }}
              </div>
            </div>

            <RouterLink
              to="/sign-up"
              class="mt-auto text-center text-xs font-medium px-4 py-2.5 rounded-md transition-opacity hover:opacity-90"
              :style="
                plan.highlight
                  ? 'background-color: var(--accent); color: var(--on-accent)'
                  : 'border: 1px solid var(--border); color: var(--text-strong)'
              "
            >
              {{ plan.cta }}
            </RouterLink>
          </div>
        </div>

        <RouterLink
          to="/pricing"
          class="inline-flex items-center gap-1 mt-8 text-sm transition-opacity hover:opacity-80"
          style="color: var(--accent-emphasis)"
        >
          See full pricing details →
        </RouterLink>
      </div>
    </section>

    <!-- Testimonials -->
    <section class="border-t" style="border-color: var(--border)">
      <div class="max-w-[1100px] mx-auto px-4 sm:px-6 py-16 sm:py-20">
        <div class="text-center mb-12">
          <h2
            class="text-2xl sm:text-3xl font-extrabold mb-3 tracking-tight"
            style="color: var(--text-strong)"
          >
            Loved by freelancers and solo devs.
          </h2>
          <p class="text-base" style="color: var(--text-dim)">
            Monitoring that gets out of your way — so you can focus on client work.
          </p>
        </div>

        <div class="grid sm:grid-cols-3 gap-4">
          <div
            v-for="t in testimonials"
            :key="t.name"
            class="rounded-2xl border p-6 flex flex-col gap-4"
            style="background-color: var(--card); border-color: var(--border)"
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
                style="color: var(--accent)"
              >
                <polygon
                  points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"
                />
              </svg>
            </div>
            <p class="text-sm leading-relaxed" style="color: var(--text-dim)">"{{ t.quote }}"</p>
          </div>
        </div>
      </div>
    </section>

    <!-- CTA banner -->
    <section class="border-t" style="border-color: var(--border)">
      <div class="max-w-[1100px] mx-auto px-4 sm:px-6 py-16 sm:py-20">
        <div
          class="rounded-[20px] text-center px-10 py-[72px] border"
          style="background-color: var(--accent-wash-dim); border-color: var(--cta-border)"
        >
          <h2
            class="text-2xl sm:text-3xl font-extrabold mb-3 tracking-tight"
            style="color: var(--text-strong)"
          >
            Start monitoring in 60 seconds.
          </h2>
          <p class="mb-8 text-base" style="color: var(--text-dim)">
            Free plan included. No credit card required.
          </p>
          <RouterLink
            to="/sign-up"
            class="inline-flex items-center gap-2 text-sm font-semibold px-7 py-3 rounded-md transition-opacity hover:opacity-90"
            style="background-color: var(--accent); color: var(--on-accent)"
          >
            Create free account →
          </RouterLink>
        </div>
      </div>
    </section>
  </LandingLayout>
</template>
