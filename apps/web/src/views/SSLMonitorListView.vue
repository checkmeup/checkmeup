<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import { monitorsApi, type SSLMonitor } from '@/api/monitors'

const router = useRouter()
const monitors = ref<SSLMonitor[]>([])
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  try {
    monitors.value = await monitorsApi.listSSL()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load monitors'
  } finally {
    loading.value = false
  }
})

const statusColors: Record<string, string> = {
  up: 'var(--status-up)',
  expiring_soon: 'var(--status-degraded)',
  expired: 'var(--status-down)',
  error: 'var(--status-down)',
  waiting: 'var(--text-muted)',
  paused: 'var(--status-paused)',
}

function statusLabel(s: string) {
  const labels: Record<string, string> = {
    up: 'Valid',
    expiring_soon: 'Expiring soon',
    expired: 'Expired',
    error: 'Error',
    waiting: 'Waiting',
    paused: 'Paused',
  }
  return labels[s] ?? s
}

function fmtExpiry(m: SSLMonitor): string {
  if (!m.expiresAt) return '—'
  const days = m.daysUntilExpiry
  if (days === null) return new Date(m.expiresAt).toLocaleDateString()
  if (days < 0) return 'Expired'
  return `${days}d (${new Date(m.expiresAt).toLocaleDateString()})`
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
</script>

<template>
  <AppLayout>
    <div class="p-8 max-w-4xl mx-auto">
      <div class="flex items-center justify-between mb-6">
        <h1 class="text-2xl font-semibold" style="color: var(--text-strong)">SSL monitors</h1>
        <Button @click="router.push({ name: 'ssl-monitor-create' })">
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
          No SSL monitors yet. Add one to track certificate expiry.
        </p>
        <Button @click="router.push({ name: 'ssl-monitor-create' })">
          Add your first monitor
        </Button>
      </div>

      <div v-else class="rounded-xl border overflow-hidden" style="border-color: var(--border)">
        <table class="w-full text-sm">
          <thead>
            <tr style="background-color: var(--surface); border-bottom: 1px solid var(--border)">
              <th class="text-left px-4 py-3 font-medium" style="color: var(--text-muted)">Name</th>
              <th class="text-left px-4 py-3 font-medium" style="color: var(--text-muted)">Hostname</th>
              <th class="text-left px-4 py-3 font-medium" style="color: var(--text-muted)">Status</th>
              <th class="text-left px-4 py-3 font-medium" style="color: var(--text-muted)">Expires</th>
              <th class="text-left px-4 py-3 font-medium" style="color: var(--text-muted)">Last checked</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="m in monitors"
              :key="m.id"
              class="cursor-pointer transition-colors"
              style="background-color: var(--surface); border-bottom: 1px solid var(--border)"
              @click="router.push({ name: 'ssl-monitor-detail', params: { id: m.id } })"
            >
              <td class="px-4 py-3 font-medium" style="color: var(--text-strong)">{{ m.name }}</td>
              <td class="px-4 py-3 font-mono text-xs" style="color: var(--text-dim)">{{ m.hostname }}</td>
              <td class="px-4 py-3">
                <span
                  class="inline-flex items-center gap-1.5 text-xs font-medium"
                  :style="{ color: statusColors[m.status] ?? 'var(--text-muted)' }"
                >
                  <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: statusColors[m.status] }"></span>
                  {{ statusLabel(m.status) }}
                </span>
              </td>
              <td class="px-4 py-3" style="color: var(--text-dim)">{{ fmtExpiry(m) }}</td>
              <td class="px-4 py-3" style="color: var(--text-dim)">{{ relativeTime(m.lastCheckedAt) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </AppLayout>
</template>
