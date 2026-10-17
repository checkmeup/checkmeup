<script setup lang="ts">
import { computed, ref } from 'vue'
import { useHead } from '@unhead/vue'
import LandingLayout from '@/layouts/LandingLayout.vue'
import CtaBanner from '@/components/CtaBanner.vue'
import HeroSection from '@/components/HeroSection.vue'
import CheckStrip from '@/components/CheckStrip.vue'
import { useSeo } from '@/composables/useSeo'
import {
  statusRows,
  plans,
  planRows,
  everyPlan,
  monitorTypes,
  ownStatusSlug,
  sampleTelegram,
  sampleWebhook,
} from '@/content/home'

// The live badge of checkmeup.net's own status page, and the snippet that
// embeds it, in whichever format the visitor picks.
const statusUrl = `https://checkmeup.net/status/${ownStatusSlug}`
const badgeUrl = `${statusUrl}/badge.svg`
const badgeFormat = ref<'markdown' | 'html'>('markdown')
const badgeSnippet = computed(() =>
  badgeFormat.value === 'markdown'
    ? `[![status](${badgeUrl})](${statusUrl})`
    : `<a href="${statusUrl}"><img src="${badgeUrl}" alt="status"></a>`,
)
const copied = ref(false)
let copiedTimer: ReturnType<typeof setTimeout> | undefined

