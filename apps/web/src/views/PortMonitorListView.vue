<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import type { PortMonitor } from '@/api/monitors'
import { usePortMonitors } from '@/composables/usePortMonitors'

const router = useRouter()
const { data, isPending: loading, error: queryError, refetch } = usePortMonitors()
const monitors = computed<PortMonitor[]>(() => data.value ?? [])
const error = computed(() => queryError.value?.message ?? '')

function load() {
  refetch()
}

const statusColors: Record<string, string> = {
  up: 'var(--status-up)',
  down: 'var(--status-down)',
  waiting: 'var(--text-muted)',
  paused: 'var(--status-paused)',
}

function statusLabel(s: string) {
  return s.charAt(0).toUpperCase() + s.slice(1)
}

// magnitude renders the m/h difference as a duration ("5m", "3h", "2d"),
// independent of direction (past/future) — kept separate from relativeTime
// so its complexity doesn't compound with the direction handling.
function magnitude(m: number, h: number): string {
  if (m < 60) return `${m}m`
  if (h < 24) return `${h}h`
  return `${Math.floor(h / 24)}d`
}

function relativeTime(iso: string | null) {
  if (!iso) return '—'
  const diff = Date.now() - new Date(iso).getTime()
  const abs = Math.abs(diff)
  const m = Math.floor(abs / 60000)
  const h = Math.floor(m / 60)
  if (diff >= 0) return m < 1 ? 'just now' : `${magnitude(m, h)} ago`
  return m < 1 ? 'in <1m' : `in ${magnitude(m, h)}`
}

function fmtPct(v: number | null): string {
  if (v === null || v === undefined) return '—'
  return `${v.toFixed(2)}%`
}

function hostPort(m: PortMonitor) {
  return `${m.host}:${m.port}`
}
</script>

<template>
  <AppLayout>
    <div class="p-4 md:p-8 max-w-4xl mx-auto">
      <div class="flex items-center justify-between mb-6">
        <h1 class="text-2xl font-semibold" style="color: var(--text-strong)">Port monitors</h1>
        <Button @click="router.push({ name: 'port-monitor-create' })">
          Add monitor
        </Button>
      </div>

      <div v-if="loading" class="text-sm" style="color: var(--text-muted)">Loading…</div>

      <div v-else-if="error" class="rounded-xl border p-6 text-center" style="background-color: var(--surface); border-color: var(--border)">
        <p class="text-sm mb-4" style="color: var(--status-down)">{{ error }}</p>
        <Button variant="secondary" size="sm" @click="load">Try again</Button>
      </div>

      <div
        v-else-if="monitors.length === 0"
        class="rounded-xl border p-12 text-center"
        style="background-color: var(--surface); border-color: var(--border)"
      >
        <p class="text-sm mb-4" style="color: var(--text-muted)">
          No port monitors yet. Add one to watch a non-HTTP service (mail, database, custom daemon).
        </p>
        <Button @click="router.push({ name: 'port-monitor-create' })">
          Add your first monitor
        </Button>
      </div>

      <template v-else>
        <!-- Mobile cards -->
        <div class="md:hidden space-y-2">
          <div
            v-for="m in monitors"
            :key="m.id"
            class="rounded-xl border p-4 cursor-pointer transition-colors hover:border-[var(--color-green-700)]"
            style="background-color: var(--surface); border-color: var(--border)"
            @click="router.push({ name: 'port-monitor-detail', params: { id: m.id } })"
          >
            <div class="flex items-center justify-between mb-2">
              <span class="font-medium text-sm" style="color: var(--text-strong)">{{ m.name }}</span>
              <span
                class="inline-flex items-center gap-1.5 text-xs font-medium"
                :style="{ color: statusColors[m.status] ?? 'var(--text-muted)' }"
              >
                <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: statusColors[m.status] }"></span>
                {{ statusLabel(m.status) }}
              </span>
            </div>
            <div class="flex items-center justify-between text-xs" style="color: var(--text-dim)">
              <span class="font-mono truncate max-w-[60%]">{{ hostPort(m) }}</span>
              <span>{{ fmtPct(m.uptime24h) }} · {{ relativeTime(m.lastCheckedAt) }}</span>
            </div>
            <div class="text-xs mt-1" style="color: var(--text-muted)">
              {{ m.expectedState === 'closed' ? '🔒 Expected closed' : '🔓 Expected open' }}
            </div>
          </div>
        </div>

        <!-- Desktop table -->
        <div class="hidden md:block rounded-xl border overflow-hidden" style="border-color: var(--border)">
          <table class="w-full text-sm">
            <thead>
              <tr style="background-color: var(--surface-raised); border-bottom: 1px solid var(--border)">
                <th class="text-left px-4 py-3 font-semibold text-[10.5px] uppercase tracking-wide" style="color: var(--text-muted)">Name</th>
                <th class="text-left px-4 py-3 font-semibold text-[10.5px] uppercase tracking-wide" style="color: var(--text-muted)">Host:Port</th>
                <th class="text-left px-4 py-3 font-semibold text-[10.5px] uppercase tracking-wide" style="color: var(--text-muted)">Expected state</th>
                <th class="text-left px-4 py-3 font-semibold text-[10.5px] uppercase tracking-wide" style="color: var(--text-muted)">Status</th>
                <th class="text-left px-4 py-3 font-semibold text-[10.5px] uppercase tracking-wide" style="color: var(--text-muted)">Uptime 24h</th>
                <th class="text-left px-4 py-3 font-semibold text-[10.5px] uppercase tracking-wide" style="color: var(--text-muted)">Last checked</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="m in monitors"
                :key="m.id"
                class="cursor-pointer transition-colors bg-[var(--surface)] hover:bg-[var(--surface-raised)]"
                style="border-bottom: 1px solid var(--border)"
                @click="router.push({ name: 'port-monitor-detail', params: { id: m.id } })"
              >
                <td class="px-4 py-3 font-medium" style="color: var(--text-strong)">{{ m.name }}</td>
                <td class="px-4 py-3 font-mono text-xs" style="color: var(--text-dim)">{{ hostPort(m) }}</td>
                <td class="px-4 py-3 text-xs" style="color: var(--text-dim)">
                  {{ m.expectedState === 'closed' ? '🔒 Closed' : '🔓 Open' }}
                </td>
                <td class="px-4 py-3">
                  <span
                    class="inline-flex items-center gap-1.5 text-xs font-medium"
                    :style="{ color: statusColors[m.status] ?? 'var(--text-muted)' }"
                  >
                    <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: statusColors[m.status] }"></span>
                    {{ statusLabel(m.status) }}
                  </span>
                </td>
                <td class="px-4 py-3" style="color: var(--text-dim)">{{ fmtPct(m.uptime24h) }}</td>
                <td class="px-4 py-3" style="color: var(--text-dim)">{{ relativeTime(m.lastCheckedAt) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </div>
  </AppLayout>
</template>
