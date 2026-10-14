<script setup lang="ts">
import { computed } from 'vue'
import { type MonitorRow, detailRouteName, fmtDaysLeft, typeLabel } from '@/lib/monitorRows'

const props = defineProps<{
  rows: MonitorRow[]
}>()

// SSL and domain rows with a known expiry, soonest first.
const expiring = computed(() =>
  props.rows
    .filter((r) => (r.type === 'ssl' || r.type === 'domain') && r.daysUntilExpiry !== null)
    .sort((a, b) => a.daysUntilExpiry! - b.daysUntilExpiry!)
    .slice(0, 8),
)

const hasSslOrDomain = computed(() => props.rows.some((r) => r.type === 'ssl' || r.type === 'domain'))

function color(days: number) {
  if (days < 7) return 'var(--status-down)'
  if (days < 30) return 'var(--status-degraded)'
  return 'var(--text)'
}
</script>

<template>
  <section aria-labelledby="expirations-heading" class="flex flex-col">
    <h2
      id="expirations-heading"
      class="mb-3 text-[22px] font-semibold tracking-[-0.02em] text-[var(--text-strong)]"
    >
      Upcoming expirations
    </h2>
    <p v-if="!hasSslOrDomain" class="text-sm text-[var(--text-dim)]">
      Add an SSL or domain monitor to track certificate and registration expirations here.
    </p>
    <p v-else-if="expiring.length === 0" class="text-sm text-[var(--text-dim)]">
      No expiry data yet — checks run shortly after a monitor is created.
    </p>
    <template v-else>
    <RouterLink
      v-for="row in expiring"
      :key="row.key"
      :to="{ name: detailRouteName[row.type], params: { id: row.id } }"
      class="flex items-center justify-between gap-4 border-t border-[var(--border)] py-3 hover:bg-[var(--surface)] focus-visible:bg-[var(--surface)] focus-visible:outline-none"
    >
      <span class="min-w-0 truncate">
        <span class="font-medium text-[var(--text-strong)]">{{ row.name }}</span>
        <span class="text-[13.5px] text-[var(--text-dim)]"> · {{ typeLabel[row.type] }}</span>
      </span>
      <span
        class="flex-shrink-0 font-mono text-[13.5px]"
        :style="{ color: color(row.daysUntilExpiry!) }"
      >
        {{ fmtDaysLeft(row.daysUntilExpiry) }}
      </span>
    </RouterLink>
    </template>
  </section>
</template>
