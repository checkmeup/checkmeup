<script setup lang="ts">
import { computed } from 'vue'

// One tick per check, oldest on the left: the run of recent results at a
// glance. `ok: null` is an empty slot (no check there).
const props = withDefaults(
  defineProps<{
    checks: { ok: boolean | null; slow?: boolean }[]
    size?: 'sm' | 'lg'
  }>(),
  { size: 'lg' },
)

const failed = computed(() => props.checks.filter((c) => c.ok === false).length)
const total = computed(() => props.checks.filter((c) => c.ok !== null).length)
const summary = computed(() =>
  failed.value === 0
    ? `Last ${total.value} checks all passed`
    : `${failed.value} of the last ${total.value} checks failed`,
)

function color(c: { ok: boolean | null; slow?: boolean }) {
  if (c.ok === null) return 'var(--tick-empty)'
  if (!c.ok) return 'var(--status-down)'
  return c.slow ? 'var(--status-degraded)' : 'var(--tick-up)'
}
</script>

<template>
  <div
    role="img"
    :aria-label="summary"
    class="flex"
    :class="size === 'lg' ? 'h-11 gap-[3px]' : 'h-[18px] gap-0.5'"
  >
    <span
      v-for="(c, i) in checks"
      :key="i"
      class="min-w-px flex-1 rounded-[2px]"
      :style="{ backgroundColor: color(c) }"
    ></span>
  </div>
</template>
