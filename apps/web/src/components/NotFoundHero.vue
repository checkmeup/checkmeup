<script setup lang="ts">
import { RouterLink } from 'vue-router'

interface Cta {
  label: string
  to: string
}

defineProps<{
  badge: string
  heading: string
  description: string
  primaryCta: Cta
  secondaryCta: Cta
}>()
</script>

<template>
  <!-- Shared between the site-wide 404 route and BlogPostView's unknown-slug state. -->
  <section class="flex items-center justify-center py-24 sm:py-32">
    <div class="max-w-[520px] px-4 text-center sm:px-6">
      <p class="mb-6 inline-flex items-center gap-2 text-[13px] font-medium text-[var(--status-down)]">
        <span class="h-2 w-2 flex-shrink-0 rounded-full bg-[var(--status-down)]"></span>
        {{ badge }}
      </p>

      <div
        class="mb-2 font-mono text-8xl leading-none font-medium tracking-[-0.06em] text-[var(--text-strong)] sm:text-9xl"
      >
        404
      </div>

      <h1 class="text-[30px] leading-[1.1] sm:text-[38px] font-semibold tracking-[-0.035em] mb-4 text-[var(--text-strong)]">
        {{ heading }}
      </h1>

      <p class="text-[15px] leading-relaxed mb-8 text-[var(--text-dim)]">
        {{ description }}
      </p>

      <div class="flex items-center justify-center gap-2.5 flex-wrap">
        <RouterLink
          :to="primaryCta.to"
          class="inline-flex items-center rounded-[10px] bg-[var(--accent)] px-5 py-2.5 text-sm font-semibold text-[var(--on-accent)] transition-colors hover:bg-[var(--accent-hover)]"
        >
          {{ primaryCta.label }}
        </RouterLink>
        <RouterLink
          :to="secondaryCta.to"
          class="inline-flex items-center rounded-[10px] border border-[var(--border-strong)] px-5 py-2.5 text-sm font-medium text-[var(--text)] transition-colors hover:bg-[var(--surface-raised)]"
        >
          {{ secondaryCta.label }}
        </RouterLink>
      </div>
    </div>
  </section>
</template>
