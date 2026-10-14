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
import { useUptimeMonitor } from '@/composables/useUptimeMonitors'
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

const { data: detail, isPending: loading, error: queryError, refetch } = useUptimeMonitor(id)
const error = computed(() => queryError.value?.message ?? '')
const { actionError, limitReached, confirmDelete, togglePause, deleteMonitor } = useMonitorActions({
  isPaused: () => detail.value?.monitor.status === 'paused',
  pause: () => monitorsApi.pauseUptime(id),
  resume: () => monitorsApi.resumeUptime(id),
  remove: () => monitorsApi.deleteUptime(id),
  refetch,
  afterDelete: () => router.push({ name: 'uptime-monitors' }),
})

const MAX_TICKS = 90

// chartData comes oldest-first over the last 24h; the strip and chart show
// the most recent checks of it.
const recent = computed(() => (detail.value?.chartData ?? []).slice(-MAX_TICKS))

const strip = computed(() => {
  const max = detail.value?.monitor.maxResponseTimeMs ?? 0
  return recent.value.map((c) => ({
    ok: c.isUp,
    slow: c.isUp && max > 0 && c.responseTimeMs > max / 2,
  }))
})

const chartPoints = computed(() => recent.value.map((c) => ({ ms: c.responseTimeMs, ok: c.isUp })))

const avgResponse = computed(() => {
  const ok = recent.value.filter((c) => c.isUp)
  if (ok.length === 0) return '—'
  return `${Math.round(ok.reduce((a, c) => a + c.responseTimeMs, 0) / ok.length)} ms`
})

// What's going on, in words — the latest failing check's reason while down.
const summary = computed(() => {
  const d = detail.value
  if (!d) return ''
  if (d.monitor.status === 'down') {
    const failing = d.checks.find((c) => !c.isUp)
    const why = failing?.failureReason ?? (failing?.statusCode ? `HTTP ${failing.statusCode}` : '')
    return why ? `Down — the last check failed: ${why}.` : 'Down — the last check failed.'
  }
  if (d.monitor.status === 'paused') return 'Paused — no checks run until you resume it.'
  if (d.monitor.status === 'waiting') return 'Waiting for the first check.'
  return 'Up — the last check passed.'
})

const nonDefaultCodes = computed(() => {
  const codes = detail.value?.monitor.acceptedStatusCodes ?? []
  return !(codes.length === 1 && codes[0] === 200)
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
          :back-to="monitorsListRoute('uptime')"
          back-label="Uptime monitors"
          :name="detail.monitor.name"
          :tone="toneOf(detail.monitor.status)"
          :status-label="statusLabel('uptime', detail.monitor.status)"
          :paused="detail.monitor.status === 'paused'"
          @edit="router.push({ name: 'uptime-monitor-edit', params: { id } })"
          @toggle-pause="togglePause"
          @delete="deleteMonitor"
        >
          <p class="text-lg text-[var(--text-dim)]">{{ summary }}</p>
          <div class="mt-1 flex flex-col gap-1 text-[13px] text-[var(--text-dim)]">
            <p class="font-mono break-all">
              {{ detail.monitor.httpMethod }} {{ detail.monitor.url }}
            </p>
            <p>
              Every {{ detail.monitor.intervalMins }} min · Last checked
              {{ relativeTime(detail.monitor.lastCheckedAt) }} · Timeout
              {{ detail.monitor.maxResponseTimeMs }} ms
            </p>
            <p v-if="detail.monitor.keyword">
              {{ detail.monitor.keywordMode === 'not_contains' ? 'Does not contain' : 'Contains' }}
              "{{ detail.monitor.keyword }}"
              <span v-if="detail.monitor.keywordCaseSensitive">(case-sensitive)</span>
            </p>
            <p v-for="(a, i) in detail.monitor.jsonAssertions" :key="i">
              JSON <code class="font-mono">{{ a.path }}</code>
              {{ a.comparator.replace('_', ' ') }}
              <code class="font-mono">{{ a.expected }}</code>
            </p>
            <p v-if="nonDefaultCodes">
              Accepts status {{ detail.monitor.acceptedStatusCodes.join(', ') }}
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
            { label: 'Average response', value: avgResponse },
          ]"
        />

        <PageSection title="Response time">
          <template v-if="chartPoints.length >= 2" #aside>failed checks shown in red</template>
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
              { key: 'code', label: 'Code' },
              { key: 'response', label: 'Response' },
              { key: 'reason', label: 'Reason' },
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
            <template #code="{ row }">
              <span class="font-mono">{{ row.statusCode ?? '—' }}</span>
            </template>
            <template #response="{ row }">
              <span class="font-mono">{{ row.responseTimeMs }} ms</span>
            </template>
            <template #reason="{ row }">
              <span class="text-[var(--text-dim)]">{{ row.failureReason ?? '—' }}</span>
            </template>
          </DataTable>
        </PageSection>
      </template>
    </div>
  </AppLayout>
</template>