function copyBadgeSnippet() {
  navigator.clipboard.writeText(badgeSnippet.value)
  copied.value = true
  clearTimeout(copiedTimer)
  copiedTimer = setTimeout(() => {
    copied.value = false
  }, 1500)
}

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
    'Cron, uptime, SSL, domain, port and DNS monitors with alerts and branded status pages — built for freelancers and agencies managing client sites.',
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

    <!-- Monitor types: a spec sheet, one row per type. Below md each row
         stacks into a card that keeps every column, labelled. -->
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

        <div class="rounded-[14px] border border-[var(--border)] md:overflow-x-auto">
          <table class="w-full border-collapse text-left text-sm max-md:block md:min-w-[760px]">
            <thead class="max-md:hidden">
              <tr class="bg-[var(--surface)] text-[13px] text-[var(--text-dim)]">
                <th scope="col" class="px-5 py-3 font-medium">Monitor</th>
                <th scope="col" class="px-5 py-3 font-medium">Watches</th>
                <th scope="col" class="px-5 py-3 font-medium">How often</th>
                <th scope="col" class="px-5 py-3 font-medium">Alerts you when</th>
              </tr>
            </thead>
            <tbody class="max-md:block">
              <tr
                v-for="m in monitorTypes"
                :id="m.id"
                :key="m.id"
                class="scroll-mt-24 border-[var(--border)] align-top max-md:flex max-md:flex-col max-md:gap-3 max-md:px-5 max-md:py-5 max-md:[&:not(:first-child)]:border-t md:border-t"
              >
                <th
                  scope="row"
                  class="font-semibold text-[var(--text-strong)] max-md:block md:px-5 md:py-4"
                >
                  {{ m.name }}
                </th>
                <td
                  data-label="Watches"
                  class="text-[var(--text-dim)] max-md:block md:px-5 md:py-4 max-md:before:mb-0.5 max-md:before:block max-md:before:text-xs max-md:before:text-[var(--text-muted)] max-md:before:content-[attr(data-label)]"
                >
                  {{ m.watches }}
                </td>
                <td
                  data-label="How often"
                  class="max-md:block md:px-5 md:py-4 md:whitespace-nowrap max-md:before:mb-0.5 max-md:before:block max-md:before:text-xs max-md:before:text-[var(--text-muted)] max-md:before:content-[attr(data-label)]"
                >
                  <span class="font-mono text-[13px] text-[var(--text)]">{{ m.often }}</span>
                  <span
                    v-if="m.oftenNote"
                    class="text-xs text-[var(--text-muted)] max-md:ml-2 md:block"
                    >{{ m.oftenNote }}</span
                  >
                </td>
                <td
                  data-label="Alerts you when"
                  class="text-[var(--text-dim)] max-md:block md:px-5 md:py-4 max-md:before:mb-0.5 max-md:before:block max-md:before:text-xs max-md:before:text-[var(--text-muted)] max-md:before:content-[attr(data-label)]"
                >
                  {{ m.alerts }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- What a missed ping actually sends: the real message and payload,
             not a description of them. -->
        <div class="mt-14 flex flex-col gap-5">
          <div class="flex max-w-[640px] flex-col gap-2">
            <h3 class="text-xl font-semibold tracking-[-0.02em] text-[var(--text-strong)]">
              What the 03:00 backup's alert looks like.
            </h3>
            <p class="text-[15px] text-[var(--text-dim)]">
              The same miss, as it reaches Telegram and a webhook. Email, Slack and SMS carry the
              same facts; SMS is on paid plans.
              <RouterLink to="/docs" class="text-[var(--accent)] hover:underline"
                >How each channel works →</RouterLink
              >
            </p>
          </div>
          <div class="grid min-w-0 items-start gap-4 lg:grid-cols-[1fr_1.4fr]">
            <!-- Telegram: drawn as the message itself, like the hero's alert. -->
            <figure class="flex min-w-0 flex-col gap-2">
              <figcaption class="font-mono text-xs text-[var(--text-muted)]">Telegram</figcaption>
              <div
                class="flex items-start gap-3.5 rounded-[14px] border border-[var(--border-strong)] bg-[var(--bg)] px-5 py-[18px] shadow-[0_18px_40px_rgb(0_0_0/7%)]"
              >
                <img src="/favicon.svg" alt="" class="h-9 w-9 flex-shrink-0" />
                <div class="flex min-w-0 flex-1 flex-col gap-2">
                  <div class="flex justify-between gap-3 text-[13px] text-[var(--text-muted)]">
                    <span class="font-semibold text-[var(--text-strong)]">Checkmeup</span>
                    <span class="font-mono">{{ sampleTelegram.time }}</span>
                  </div>
                  <div class="flex items-center gap-2 text-[15px] text-[var(--text-strong)]">
                    <span class="h-2 w-2 flex-shrink-0 rounded-full bg-[var(--status-down)]"></span>
                    <span
                      ><strong class="font-semibold">{{ sampleTelegram.name }}</strong>
                      {{ sampleTelegram.headline }}</span
                    >
                  </div>
                  <dl class="flex flex-col gap-1 text-sm">
                    <div v-for="d in sampleTelegram.details" :key="d.label" class="flex gap-1.5">
                      <dt class="text-[var(--text-dim)]">{{ d.label }}:</dt>
                      <dd
                        :class="
                          d.code
                            ? 'rounded bg-[var(--surface-raised)] px-1.5 font-mono text-[13px] text-[var(--text-strong)]'
                            : 'text-[var(--text)]'
                        "
                      >
                        {{ d.value }}
                      </dd>
                    </div>
                  </dl>
                </div>
              </div>
            </figure>
            <!-- Webhook: the request as sent, lightly highlighted. -->
            <figure class="flex min-w-0 flex-col gap-2">
              <figcaption class="font-mono text-xs text-[var(--text-muted)]">
                Webhook · signed with HMAC-SHA256 of the body
              </figcaption>
              <pre
                class="rounded-[14px] border border-[var(--border-strong)] bg-[var(--surface-raised)] px-5 py-4 font-mono text-[12.5px] leading-[1.7] whitespace-pre-wrap text-[var(--text-strong)] [overflow-wrap:anywhere]"
              ><span class="font-semibold">{{ sampleWebhook.requestLine }}</span>
<template v-for="h in sampleWebhook.headers" :key="h.name"><span class="text-[var(--text-dim)]">{{ h.name }}:</span> {{ h.value }}
</template>
{<template v-for="([key, value], i) in Object.entries(sampleWebhook.body)" :key="key">
  <span class="text-[var(--accent)]">"{{ key }}"</span>: "{{ value }}"{{ i < Object.keys(sampleWebhook.body).length - 1 ? ',' : '' }}</template>
}</pre>
            </figure>
          </div>
        </div>
      </div>
    </section>

    <!-- Status pages highlight -->
    <section class="border-t border-[var(--border)]">
      <div
        class="mx-auto grid max-w-[1200px] items-center gap-12 px-4 py-20 sm:px-6 lg:grid-cols-[1fr_1.1fr] lg:gap-16"
      >
        <div class="flex min-w-0 flex-col gap-5">
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

          <!-- The live badge of checkmeup.net's own page, and its embed snippet. -->
          <div class="mt-2 flex min-w-0 flex-col gap-3 border-t border-[var(--border)] pt-5">
            <p class="text-sm text-[var(--text-dim)]">
              Every page has a live badge for a README or a client's footer. This one is ours:
            </p>
            <a :href="statusUrl" class="self-start">
              <img :src="badgeUrl" alt="checkmeup.net status" />
            </a>
            <div class="flex min-w-0 flex-col overflow-hidden rounded-[10px] border border-[var(--border)]">
              <div
                class="flex items-center justify-between gap-2 border-b border-[var(--border)] bg-[var(--surface)] px-2 py-1.5"
              >
                <div class="flex gap-0.5" role="group" aria-label="Snippet format">
                  <button
                    v-for="f in ['markdown', 'html'] as const"
                    :key="f"
                    type="button"
                    class="rounded-md px-2.5 py-1 font-mono text-xs transition-colors hover:cursor-pointer"
                    :class="
                      badgeFormat === f
                        ? 'bg-[var(--bg)] text-[var(--text-strong)]'
                        : 'text-[var(--text-dim)] hover:text-[var(--text)]'
                    "
                    :aria-pressed="badgeFormat === f"
                    @click="badgeFormat = f"
                  >
                    {{ f === 'markdown' ? 'Markdown' : 'HTML' }}
                  </button>
                </div>
                <button
                  type="button"
                  class="rounded-md px-2.5 py-1 text-xs font-semibold text-[var(--accent)] transition-colors hover:cursor-pointer hover:bg-[var(--surface-raised)]"
                  @click="copyBadgeSnippet"
                >
                  {{ copied ? 'Copied!' : 'Copy' }}
                </button>
              </div>
              <code
                class="block px-3 py-2.5 font-mono text-xs leading-relaxed break-all text-[var(--text)]"
                >{{ badgeSnippet }}</code
              >
            </div>
          </div>
        </div>

        <!-- Status page mockup: mirrors the real public page's layout, with
             the kind of client brand and monitors a freelancer would show. -->
        <div
          class="flex flex-col gap-6 rounded-[18px] border border-[var(--border)] bg-[var(--bg)] p-6 shadow-[0_24px_60px_rgb(0_0_0/8%)] sm:p-8"
        >
          <span class="text-[17px] font-semibold tracking-[-0.03em] text-[var(--text-strong)]"
            >Acme Shop <span class="font-normal text-[var(--text-dim)]">status</span></span
          >
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
                <span class="font-mono text-sm font-medium text-[var(--text-strong)]">{{
                  row.name
                }}</span>
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

    <!-- Pricing: one table of what differs between plans; what doesn't is
         one line under it. -->
    <section class="border-t border-[var(--border)]">
      <div class="mx-auto max-w-[1200px] px-4 py-20 sm:px-6">
        <div class="mb-10 flex flex-wrap items-end justify-between gap-4">
          <div class="flex max-w-[640px] flex-col gap-3">
            <h2
              class="text-[30px] leading-[1.1] font-semibold tracking-[-0.035em] text-[var(--text-strong)] sm:text-[38px]"
            >
              Pay for monitors, not seats.
            </h2>
            <p class="text-[17px] text-[var(--text-dim)]">
              Hobby is free for 10 monitors, no card. Paid plans check every minute and add status
              pages, channels and SMS.
            </p>
          </div>
          <RouterLink
            to="/pricing"
            class="text-sm font-semibold text-[var(--accent)] hover:underline"
          >
            See full pricing details →
          </RouterLink>
        </div>

        <div class="overflow-x-auto rounded-[14px] border border-[var(--border)]">
          <!-- Fixed layout so the four plan columns are equal, whatever their
               descriptions' lengths; the label column takes a set share. On
               phones the label column collapses and each label gets its own
               full-width line above its values, so the plan columns get the
               whole width ("Enterprise" overflows a quarter of what's left). -->
          <table class="w-full table-fixed border-collapse text-left text-[13px] sm:text-sm">
            <colgroup>
              <col class="w-0 sm:w-[24%]" />
              <col v-for="plan in plans" :key="plan.name" />
            </colgroup>
            <thead>
              <tr class="align-bottom">
                <th scope="col" class="px-2 py-4 max-sm:hidden sm:px-5">
                  <span class="sr-only">Plan</span>
                </th>
                <th
                  v-for="plan in plans"
                  :key="plan.name"
                  scope="col"
                  class="px-2 py-4 font-normal sm:px-5"
                  :class="
                    plan.highlight ? 'bg-[color-mix(in_srgb,var(--accent)_6%,var(--bg))]' : ''
                  "
                >
                  <span
                    v-if="plan.highlight"
                    class="mb-1 block text-xs font-semibold text-[var(--accent)]"
                    >Most popular</span
                  >
                  <span class="block text-xs font-semibold text-[var(--text-strong)] sm:text-sm">{{
                    plan.name
                  }}</span>
                  <span
                    class="block text-[18px] font-semibold tracking-[-0.035em] text-[var(--text-strong)] sm:text-[28px]"
                  >
                    {{ plan.price === 0 ? 'Free' : `$${plan.price}`
                    }}<span
                      v-if="plan.price > 0"
                      class="text-xs font-normal tracking-normal text-[var(--text-dim)]"
                      >/mo</span
                    >
                  </span>
                  <span class="block text-xs text-[var(--text-dim)] max-sm:hidden">{{
                    plan.description
                  }}</span>
                </th>
              </tr>
            </thead>
            <tbody>
              <template v-for="row in planRows" :key="row.label">
              <tr class="border-t border-[var(--border)] sm:hidden">
                <td colspan="5" class="px-2 pt-3 text-xs text-[var(--text-dim)]">
                  {{ row.label }}
                </td>
              </tr>
              <tr class="sm:border-t sm:border-[var(--border)]">
                <th
                  scope="row"
                  class="px-2 py-3 font-normal text-[var(--text-dim)] max-sm:hidden sm:px-5"
                >
                  {{ row.label }}
                </th>
                <td
                  v-for="(value, i) in row.values"
                  :key="i"
                  class="px-2 py-3 font-mono text-[var(--text-strong)] sm:px-5"
                  :class="
                    plans[i].highlight
                      ? 'bg-[color-mix(in_srgb,var(--accent)_6%,var(--bg))]'
                      : ''
                  "
                >
                  {{ value }}
                </td>
              </tr>
              </template>
              <tr class="border-t border-[var(--border)] max-sm:hidden">
                <td class="px-5 py-4"></td>
                <td
                  v-for="plan in plans"
                  :key="plan.name"
                  class="px-5 py-4"
                  :class="
                    plan.highlight ? 'bg-[color-mix(in_srgb,var(--accent)_6%,var(--bg))]' : ''
                  "
                >
                  <RouterLink
                    to="/sign-up"
                    class="block rounded-[10px] px-3 py-2 text-center text-sm font-semibold whitespace-nowrap transition-colors"
                    :class="
                      plan.highlight
                        ? 'bg-[var(--accent)] text-[var(--on-accent)] hover:bg-[var(--accent-hover)]'
                        : 'border border-[var(--border-strong)] text-[var(--text-strong)] hover:bg-[var(--surface-raised)]'
                    "
                  >
                    {{ plan.cta }}
                  </RouterLink>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <p class="mt-4 text-[13px] text-[var(--text-muted)]">{{ everyPlan }}</p>
      </div>
    </section>

    <CtaBanner
      title="Find out whether tonight's backup ran."
      subtitle="Free for 10 monitors, no credit card. Paste the curl line and you're done."
    />
  </LandingLayout>
</template>
