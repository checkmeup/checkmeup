<script setup lang="ts">
import {
  type MonitorRow,
  attentionSentence,
  detailRouteName,
  toneColor,
} from '@/lib/monitorRows'

defineProps<{
  rows: MonitorRow[]
}>()

function actionLabel(row: MonitorRow) {
  return row.status === 'expired' || row.status === 'expiring_soon' ? 'Renew' : 'Open'
}
</script>

<template>
  <section v-if="rows.length > 0" aria-labelledby="attention-heading" class="flex flex-col">
    <h2 id="attention-heading" class="mb-3 text-[13px] font-semibold text-[var(--text-dim)]">
      Needs attention
    </h2>
    <RouterLink
      v-for="row in rows"
      :key="row.key"
      :to="{ name: detailRouteName[row.type], params: { id: row.id } }"
      class="group flex flex-wrap items-center gap-x-6 gap-y-2 border-t border-[var(--border)] py-5 focus-visible:outline-none"
    >
      <span
        class="h-3 w-3 flex-shrink-0 rounded-full"
        :style="{
          backgroundColor: toneColor[row.tone],
          boxShadow: `0 0 0 5px color-mix(in srgb, ${toneColor[row.tone]} 16%, transparent)`,
        }"
      ></span>
      <span class="flex min-w-0 flex-[1_1_280px] flex-col gap-0.5">
        <span
          class="text-[19px] font-semibold tracking-[-0.015em] text-[var(--text-strong)] group-hover:underline group-focus-visible:underline"
        >
          {{ row.name }}
        </span>
        <span class="text-[14.5px]" :style="{ color: toneColor[row.tone] }">
          {{ attentionSentence(row) }}
        </span>
      </span>
      <span class="min-w-[80px] text-right text-sm font-semibold text-[var(--accent)]">
        {{ actionLabel(row) }} →
      </span>
    </RouterLink>
  </section>
</template>
