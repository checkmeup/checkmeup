<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import UpgradePrompt from '@/components/UpgradePrompt.vue'
import MonitorDetailHeader from '@/components/MonitorDetailHeader.vue'
import PageSection from '@/components/PageSection.vue'
import StatGrid from '@/components/StatGrid.vue'
import CheckStrip from '@/components/CheckStrip.vue'
import ResponseChart from '@/components/ResponseChart.vue'
import DataTable from '@/components/DataTable.vue'
import StatusLabel from '@/components/StatusLabel.vue'
import { monitorsApi } from '@/api/monitors'
import { useDNSMonitor } from '@/composables/useDNSMonitors'
import { useMonitorActions } from '@/composables/useMonitorActions'
import {
  fmtPct,
  incidentDuration,
  monitorsListRoute,
  relativeTime,
  statusLabel,
  toneOf,
} from '@/lib/monitorRows'

const router = useRouter()
const route = useRoute()
const id = route.params.id as string

const { data: detail, isPending: loading, error: queryError, refetch } = useDNSMonitor(id)
const error = computed(() => queryError.value?.message ?? '')

const { actionError, limitReached, confirmDelete, togglePause, deleteMonitor } = useMonitorActions({
  isPaused: () => detail.value?.monitor.status === 'paused',
  pause: () => monitorsApi.pauseDns(id),
  resume: () => monitorsApi.resumeDns(id),
  remove: () => monitorsApi.deleteDns(id),
  refetch,
  afterDelete: () => router.push({ name: 'dns-monitors' }),
})

const MAX_TICKS = 90
const recent = computed(() => (detail.value?.chartData ?? []).slice(-MAX_TICKS))
const strip = computed(() => recent.value.map((c) => ({ ok: c.isUp })))
const chartPoints = computed(() => recent.value.map((c) => ({ ms: c.responseTimeMs, ok: c.isUp })))

const avgLookup = computed(() => {
  const ok = recent.value.filter((c) => c.isUp)
  if (ok.length === 0) return '—'
  return `${Math.round(ok.reduce((a, c) => a + c.responseTimeMs, 0) / ok.length)} ms`
})

const summary = computed(() => {
  const d = detail.value
  if (!d) return ''
  if (d.monitor.status === 'down') {
    const failing = d.checks.find((c) => !c.isUp)
    if (failing?.failureReason) return `Down — the lookup failed: ${failing.failureReason}.`
    if (failing?.resolvedValue) return `Down — the record now resolves to ${failing.resolvedValue}.`
    return 'Down — the record no longer matches.'
  }
  if (d.monitor.status === 'paused') return 'Paused — no checks run until you resume it.'
  if (d.monitor.status === 'waiting') return 'Waiting for the first check.'
  return 'Up — the record matches.'
})
</script>

