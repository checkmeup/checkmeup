<script setup lang="ts">
import { RouterLink } from 'vue-router'

defineProps<{
  message: string
}>()
</script>

<template>
  <div
    class="flex items-start gap-3 rounded-[14px] bg-[var(--accent-wash)] p-4"
  >
    <svg
      class="mt-0.5 flex-shrink-0 text-[var(--accent)]"
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
      aria-hidden="true"
    >
      <path d="M12 2a7 7 0 0 0-4 12.74V17a2 2 0 0 0 2 2h4a2 2 0 0 0 2-2v-2.26A7 7 0 0 0 12 2z" />
      <line x1="9" y1="22" x2="15" y2="22" />
    </svg>
    <div class="text-sm text-[var(--text)]">
      {{ message }}
      <RouterLink to="/billing" class="ml-1 font-semibold text-[var(--accent)] underline">
        View plans →
      </RouterLink>
    </div>
  </div>
</template>
