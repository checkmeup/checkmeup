<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import { monitorsApi } from '@/api/monitors'
import { usePortMonitor } from '@/composables/usePortMonitors'

const router = useRouter()
const route = useRoute()
const id = route.params.id as string

const { data: detail, isPending: loading, error: queryError, refetch } = usePortMonitor(id)
const error = computed(() => queryError.value?.message ?? '')
const actionError = ref('')
const confirmDelete = ref(false)

async function togglePause() {
  if (!detail.value) return
  actionError.value = ''
  try {
    if (detail.value.monitor.status === 'paused') {
      await monitorsApi.resumePort(id)
    } else {
      await monitorsApi.pausePort(id)
    }
    await refetch()
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : 'Action failed'
  }
}

async function deleteMonitor() {
  actionError.value = ''
  try {
    await monitorsApi.deletePort(id)
    router.push({ name: 'port-monitors' })
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : 'Delete failed'
    confirmDelete.value = false
  }
}

const statusColors: Record<string, string> = {
  up: 'var(--status-up)',
  down: 'var(--status-down)',
  waiting: 'var(--text-muted)',
  paused: 'var(--status-paused)',
}

function fmtPct(v: number | null | undefined): string {
  if (v === null || v === undefined) return '—'
  return `${v.toFixed(2)}%`
}

function relativeTime(iso: string | null) {
  if (!iso) return '—'
  const diff = Date.now() - new Date(iso).getTime()
  const abs = Math.abs(diff)
  const m = Math.floor(abs / 60000)
  const h = Math.floor(m / 60)
  if (m < 1) return 'just now'
  if (m < 60) return `${m}m ago`
  if (h < 24) return `${h}h ago`
  return `${Math.floor(h / 24)}d ago`
}

