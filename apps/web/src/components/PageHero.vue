<script setup lang="ts">
// The title block that opens a marketing or content page: left-aligned on the
// page grid, with an optional eyebrow above and a lead paragraph (default
// slot) below.
defineProps<{
  title: string
  eyebrow?: string
}>()
</script>

<template>
  <section class="mx-auto max-w-[1200px] px-4 pt-16 pb-10 sm:px-6 sm:pt-24 sm:pb-12">
    <p v-if="eyebrow" class="mb-4 text-[13px] font-medium text-[var(--accent)]">{{ eyebrow }}</p>
    <h1
      class="max-w-[900px] text-[44px] leading-[1.02] font-semibold tracking-[-0.045em] text-[var(--text-strong)] sm:text-[60px]"
    >
      {{ title }}
    </h1>
    <div
      v-if="$slots.default"
      class="mt-4 max-w-2xl text-lg leading-relaxed text-[var(--text-dim)]"
    >
      <slot />
    </div>
  </section>
</template>
