<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import UpgradePrompt from '@/components/UpgradePrompt.vue'
import MonitorDetailHeader from '@/components/MonitorDetailHeader.vue'
import PageSection from '@/components/PageSection.vue'
import StatGrid from '@/components/StatGrid.vue'
import DataTable from '@/components/DataTable.vue'
import StatusLabel from '@/components/StatusLabel.vue'
import { monitorsApi } from '@/api/monitors'
import { useCronMonitor } from '@/composables/useCronMonitors'
import { useMonitorActions } from '@/composables/useMonitorActions'
import {
  incidentDuration,
  monitorsListRoute,
  relativeTime,
  statusLabel,
  toneOf,
} from '@/lib/monitorRows'

const router = useRouter()
const route = useRoute()
const id = route.params.id as string

const { data: detail, isPending: loading, error: queryError, refetch } = useCronMonitor(id)
const error = computed(() => queryError.value?.message ?? '')
const monitor = computed(() => detail.value?.monitor)

const { actionError, limitReached, confirmDelete, togglePause, deleteMonitor } = useMonitorActions({
  isPaused: () => monitor.value?.status === 'paused',
  pause: () => monitorsApi.pauseCron(id),
  resume: () => monitorsApi.resumeCron(id),
  remove: () => monitorsApi.deleteCron(id),
  refetch,
  afterDelete: () => router.push({ name: 'cron-monitors' }),
})

function fmt(iso: string | null | undefined) {
  if (!iso) return '—'
  return new Date(iso).toLocaleString()
}

const summary = computed(() => {
  const m = monitor.value
  if (!m) return ''
  switch (m.status) {
    case 'down':
      return `Missed its expected ping — the last one came ${relativeTime(m.lastPingAt)}.`
    case 'paused':
      return 'Paused — missed pings raise no alerts until you resume it.'
    case 'waiting':
      return 'Waiting for the first ping — add the ping URL below to your job.'
    default:
      return m.nextPingAt
        ? `On schedule — last ping ${relativeTime(m.lastPingAt)}, next expected ${relativeTime(m.nextPingAt)}.`
        : `On schedule — last ping ${relativeTime(m.lastPingAt)}.`
  }
})

function copyPingUrl() {
  if (monitor.value?.pingUrl) {
    navigator.clipboard.writeText(monitor.value.pingUrl)
  }
}
</script>

<template>
  <AppLayout>
    <div class="mx-auto flex max-w-[1200px] flex-col gap-11 px-4 pt-8 pb-20 md:px-6">
      <p v-if="loading" class="text-sm text-[var(--text-dim)]">Loading…</p>
      <p v-else-if="error" class="text-sm text-[var(--status-down)]">{{ error }}</p>

      <template v-else-if="monitor">
        <MonitorDetailHeader
          v-model:confirm-delete="confirmDelete"
          :back-to="monitorsListRoute('cron')"
          back-label="Cron monitors"
          :name="monitor.name"
          :tone="toneOf(monitor.status)"
          :status-label="statusLabel('cron', monitor.status)"
          :paused="monitor.status === 'paused'"
          @edit="router.push({ name: 'cron-monitor-edit', params: { id } })"
          @toggle-pause="togglePause"
          @delete="deleteMonitor"
        >
          <p class="text-lg text-[var(--text-dim)]">{{ summary }}</p>
          <p v-if="confirmDelete" class="text-sm text-[var(--status-down)]">
            Deleting removes all ping history, and the ping URL stops working immediately.
          </p>
        </MonitorDetailHeader>

        <UpgradePrompt v-if="limitReached" :message="actionError" />
        <p v-else-if="actionError" class="text-sm text-[var(--status-down)]">{{ actionError }}</p>

        <StatGrid
          :items="[
            { label: 'Schedule', value: monitor.schedule },
            { label: 'Grace period', value: `${monitor.gracePeriodMins} min` },
            { label: 'Last ping', value: relativeTime(monitor.lastPingAt) },
            { label: 'Next expected', value: relativeTime(monitor.nextPingAt) },
          ]"
        />

        <PageSection title="Ping URL">
          <template #aside>call it when the job finishes</template>
          <div class="flex flex-wrap items-center gap-2">
            <code
              class="min-w-0 flex-1 truncate rounded-[10px] bg-[var(--surface)] px-4 py-2.5 font-mono text-[13px] text-[var(--text)]"
            >
              {{ monitor.pingUrl }}
            </code>
            <Button variant="secondary" @click="copyPingUrl">Copy</Button>
          </div>
        </PageSection>

        <PageSection v-if="detail!.incidents.length > 0" title="Incidents">
          <DataTable
            :columns="[
              { key: 'started', label: 'Started' },
              { key: 'outcome', label: 'Outcome' },
            ]"
            :rows="detail!.incidents"
            :row-key="(i) => i.id"
            empty=""
          >
            <template #started="{ row }">{{ fmt(row.startedAt) }}</template>
            <template #outcome="{ row }">
              <span v-if="row.resolvedAt" class="text-[var(--text-dim)]">
                Resolved after {{ incidentDuration(row.startedAt, row.resolvedAt) }}
              </span>
              <StatusLabel v-else tone="down" label="Ongoing" />
            </template>
          </DataTable>
        </PageSection>

        <PageSection title="Execution log">
          <DataTable
            :columns="[
              { key: 'received', label: 'Received' },
              { key: 'duration', label: 'Run time' },
              { key: 'source', label: 'Source IP' },
              { key: 'metadata', label: 'Metadata' },
            ]"
            :rows="detail!.pings"
            :row-key="(p) => p.id"
            empty="No pings received yet. Add the ping URL to your cron job."
          >
            <template #received="{ row }">{{ fmt(row.receivedAt) }}</template>
            <template #duration="{ row }">
              <span
                class="font-mono"
                :title="row.runStartedAt ? 'Start ping to completion ping' : 'No start ping received'"
              >
                {{ row.runStartedAt ? incidentDuration(row.runStartedAt, row.receivedAt) : '—' }}
              </span>
            </template>
            <template #source="{ row }">
              <span class="font-mono text-[var(--text-dim)]">{{ row.sourceIp || '—' }}</span>
            </template>
            <template #metadata="{ row }">
              <div v-if="row.metadata" class="flex flex-wrap gap-1.5">
                <span
                  v-for="(value, key) in row.metadata"
                  :key="key"
                  :title="`${key}: ${value}`"
                  class="inline-block max-w-[220px] truncate rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-xs text-[var(--text-dim)]"
                >
                  {{ key }}: {{ value }}
                </span>
              </div>
              <span v-else class="text-[var(--text-muted)]">—</span>
            </template>
          </DataTable>
        </PageSection>
      </template>
    </div>
  </AppLayout>
</template>
