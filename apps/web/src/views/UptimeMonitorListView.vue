<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import type { UptimeMonitor } from '@/api/monitors'
import { useUptimeMonitors } from '@/composables/useUptimeMonitors'

const router = useRouter()
const { data, isPending: loading, error: queryError, refetch } = useUptimeMonitors()
const monitors = computed<UptimeMonitor[]>(() => data.value ?? [])
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

function truncate(s: string, n = 40) {
  return s.length > n ? s.slice(0, n) + '…' : s
}

function keywordLabel(m: UptimeMonitor): string {
  if (!m.keyword) return ''
  const verb = m.keywordMode === 'not_contains' ? 'does not contain' : 'contains'
  return `🔍 ${verb} "${truncate(m.keyword, 24)}"`
}
</script>

<template>
  <AppLayout>
    <div class="p-4 md:p-8 max-w-4xl mx-auto">
      <div class="flex items-center justify-between mb-6">
        <h1 class="text-2xl font-semibold" style="color: var(--text-strong)">Uptime monitors</h1>
        <Button @click="router.push({ name: 'uptime-monitor-create' })"> Add monitor </Button>
      </div>

      <div v-if="loading" class="text-sm" style="color: var(--text-muted)">Loading…</div>

      <div
        v-else-if="error"
        class="rounded-xl border p-6 text-center"
        style="background-color: var(--surface); border-color: var(--border)"
      >
        <p class="text-sm mb-4" style="color: var(--status-down)">{{ error }}</p>
        <Button variant="secondary" size="sm" @click="load">Try again</Button>
      </div>

      <div
        v-else-if="monitors.length === 0"
        class="rounded-xl border p-12 text-center"
        style="background-color: var(--surface); border-color: var(--border)"
      >
        <p class="text-sm mb-4" style="color: var(--text-muted)">
          No uptime monitors yet. Add one to start watching your URLs.
        </p>
        <Button @click="router.push({ name: 'uptime-monitor-create' })">
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
            @click="router.push({ name: 'uptime-monitor-detail', params: { id: m.id } })"
          >
            <div class="flex items-center justify-between mb-2">
              <span class="font-medium text-sm" style="color: var(--text-strong)">{{
                m.name
              }}</span>
              <span
                class="inline-flex items-center gap-1.5 text-xs font-medium"
                :style="{ color: statusColors[m.status] ?? 'var(--text-muted)' }"
              >
                <span
                  class="w-1.5 h-1.5 rounded-full"
                  :style="{ backgroundColor: statusColors[m.status] }"
                ></span>
                {{ statusLabel(m.status) }}
              </span>
            </div>
            <div class="flex items-center justify-between text-xs" style="color: var(--text-dim)">
              <span class="font-mono truncate max-w-[60%]">{{ m.url }}</span>
              <span>{{ fmtPct(m.uptime24h) }} · {{ relativeTime(m.lastCheckedAt) }}</span>
            </div>
            <div v-if="m.keyword" class="text-xs mt-1" style="color: var(--text-muted)">
              {{ keywordLabel(m) }}
            </div>
          </div>
        </div>

        <!-- Desktop table -->
        <div
          class="hidden md:block rounded-xl border overflow-hidden"
          style="border-color: var(--border)"
        >
          <table class="w-full text-sm">
            <thead>
              <tr style="background-color: var(--surface); border-bottom: 1px solid var(--border)">
                <th class="text-left px-4 py-3 font-medium" style="color: var(--text-muted)">
                  Name
                </th>
                <th class="text-left px-4 py-3 font-medium" style="color: var(--text-muted)">
                  URL
                </th>
                <th class="text-left px-4 py-3 font-medium" style="color: var(--text-muted)">
                  Status
                </th>
                <th class="text-left px-4 py-3 font-medium" style="color: var(--text-muted)">
                  Uptime 24h
                </th>
                <th class="text-left px-4 py-3 font-medium" style="color: var(--text-muted)">
                  Last checked
                </th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="m in monitors"
                :key="m.id"
                class="cursor-pointer transition-colors bg-[var(--surface)] hover:bg-[var(--surface-raised)]"
                style="border-bottom: 1px solid var(--border)"
                @click="router.push({ name: 'uptime-monitor-detail', params: { id: m.id } })"
              >
                <td class="px-4 py-3 font-medium" style="color: var(--text-strong)">
                  {{ m.name }}
                </td>
                <td class="px-4 py-3 font-mono text-xs" style="color: var(--text-dim)">
                  {{ truncate(m.url) }}
                  <div v-if="m.keyword" class="font-sans mt-0.5" style="color: var(--text-muted)">
                    {{ keywordLabel(m) }}
                  </div>
                </td>
                <td class="px-4 py-3">
                  <span
                    class="inline-flex items-center gap-1.5 text-xs font-medium"
                    :style="{ color: statusColors[m.status] ?? 'var(--text-muted)' }"
                  >
                    <span
                      class="w-1.5 h-1.5 rounded-full"
                      :style="{ backgroundColor: statusColors[m.status] }"
                    ></span>
                    {{ statusLabel(m.status) }}
                  </span>
                </td>
                <td class="px-4 py-3" style="color: var(--text-dim)">{{ fmtPct(m.uptime24h) }}</td>
                <td class="px-4 py-3" style="color: var(--text-dim)">
                  {{ relativeTime(m.lastCheckedAt) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </div>
  </AppLayout>
</template>
