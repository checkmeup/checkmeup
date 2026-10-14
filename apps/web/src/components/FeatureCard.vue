<script setup lang="ts">
defineProps<{
  title: string
  description: string
  bullets: string[]
}>()
</script>

<template>
  <div class="flex flex-col border-t border-[var(--border)] pt-6">
    <div class="mb-3 text-[var(--accent)]" aria-hidden="true">
      <slot name="icon" />
    </div>
    <h3 class="mb-2 text-[17px] font-semibold tracking-[-0.015em] text-[var(--text-strong)]">
      {{ title }}
    </h3>
    <p class="text-[15px] leading-relaxed text-[var(--text-dim)]">
      {{ description }}
    </p>
    <ul class="mt-4 space-y-2">
      <li
        v-for="bullet in bullets"
        :key="bullet"
        class="flex items-center gap-2 text-[13px] text-[var(--text-dim)]"
      >
        <svg
          aria-hidden="true"
          width="12"
          height="12"
          viewBox="0 0 24 24"
          fill="none"
          stroke="var(--accent)"
          stroke-width="2.5"
        >
          <polyline points="20 6 9 17 4 12" />
        </svg>
        {{ bullet }}
      </li>
    </ul>
  </div>
</template>