<template>
  <AppLayout>
    <div class="mx-auto flex max-w-[1200px] flex-col gap-11 px-4 pt-8 pb-20 md:px-6">
      <p v-if="loading" class="text-sm text-[var(--text-dim)]">Loading…</p>
      <p v-else-if="error" class="text-sm text-[var(--status-down)]">{{ error }}</p>

      <template v-else-if="detail">
        <MonitorDetailHeader
          v-model:confirm-delete="confirmDelete"
          :back-to="monitorsListRoute('dns')"
          back-label="DNS monitors"
          :name="detail.monitor.name"
          :tone="toneOf(detail.monitor.status)"
          :status-label="statusLabel('dns', detail.monitor.status)"
          :paused="detail.monitor.status === 'paused'"
          @edit="router.push({ name: 'dns-monitor-edit', params: { id } })"
          @toggle-pause="togglePause"
          @delete="deleteMonitor"
        >
          <p class="text-lg text-[var(--text-dim)]">{{ summary }}</p>
          <div class="mt-1 flex flex-col gap-1 text-[13px] text-[var(--text-dim)]">
            <p class="font-mono break-all">
              {{ detail.monitor.hostname }} ({{ detail.monitor.recordType }})
            </p>
            <p>
              Every {{ detail.monitor.intervalMins }} min · Last checked
              {{ relativeTime(detail.monitor.lastCheckedAt) }}
            </p>
            <p>
              {{ detail.monitor.baselineCaptured ? 'Baseline' : 'Expected value' }}:
              <span class="font-mono">{{
                detail.monitor.expectedValue ?? '(capturing on next check)'
              }}</span>
            </p>
            <p v-if="detail.monitor.lastResolvedValue">
              Current value: <span class="font-mono">{{ detail.monitor.lastResolvedValue }}</span>
            </p>
          </div>
        </MonitorDetailHeader>

        <UpgradePrompt v-if="limitReached" :message="actionError" />
        <p v-else-if="actionError" class="text-sm text-[var(--status-down)]">{{ actionError }}</p>

        <section v-if="strip.length > 0" aria-label="Recent checks" class="flex flex-col gap-2.5">
          <CheckStrip :checks="strip" />
          <div class="flex justify-between text-[13px] text-[var(--text-dim)]">
            <span>{{ relativeTime(recent[0]!.checkedAt) }}</span>
            <span>last {{ strip.length }} checks</span>
            <span>now</span>
          </div>
        </section>

        <StatGrid
          :items="[
            { label: 'Uptime 24h', value: fmtPct(detail.stats.uptime24h) },
            { label: 'Uptime 7d', value: fmtPct(detail.stats.uptime7d) },
            { label: 'Uptime 30d', value: fmtPct(detail.stats.uptime30d) },
            { label: 'Average lookup', value: avgLookup },
          ]"
        />

        <PageSection title="Lookup time">
          <template v-if="chartPoints.length >= 2" #aside
            >failed checks (mismatch or lookup error) shown in red</template
          >
          <ResponseChart v-if="chartPoints.length >= 2" :points="chartPoints" />
          <p v-else-if="detail.chartData.length === 0" class="text-sm text-[var(--text-dim)]">
            No checks yet — first check runs within the configured interval.
          </p>
        </PageSection>

        <PageSection title="Incidents">
          <DataTable
            :columns="[
              { key: 'started', label: 'Started' },
              { key: 'resolved', label: 'Resolved' },
              { key: 'duration', label: 'Duration' },
            ]"
            :rows="detail.incidents"
            :row-key="(i) => i.id"
            empty="No incidents recorded."
          >
            <template #started="{ row }">{{ relativeTime(row.startedAt) }}</template>
            <template #resolved="{ row }">
              <span v-if="row.resolvedAt">{{ relativeTime(row.resolvedAt) }}</span>
              <StatusLabel v-else tone="down" label="Ongoing" />
            </template>
            <template #duration="{ row }">
              <span class="font-mono">{{ incidentDuration(row.startedAt, row.resolvedAt) }}</span>
            </template>
          </DataTable>
        </PageSection>

        <PageSection title="Check log">
          <DataTable
            :columns="[
              { key: 'time', label: 'Time' },
              { key: 'status', label: 'Status' },
              { key: 'lookup', label: 'Lookup time' },
              { key: 'value', label: 'Resolved value' },
              { key: 'error', label: 'Error' },
            ]"
            :rows="detail.checks"
            :row-key="(c) => c.id"
            empty="No checks yet."
          >
            <template #time="{ row }">
              <span class="font-mono text-[var(--text-dim)]">{{ relativeTime(row.checkedAt) }}</span>
            </template>
            <template #status="{ row }">
              <StatusLabel :tone="row.isUp ? 'up' : 'down'" :label="row.isUp ? 'Up' : 'Down'" />
            </template>
            <template #lookup="{ row }">
              <span class="font-mono">{{ row.responseTimeMs }} ms</span>
            </template>
            <template #value="{ row }">
              <span class="font-mono">{{ row.resolvedValue ?? '—' }}</span>
            </template>
            <template #error="{ row }">
              <span class="text-[var(--text-dim)]">{{ row.failureReason ?? '—' }}</span>
            </template>
          </DataTable>
        </PageSection>
      </template>
    </div>
  </AppLayout>
</template>
