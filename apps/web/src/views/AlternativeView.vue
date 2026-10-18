<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import LandingLayout from '@/layouts/LandingLayout.vue'
import PageHero from '@/components/PageHero.vue'
import LabeledSection from '@/components/LabeledSection.vue'
import CtaBanner from '@/components/CtaBanner.vue'
import { useSeo } from '@/composables/useSeo'
import { alternatives, type AlternativeSlug } from '@/content/alternatives'

// One "<tool> alternative" page per competitor (PRD marketing-pages); the
// route passes which one. Their facts are dated and sourced on the page.
const props = defineProps<{ slug: AlternativeSlug }>()

const page = computed(() => alternatives.find((a) => a.slug === props.slug)!)
const asOf = computed(() =>
  new Date(`${page.value.asOf}T00:00:00Z`).toLocaleDateString('en-US', {
    month: 'long',
    day: 'numeric',
    year: 'numeric',
    timeZone: 'UTC',
  }),
)

useSeo({
  title: () => page.value.seoTitle,
  description: () => page.value.description,
  path: () => `/alternatives/${page.value.slug}`,
})
</script>

<template>
  <LandingLayout>
    <PageHero :title="page.title" :eyebrow="`${page.name} alternative`">
      <p>{{ page.lead }}</p>
      <div class="mt-8 flex flex-wrap items-center gap-4">
        <RouterLink
          to="/sign-up"
          class="rounded-[10px] bg-[var(--accent)] px-6 py-2.5 text-[15px] font-semibold text-[var(--on-accent)] transition-colors hover:bg-[var(--accent-hover)]"
        >
          Try Checkmeup free →
        </RouterLink>
        <RouterLink
          to="/pricing"
          class="text-sm font-semibold text-[var(--accent)] hover:underline"
        >
          See pricing
        </RouterLink>
      </div>
    </PageHero>

    <div class="mx-auto max-w-[1200px] px-4 pb-16 sm:px-6">
      <LabeledSection title="Side by side">
        <template #description>As of {{ asOf }}, from each product's own pages.</template>
        <div class="overflow-x-auto rounded-[14px] border border-[var(--border)]">
          <table class="w-full border-collapse text-left text-sm">
            <thead>
              <tr class="bg-[var(--surface)] text-[13px] text-[var(--text-dim)]">
                <th scope="col" class="px-4 py-3 font-medium">
                  <span class="sr-only">Feature</span>
                </th>
                <th scope="col" class="px-4 py-3 font-medium text-[var(--text-strong)]">
                  Checkmeup
                </th>
                <th scope="col" class="px-4 py-3 font-medium">{{ page.name }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="row in page.rows"
                :key="row.label"
                class="border-t border-[var(--border)] align-top"
              >
                <th scope="row" class="px-4 py-3 font-normal text-[var(--text-dim)]">
                  {{ row.label }}
                </th>
                <td class="px-4 py-3 text-[var(--text-strong)]">{{ row.checkmeup }}</td>
                <td class="px-4 py-3 text-[var(--text)]">{{ row.them }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </LabeledSection>

      <LabeledSection title="Why switch">
        <ul
          class="max-w-[680px] list-disc space-y-2 pl-5 text-[15px] leading-relaxed text-[var(--text-dim)]"
        >
          <li v-for="(reason, i) in page.reasons" :key="i">{{ reason }}</li>
        </ul>
      </LabeledSection>

      <LabeledSection :title="`When ${page.name} is the better choice`">
        <ul
          class="max-w-[680px] list-disc space-y-2 pl-5 text-[15px] leading-relaxed text-[var(--text-dim)]"
        >
          <li v-for="(reason, i) in page.theyWin" :key="i">{{ reason }}</li>
        </ul>
      </LabeledSection>

      <LabeledSection title="Switching">
        <ol
          class="max-w-[680px] list-decimal space-y-2 pl-5 text-[15px] leading-relaxed text-[var(--text-dim)]"
        >
          <li v-for="(step, i) in page.switching" :key="i">{{ step }}</li>
        </ol>
      </LabeledSection>

      <LabeledSection title="Sources">
        <template #description
          >Prices and plans change; check the source before you decide.</template
        >
        <ul class="space-y-1.5 text-sm">
          <li v-for="source in page.sources" :key="source.url">
            <a
              :href="source.url"
              rel="nofollow noopener"
              class="text-[var(--accent)] hover:underline"
              >{{ source.label }}</a
            >
          </li>
        </ul>
      </LabeledSection>
    </div>

    <CtaBanner
      title="Start with 10 free monitors."
      subtitle="No credit card, no expiry date on the free plan."
    />
  </LandingLayout>
</template>
