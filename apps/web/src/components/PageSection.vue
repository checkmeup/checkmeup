<script setup lang="ts">
// A titled block of a page: heading on the left, optional aside on the right
// (a hint, a link), content below. Sections are separated by whitespace, not
// card borders.
defineProps<{
  title: string
}>()
</script>

<template>
  <section class="flex flex-col gap-3.5">
    <div class="flex flex-wrap items-baseline justify-between gap-3">
      <h2 class="text-[22px] font-semibold tracking-[-0.02em] text-[var(--text-strong)]">
        {{ title }}
      </h2>
      <span v-if="$slots.aside" class="text-[13px] text-[var(--text-dim)]"><slot name="aside" /></span>
    </div>
    <slot />
  </section>
</template>
