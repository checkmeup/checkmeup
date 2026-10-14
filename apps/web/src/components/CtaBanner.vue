<script setup lang="ts">
import { RouterLink } from 'vue-router'

// The closing call to action on marketing pages: a big left-aligned headline
// beside the sign-up button, between hairlines.
defineProps<{
  title: string
  subtitle: string
}>()
</script>

<template>
  <section class="border-t border-[var(--border)]">
    <div
      class="mx-auto flex max-w-[1200px] flex-wrap items-end justify-between gap-8 px-4 py-20 sm:px-6"
    >
      <div class="flex flex-col gap-3">
        <h2
          class="text-[34px] leading-[1.05] font-semibold tracking-[-0.04em] text-[var(--text-strong)] sm:text-[48px]"
        >
          {{ title }}
        </h2>
        <p class="text-[17px] text-[var(--text-dim)]">{{ subtitle }}</p>
      </div>
      <RouterLink
        to="/sign-up"
        class="rounded-[10px] bg-[var(--accent)] px-7 py-3 text-[15px] font-semibold text-[var(--on-accent)] transition-colors hover:bg-[var(--accent-hover)]"
      >
        Create free account →
      </RouterLink>
    </div>
  </section>
</template>
