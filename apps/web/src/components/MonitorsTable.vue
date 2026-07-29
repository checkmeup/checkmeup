<script setup lang="ts">
import { useRouter } from 'vue-router'
import Button from '@/components/ui/Button.vue'
import {
  type MonitorType,
  type Row,
  typeLabel,
  detailRouteName,
  statusColors,
  statusLabels,
} from '@/views/dashboardHelpers'

defineProps<{
  chips: { id: string; label: string; count: number }[]
  filter: 'all' | MonitorType
  hasAnyMonitors: boolean
  filteredRows: Row[]
}>()

const emit = defineEmits<{
  'update:filter': [value: 'all' | MonitorType]
}>()

const router = useRouter()
</script>

<template>
  <div
    class="flex-1 min-w-0 rounded-2xl border overflow-hidden"
    style="background-color: var(--surface); border-color: var(--border); flex-basis: 560px"
  >
    <div
      class="flex items-center justify-between gap-3 px-5 py-4 border-b flex-wrap"
      style="border-color: var(--border)"
    >
      <span class="text-sm font-semibold" style="color: var(--text-strong)">Monitors</span>
      <div class="flex gap-1.5 flex-wrap">
        <button
          v-for="chip in chips"
          :key="chip.id"
          class="px-3 py-1 rounded-full text-xs font-medium border transition-colors"
          :style="{
            borderColor: filter === chip.id ? 'var(--accent)' : 'var(--border)',
            backgroundColor: filter === chip.id ? 'var(--accent-wash)' : 'transparent',
            color: filter === chip.id ? 'var(--accent)' : 'var(--text-dim)',
          }"
          @click="emit('update:filter', chip.id as 'all' | MonitorType)"
        >
          {{ chip.label }} <span class="opacity-60">{{ chip.count }}</span>
        </button>
      </div>
    </div>

    <div v-if="!hasAnyMonitors" class="p-10 text-center">
      <p class="text-sm mb-4" style="color: var(--text-muted)">
        No monitors yet. Add a cron, uptime, SSL, domain, port, or DNS monitor to get started.
      </p>
      <Button @click="router.push({ name: 'uptime-monitor-create' })"
        >Add your first monitor</Button
      >
    </div>

    <template v-else>
      <!-- Mobile cards -->
      <div class="md:hidden">
        <div
          v-for="m in filteredRows"
          :key="m.key"
          class="p-4 border-b cursor-pointer transition-colors hover:bg-[var(--surface-raised)]"
          style="border-color: var(--border)"
          @click="router.push({ name: detailRouteName[m.type], params: { id: m.id } })"
        >
          <div class="flex items-center justify-between mb-1.5">
            <span class="font-medium text-sm truncate" style="color: var(--text-strong)">{{
              m.name
            }}</span>
            <span
              class="inline-flex items-center gap-1.5 text-xs font-medium flex-shrink-0"
              :style="{ color: statusColors[m.status] }"
            >
              <span
                class="w-1.5 h-1.5 rounded-full"
                :style="{ backgroundColor: statusColors[m.status] }"
              ></span>
              {{ statusLabels[m.status] ?? m.status }}
            </span>
          </div>
          <div class="flex items-center justify-between text-xs">
            <span class="font-mono truncate" style="color: var(--text-muted)">{{
              m.target
            }}</span>
            <span class="font-mono flex-shrink-0" style="color: var(--text-dim)">{{
              m.metricValue
            }}</span>
          </div>
        </div>
      </div>

      <!-- Desktop table -->
      <div class="hidden md:block overflow-x-auto">
        <table class="w-full text-sm" style="min-width: 640px">
          <thead>
            <tr style="background-color: var(--surface-raised)">
              <th class="w-6"></th>
              <th
                class="text-left px-3 py-2.5 text-[10.5px] font-semibold uppercase tracking-wider"
                style="color: var(--text-muted)"
              >
                Monitor
              </th>
              <th
                class="text-left px-3 py-2.5 text-[10.5px] font-semibold uppercase tracking-wider"
                style="color: var(--text-muted)"
              >
                Metric
              </th>
              <th
                class="text-right px-5 py-2.5 text-[10.5px] font-semibold uppercase tracking-wider"
                style="color: var(--text-muted)"
              >
                Last checked
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="m in filteredRows"
              :key="m.key"
              class="cursor-pointer transition-colors border-b hover:bg-[var(--surface-raised)]"
              style="border-color: var(--border)"
              @click="router.push({ name: detailRouteName[m.type], params: { id: m.id } })"
            >
              <td class="pl-5">
                <span
                  class="inline-block w-[9px] h-[9px] rounded-full"
                  :style="{ backgroundColor: statusColors[m.status] }"
                ></span>
              </td>
              <td class="px-3 py-3 min-w-0">
                <div class="flex items-center gap-1.5">
                  <span class="font-semibold truncate" style="color: var(--text-strong)">{{
                    m.name
                  }}</span>
                  <span
                    class="flex-shrink-0 px-1.5 py-0.5 rounded-full text-[10px] font-medium"
                    style="background-color: var(--surface-raised); color: var(--text-dim)"
                  >
                    {{ typeLabel[m.type] }}
                  </span>
                </div>
                <div class="font-mono text-[11.5px] truncate mt-0.5" style="color: var(--text-muted)">
                  {{ m.target }}
                </div>
              </td>
              <td class="px-3 py-3">
                <div class="text-[11.5px] mb-0.5" style="color: var(--text-muted)">
                  {{ m.metricLabel }}
                </div>
                <div
                  class="font-mono text-[13.5px] font-semibold"
                  :style="{
                    color: ['down', 'error', 'expired', 'expiring_soon'].includes(m.status)
                      ? statusColors[m.status]
                      : 'var(--text-strong)',
                  }"
                >
                  {{ m.metricValue }}
                </div>
              </td>
              <td class="px-5 py-3 font-mono text-[11.5px] text-right" style="color: var(--text-muted)">
                {{ m.lastChecked }}
              </td>
            </tr>
          </tbody>
        </table>
        <div
          v-if="filteredRows.length === 0"
          class="p-8 text-center text-sm"
          style="color: var(--text-muted)"
        >
          No monitors match this filter.
        </div>
      </div>
    </template>
  </div>
</template>