function duration(start: string, end: string | null): string {
  if (!end) return 'Ongoing'
  const ms = new Date(end).getTime() - new Date(start).getTime()
  const mins = Math.floor(ms / 60000)
  if (mins < 60) return `${mins}m`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ${mins % 60}m`
  return `${Math.floor(hrs / 24)}d ${hrs % 24}h`
}

// SVG chart computed from chartData
const chart = computed(() => {
  const data = detail.value?.chartData ?? []
  if (data.length < 2) return null

  const W = 600
  const H = 80
  const pad = { top: 8, bottom: 8, left: 4, right: 4 }
  const innerW = W - pad.left - pad.right
  const innerH = H - pad.top - pad.bottom

  const times = data.map(c => new Date(c.checkedAt).getTime())
  const minT = Math.min(...times)
  const maxT = Math.max(...times)
  const tRange = maxT - minT || 1

  const rts = data.map(c => c.responseTimeMs)
  const maxRt = Math.max(...rts, 100)

  const points = data.map((c, i) => {
    const x = pad.left + ((times[i] - minT) / tRange) * innerW
    const y = pad.top + innerH - (c.responseTimeMs / maxRt) * innerH
    return { x, y, isUp: c.isUp }
  })

  const polyline = points.map(p => `${p.x},${p.y}`).join(' ')
  return { W, H, points, polyline }
})
</script>

<template>
  <AppLayout>
    <div class="p-8 max-w-4xl mx-auto">
      <div class="flex items-center gap-3 mb-6">
        <button
          class="text-sm transition-colors"
          style="color: var(--text-muted)"
          @click="router.push({ name: 'port-monitors' })"
        >
          ← Back
        </button>
        <h1 class="text-2xl font-semibold" style="color: var(--text-strong)">
          {{ detail?.monitor.name ?? 'Monitor' }}
        </h1>
      </div>

      <div v-if="loading" class="text-sm" style="color: var(--text-muted)">Loading…</div>
      <div v-else-if="error" class="text-sm" style="color: var(--status-down)">{{ error }}</div>

      <template v-else-if="detail">
        <!-- Header card -->
        <div
          class="rounded-xl border p-5 mb-6 flex items-start justify-between gap-4"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div class="space-y-1 min-w-0">
            <div class="flex items-center gap-2">
              <span
                class="inline-flex items-center gap-1.5 text-sm font-medium"
                :style="{ color: statusColors[detail.monitor.status] ?? 'var(--text-muted)' }"
              >
                <span class="w-2 h-2 rounded-full" :style="{ backgroundColor: statusColors[detail.monitor.status] }"></span>
                {{ detail.monitor.status.charAt(0).toUpperCase() + detail.monitor.status.slice(1) }}
              </span>
            </div>
            <p class="font-mono text-xs break-all" style="color: var(--text-dim)">
              {{ detail.monitor.host }}:{{ detail.monitor.port }}
            </p>
            <p class="text-xs" style="color: var(--text-muted)">
              Every {{ detail.monitor.intervalMins }} min
              · Last checked {{ relativeTime(detail.monitor.lastCheckedAt) }}
            </p>
            <p class="text-xs" style="color: var(--text-muted)">
              {{ detail.monitor.expectedState === 'closed' ? '🔒 Expected closed — alerts if reachable' : '🔓 Expected open — alerts if unreachable' }}
            </p>
          </div>
          <div class="flex items-center gap-2 flex-shrink-0">
            <Button
              variant="secondary"
              size="sm"
              @click="router.push({ name: 'port-monitor-edit', params: { id } })"
            >
              Edit
            </Button>
            <Button variant="secondary" size="sm" @click="togglePause">
              {{ detail.monitor.status === 'paused' ? 'Resume' : 'Pause' }}
            </Button>
            <Button
              v-if="!confirmDelete"
              variant="secondary"
              size="sm"
              style="color: var(--status-down)"
              @click="confirmDelete = true"
            >
              Delete
            </Button>
            <template v-else>
              <Button size="sm" style="background-color: var(--status-down)" @click="deleteMonitor">
                Confirm delete
              </Button>
              <Button variant="secondary" size="sm" @click="confirmDelete = false">Cancel</Button>
            </template>
          </div>
        </div>

        <p v-if="actionError" class="text-sm mb-4" style="color: var(--status-down)">{{ actionError }}</p>

        <!-- Uptime stats -->
        <div class="grid grid-cols-3 gap-4 mb-6">
          <div
            v-for="({ label, value }) in [
              { label: 'Uptime 24h', value: fmtPct(detail.stats.uptime24h) },
              { label: 'Uptime 7d', value: fmtPct(detail.stats.uptime7d) },
              { label: 'Uptime 30d', value: fmtPct(detail.stats.uptime30d) },
            ]"
            :key="label"
            class="rounded-xl border p-4 text-center"
            style="background-color: var(--surface); border-color: var(--border)"
          >
            <div class="text-xl font-bold" style="color: var(--text-strong)">{{ value }}</div>
            <div class="text-xs mt-1" style="color: var(--text-muted)">{{ label }}</div>
          </div>
        </div>

        <!-- Connect time chart -->
        <div
          v-if="chart"
          class="rounded-xl border p-5 mb-6"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <h2 class="text-sm font-medium mb-3" style="color: var(--text-strong)">Connect time — last 24h</h2>
          <svg
            :viewBox="`0 0 ${chart.W} ${chart.H}`"
            class="w-full"
            :height="chart.H"
            style="overflow: visible"
          >
            <polyline
              :points="chart.polyline"
              fill="none"
              stroke="var(--status-up)"
              stroke-width="1.5"
              stroke-linejoin="round"
              stroke-linecap="round"
            />
            <circle
              v-for="(p, i) in chart.points"
              :key="i"
              :cx="p.x"
              :cy="p.y"
              r="2.5"
              :fill="p.isUp ? 'var(--status-up)' : 'var(--status-down)'"
            />
          </svg>
          <p class="text-xs mt-1" style="color: var(--text-muted)">
            Red dots = failed checks. Y axis = connect time (max {{ Math.max(...(detail.chartData.map(c => c.responseTimeMs)), 0) }}ms).
          </p>
        </div>

        <div
          v-else-if="detail.chartData.length === 0"
          class="rounded-xl border p-5 mb-6 text-center"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <p class="text-sm" style="color: var(--text-muted)">No checks yet — first check runs within the configured interval.</p>
        </div>

        <!-- Incident log -->
        <div
          class="rounded-xl border mb-6"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div class="px-5 py-4 border-b" style="border-color: var(--border)">
            <h2 class="text-sm font-medium" style="color: var(--text-strong)">Incidents</h2>
          </div>
          <div v-if="detail.incidents.length === 0" class="px-5 py-4 text-sm" style="color: var(--text-muted)">
            No incidents recorded.
          </div>
          <table v-else class="w-full text-sm">
            <thead>
              <tr style="border-bottom: 1px solid var(--border)">
                <th class="text-left px-5 py-2 font-medium text-xs" style="color: var(--text-muted)">Started</th>
                <th class="text-left px-5 py-2 font-medium text-xs" style="color: var(--text-muted)">Resolved</th>
                <th class="text-left px-5 py-2 font-medium text-xs" style="color: var(--text-muted)">Duration</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="inc in detail.incidents"
                :key="inc.id"
                style="border-bottom: 1px solid var(--border)"
              >
                <td class="px-5 py-2.5" style="color: var(--text-dim)">{{ relativeTime(inc.startedAt) }}</td>
                <td class="px-5 py-2.5" style="color: var(--text-dim)">
                  <span v-if="inc.resolvedAt">{{ relativeTime(inc.resolvedAt) }}</span>
                  <span v-else class="text-xs font-medium" style="color: var(--status-down)">Ongoing</span>
                </td>
                <td class="px-5 py-2.5" style="color: var(--text-dim)">{{ duration(inc.startedAt, inc.resolvedAt) }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Check log -->
        <div
          class="rounded-xl border"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div class="px-5 py-4 border-b" style="border-color: var(--border)">
            <h2 class="text-sm font-medium" style="color: var(--text-strong)">Check log</h2>
          </div>
          <div v-if="detail.checks.length === 0" class="px-5 py-4 text-sm" style="color: var(--text-muted)">
            No checks yet.
          </div>
          <table v-else class="w-full text-sm">
            <thead>
              <tr style="border-bottom: 1px solid var(--border)">
                <th class="text-left px-5 py-2 font-medium text-xs" style="color: var(--text-muted)">Time</th>
                <th class="text-left px-5 py-2 font-medium text-xs" style="color: var(--text-muted)">Status</th>
                <th class="text-left px-5 py-2 font-medium text-xs" style="color: var(--text-muted)">Connect time</th>
                <th class="text-left px-5 py-2 font-medium text-xs" style="color: var(--text-muted)">Reason</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="c in detail.checks"
                :key="c.id"
                style="border-bottom: 1px solid var(--border)"
              >
                <td class="px-5 py-2.5" style="color: var(--text-dim)">{{ relativeTime(c.checkedAt) }}</td>
                <td class="px-5 py-2.5">
                  <span
                    class="text-xs font-medium"
                    :style="{ color: c.isUp ? 'var(--status-up)' : 'var(--status-down)' }"
                  >
                    {{ c.isUp ? '✓ Up' : '✗ Down' }}
                  </span>
                </td>
                <td class="px-5 py-2.5" style="color: var(--text-dim)">{{ c.responseTimeMs }}ms</td>
                <td class="px-5 py-2.5 text-xs" style="color: var(--text-dim)">{{ c.failureReason ?? '—' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </div>
  </AppLayout>
</template>
