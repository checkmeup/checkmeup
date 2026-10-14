<script setup lang="ts">
import { computed } from 'vue'

// Response time per check as bars, oldest on the left; a failed check is a
// full-height red bar so outages read as gaps in the pattern.
const props = defineProps<{
  points: { ms: number; ok: boolean }[]
}>()

const maxMs = computed(() => {
  const top = Math.max(100, ...props.points.filter((p) => p.ok).map((p) => p.ms))
  // Round the scale up to a readable step.
  const step = top <= 500 ? 100 : top <= 2000 ? 500 : 1000
  return Math.ceil(top / step) * step
})

const bars = computed(() =>
  props.points.map((p) => ({
    height: p.ok ? `${Math.max(2, (p.ms / maxMs.value) * 100)}%` : '100%',
    color: p.ok ? 'var(--accent)' : 'color-mix(in srgb, var(--status-down) 35%, transparent)',
    label: p.ok ? `${p.ms} ms` : 'failed',
  })),
)
</script>

<template>
  <div class="flex gap-3">
    <div
      aria-hidden="true"
      class="flex h-[180px] w-10 flex-col justify-between py-3 text-right font-mono text-[11px] text-[var(--text-muted)]"
    >
      <span>{{ maxMs }}</span>
      <span>{{ maxMs / 2 }}</span>
      <span>0</span>
    </div>
    <div
      role="img"
      :aria-label="`Response time for the last ${points.length} checks, up to ${maxMs} ms`"
      class="flex h-[180px] min-w-0 flex-1 items-end gap-[2px] rounded-[14px] bg-[var(--surface)] p-3"
    >
      <span
        v-for="(b, i) in bars"
        :key="i"
        :title="b.label"
        class="min-w-px flex-1 rounded-t-[2px] opacity-80"
        :style="{ height: b.height, backgroundColor: b.color }"
      ></span>
    </div>
  </div>
</template>
