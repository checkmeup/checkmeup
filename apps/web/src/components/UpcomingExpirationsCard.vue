<script setup lang="ts">
import { useRouter } from 'vue-router'
import { type MonitorType, detailRouteName, typeLabel } from '@/views/dashboardHelpers'

interface ExpiryRow {
  key: string
  type: MonitorType
  id: string
  name: string
  daysUntilExpiry: number
  color: string
}

defineProps<{
  hasSslOrDomainMonitors: boolean
  expiryRows: ExpiryRow[]
}>()

const router = useRouter()
</script>

<template>
  <div
    class="rounded-2xl border p-5"
    style="
      flex: 1 1 320px;
      min-width: 300px;
      background-color: var(--surface);
      border-color: var(--border);
    "
  >
    <div class="text-sm font-semibold mb-4" style="color: var(--text-strong)">
      Upcoming expirations
    </div>

    <p v-if="!hasSslOrDomainMonitors" class="text-xs" style="color: var(--text-muted)">
      Add an SSL or domain monitor to track certificate and registration expirations here.
    </p>
    <p v-else-if="expiryRows.length === 0" class="text-xs" style="color: var(--text-muted)">
      No expiry data yet — checks run shortly after a monitor is created.
    </p>
    <div v-else class="flex flex-col gap-3">
      <div
        v-for="row in expiryRows"
        :key="row.key"
        class="flex items-center justify-between gap-3 cursor-pointer"
        @click="router.push({ name: detailRouteName[row.type], params: { id: row.id } })"
      >
        <div class="min-w-0">
          <div class="text-[13px] font-medium truncate" style="color: var(--text-strong)">
            {{ row.name }}
          </div>
          <div class="text-[11px] mt-0.5" style="color: var(--text-muted)">
            {{ typeLabel[row.type] }}
          </div>
        </div>
        <div class="font-mono text-xs font-semibold flex-shrink-0" :style="{ color: row.color }">
          {{ row.daysUntilExpiry < 0 ? 'Expired' : `${row.daysUntilExpiry}d` }}
        </div>
      </div>
    </div>
  </div>
</template>
