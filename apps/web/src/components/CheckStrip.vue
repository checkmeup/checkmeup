<script setup lang="ts">
import { computed } from 'vue'

interface Check {
  ok: boolean | null
  slow?: boolean
  label?: string
}

// One tick per check, oldest on the left: the run of recent results at a
// glance. `ok: null` is an empty slot (no check there), and a `label` becomes
// the tick's tooltip. `narrowLast` keeps only the newest that many ticks on
// phone-width screens, where 90 ticks would be hairlines.
const props = withDefaults(
  defineProps<{
    checks: Check[]
    size?: 'sm' | 'lg'
    narrowLast?: number
  }>(),
  { size: 'lg', narrowLast: undefined },
)

const failed = computed(() => props.checks.filter((c) => c.ok === false).length)
const total = computed(() => props.checks.filter((c) => c.ok !== null).length)
const summary = computed(() =>
  failed.value === 0
    ? `Last ${total.value} checks all passed`
    : `${failed.value} of the last ${total.value} checks failed`,
)

function color(c: Check) {
  if (c.ok === null) return 'var(--tick-empty)'
  if (!c.ok) return 'var(--status-down)'
  return c.slow ? 'var(--status-degraded)' : 'var(--tick-up)'
}

function hiddenWhenNarrow(i: number) {
  return props.narrowLast !== undefined && i < props.checks.length - props.narrowLast
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
      :title="c.label"
      class="min-w-px flex-1 rounded-[2px]"
      :class="{ 'max-sm:hidden': hiddenWhenNarrow(i) }"
      :style="{ backgroundColor: color(c) }"
    ></span>
  </div>
</template>
