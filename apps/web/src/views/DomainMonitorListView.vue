<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import type { DomainMonitor } from '@/api/monitors'
import { useDomainMonitors } from '@/composables/useDomainMonitors'

const router = useRouter()
const { data, isPending: loading, error: queryError, refetch } = useDomainMonitors()
const monitors = computed<DomainMonitor[]>(() => data.value ?? [])
const error = computed(() => queryError.value?.message ?? '')

function load() {
  refetch()
}

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

function fmtExpiry(m: DomainMonitor): string {
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
    <div class="p-4 md:p-8 max-w-4xl mx-auto">
      <div class="flex items-center justify-between mb-6">
        <h1 class="text-2xl font-semibold" style="color: var(--text-strong)">Domain monitors</h1>
        <Button @click="router.push({ name: 'domain-monitor-create' })">
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
          No domain monitors yet. Add one to track registration expiry.
        </p>
        <Button @click="router.push({ name: 'domain-monitor-create' })">
          Add your first monitor
        </Button>
      </div>

      <template v-else>
        <!-- Mobile cards -->
        <div class="md:hidden space-y-2">
          <div
            v-for="m in monitors"
            :key="m.id"
            class="rounded-xl border p-4 cursor-pointer"
            style="background-color: var(--surface); border-color: var(--border)"
            @click="router.push({ name: 'domain-monitor-detail', params: { id: m.id } })"
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
              <span class="font-mono">{{ m.domain }}</span>
              <span>{{ fmtExpiry(m) }}</span>
            </div>
          </div>
        </div>

        <!-- Desktop table -->
        <div class="hidden md:block rounded-xl border overflow-hidden" style="border-color: var(--border)">
          <table class="w-full text-sm">
            <thead>
              <tr style="background-color: var(--surface); border-bottom: 1px solid var(--border)">
                <th class="text-left px-4 py-3 font-medium" style="color: var(--text-muted)">Name</th>
                <th class="text-left px-4 py-3 font-medium" style="color: var(--text-muted)">Domain</th>
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
                @click="router.push({ name: 'domain-monitor-detail', params: { id: m.id } })"
              >
                <td class="px-4 py-3 font-medium" style="color: var(--text-strong)">{{ m.name }}</td>
                <td class="px-4 py-3 font-mono text-xs" style="color: var(--text-dim)">{{ m.domain }}</td>
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
      </template>
    </div>
  </AppLayout>
</template>
