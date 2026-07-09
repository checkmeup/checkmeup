<script setup lang="ts">
defineProps<{
  title: string
  description: string
  bullets: string[]
}>()
</script>

<template>
  <div
    class="rounded-2xl border p-6"
    style="background-color: var(--card); border-color: var(--border)"
  >
    <div
      class="w-10 h-10 rounded-lg flex items-center justify-center mb-4"
      style="background-color: var(--accent-wash)"
    >
      <slot name="icon" />
    </div>
    <h3 class="font-semibold text-base mb-2" style="color: var(--text-strong)">
      {{ title }}
    </h3>
    <p class="text-sm leading-relaxed" style="color: var(--text-dim)">
      {{ description }}
    </p>
    <ul class="mt-4 space-y-2">
      <li
        v-for="bullet in bullets"
        :key="bullet"
        class="flex items-center gap-2 text-xs"
        style="color: var(--text-dim)"
      >
        <svg
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
