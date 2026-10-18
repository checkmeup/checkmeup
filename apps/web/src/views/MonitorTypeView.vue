<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import LandingLayout from '@/layouts/LandingLayout.vue'
import PageHero from '@/components/PageHero.vue'
import LabeledSection from '@/components/LabeledSection.vue'
import PlanTable from '@/components/PlanTable.vue'
import CtaBanner from '@/components/CtaBanner.vue'
import { useSeo } from '@/composables/useSeo'
import { monitorPages, type MonitorTypeId } from '@/content/monitorPages'
import { everyPlan, monitorTypes } from '@/content/home'

// One product page per monitor type (PRD marketing-pages); the route passes
// which one.
const props = defineProps<{ type: MonitorTypeId }>()

const page = computed(() => monitorPages.find((p) => p.id === props.type)!)
const spec = computed(() => monitorTypes.find((m) => m.id === props.type)!)

useSeo({
  title: () => page.value.seoTitle,
  description: () => page.value.description,
  path: () => page.value.path,
})
</script>

<template>
  <LandingLayout>
    <PageHero :title="page.title" eyebrow="Monitoring · Free for 10 monitors">
      <p>{{ page.lead }}</p>
      <div class="mt-8 flex flex-wrap items-center gap-4">
        <RouterLink
          to="/sign-up"
          class="rounded-[10px] bg-[var(--accent)] px-6 py-2.5 text-[15px] font-semibold text-[var(--on-accent)] transition-colors hover:bg-[var(--accent-hover)]"
        >
          Start monitoring free →
        </RouterLink>
        <RouterLink to="/docs" class="text-sm font-semibold text-[var(--accent)] hover:underline">
          Read the docs
        </RouterLink>
      </div>
    </PageHero>

    <div class="mx-auto max-w-[1200px] px-4 pb-16 sm:px-6">
      <LabeledSection title="At a glance">
        <dl class="grid gap-4 text-[15px] sm:grid-cols-3">
          <div>
            <dt class="text-xs text-[var(--text-muted)]">Watches</dt>
            <dd class="mt-1 text-[var(--text-strong)]">{{ spec.watches }}</dd>
          </div>
          <div>
            <dt class="text-xs text-[var(--text-muted)]">How often</dt>
            <dd class="mt-1 text-[var(--text-strong)]">
              <span class="font-mono">{{ spec.often }}</span>
              <span v-if="spec.oftenNote" class="text-[var(--text-dim)]">
                · {{ spec.oftenNote }}</span
              >
            </dd>
          </div>
          <div>
            <dt class="text-xs text-[var(--text-muted)]">Alerts when</dt>
            <dd class="mt-1 text-[var(--text-strong)]">{{ spec.alerts }}</dd>
          </div>
        </dl>
      </LabeledSection>

      <LabeledSection title="How it works">
        <div class="max-w-[680px] space-y-4 text-[15px] leading-relaxed text-[var(--text-dim)]">
          <p v-for="(paragraph, i) in page.how" :key="i">{{ paragraph }}</p>
        </div>
      </LabeledSection>

      <LabeledSection title="Alerts">
        <p class="max-w-[680px] text-[15px] leading-relaxed text-[var(--text-dim)]">
          Down and recovery alerts go to email, Telegram, Slack or a signed webhook on every plan,
          and by SMS on paid plans. A monitor with nowhere else to alert emails everyone in your
          account, so an alert is never lost.
        </p>
      </LabeledSection>

      <LabeledSection title="What it doesn't tell you">
        <p class="max-w-[680px] text-[15px] leading-relaxed text-[var(--text-dim)]">
          {{ page.limits }}
        </p>
      </LabeledSection>

      <LabeledSection title="Read more">
        <ul class="space-y-2 text-[15px]">
          <li v-for="link in page.related" :key="link.to">
            <RouterLink :to="link.to" class="text-[var(--accent)] hover:underline">{{
              link.label
            }}</RouterLink>
          </li>
        </ul>
      </LabeledSection>
    </div>

    <section class="border-t border-[var(--border)]">
      <div class="mx-auto max-w-[1200px] px-4 py-16 sm:px-6">
        <h2
          class="mb-3 text-[30px] leading-[1.1] font-semibold tracking-[-0.035em] text-[var(--text-strong)]"
        >
          Plans
        </h2>
        <p class="mb-8 max-w-[680px] text-[15px] text-[var(--text-dim)]">{{ everyPlan }}</p>
        <PlanTable />
      </div>
    </section>

    <CtaBanner
      title="Start with 10 free monitors."
      subtitle="No credit card, no expiry date on the free plan."
    />
  </LandingLayout>
</template>
