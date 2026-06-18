<script setup lang="ts">
import { RouterLink } from 'vue-router'

defineProps<{
  message: string
}>()
</script>

<template>
  <div
    class="rounded-lg border p-4 flex items-start gap-3"
    style="background-color: var(--surface-raised); border-color: var(--color-green-500)"
  >
    <svg
      class="flex-shrink-0 mt-0.5"
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
      style="color: var(--color-green-500)"
    >
      <path d="M12 2a7 7 0 0 0-4 12.74V17a2 2 0 0 0 2 2h4a2 2 0 0 0 2-2v-2.26A7 7 0 0 0 12 2z" />
      <line x1="9" y1="22" x2="15" y2="22" />
    </svg>
    <div class="text-sm" style="color: var(--text)">
      {{ message }}
      <RouterLink to="/billing" class="font-medium underline ml-1" style="color: var(--color-green-500)">
        View plans →
      </RouterLink>
    </div>
  </div>
</template>
