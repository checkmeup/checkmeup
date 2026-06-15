<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import { monitorsApi, type UptimeMonitor } from '@/api/monitors'

const router = useRouter()
const monitors = ref<UptimeMonitor[]>([])
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    monitors.value = await monitorsApi.listUptime()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load monitors'
  } finally {
    loading.value = false
  }
})

const statusColors: Record<string, string> = {
  up: 'var(--status-up)',
  down: 'var(--status-down)',
  waiting: 'var(--text-muted)',
  paused: 'var(--status-paused)',
}

function statusLabel(s: string) {
  return s.charAt(0).toUpperCase() + s.slice(1)
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

function fmtPct(v: number | null): string {
  if (v === null || v === undefined) return '—'
  return `${v.toFixed(2)}%`
}

function truncate(s: string, n = 40) {
  return s.length > n ? s.slice(0, n) + '…' : s
}
</script>

<template>
  <AppLayout>
    <div class="p-8 max-w-4xl mx-auto">
      <div class="flex items-center justify-between mb-6">
        <h1 class="text-2xl font-semibold" style="color: var(--text-strong)">Uptime monitors</h1>
        <Button @click="router.push({ name: 'uptime-monitor-create' })">
          Add monitor
        </Button>
      </div>

      <div v-if="loading" class="text-sm" style="color: var(--text-muted)">Loading…</div>

      <div v-else-if="error" class="text-sm" style="color: var(--status-down)">{{ error }}</div>

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

      <div v-else class="rounded-xl border overflow-hidden" style="border-color: var(--border)">
        <table class="w-full text-sm">
          <thead>
            <tr style="background-color: var(--surface); border-bottom: 1px solid var(--border)">
              <th class="text-left px-4 py-3 font-medium" style="color: var(--text-muted)">Name</th>
              <th class="text-left px-4 py-3 font-medium" style="color: var(--text-muted)">URL</th>
              <th class="text-left px-4 py-3 font-medium" style="color: var(--text-muted)">Status</th>
              <th class="text-left px-4 py-3 font-medium" style="color: var(--text-muted)">Uptime 24h</th>
              <th class="text-left px-4 py-3 font-medium" style="color: var(--text-muted)">Last checked</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="m in monitors"
              :key="m.id"
              class="cursor-pointer transition-colors"
              style="background-color: var(--surface); border-bottom: 1px solid var(--border)"
              @click="router.push({ name: 'uptime-monitor-detail', params: { id: m.id } })"
            >
              <td class="px-4 py-3 font-medium" style="color: var(--text-strong)">{{ m.name }}</td>
              <td class="px-4 py-3 font-mono text-xs" style="color: var(--text-dim)">{{ truncate(m.url) }}</td>
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
    </div>
  </AppLayout>
</template>
