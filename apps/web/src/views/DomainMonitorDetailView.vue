<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import UpgradePrompt from '@/components/UpgradePrompt.vue'
import { ApiError } from '@/api/client'
import { monitorsApi } from '@/api/monitors'
import { useDomainMonitor } from '@/composables/useDomainMonitors'

const router = useRouter()
const route = useRoute()
const id = route.params.id as string

const { data: monitor, isPending: loading, error: queryError, refetch } = useDomainMonitor(id)
const error = computed(() => queryError.value?.message ?? '')
const actionError = ref('')
const limitReached = ref(false)
const confirmDelete = ref(false)

async function togglePause() {
  if (!monitor.value) return
  actionError.value = ''
  limitReached.value = false
  try {
    if (monitor.value.status === 'paused') {
      await monitorsApi.resumeDomain(id)
    } else {
      await monitorsApi.pauseDomain(id)
    }
    await refetch()
  } catch (e: unknown) {
    if (e instanceof ApiError && e.code === 'plan_limit_reached') {
      limitReached.value = true
      actionError.value = e.message
    } else {
      actionError.value = e instanceof Error ? e.message : 'Action failed'
    }
  }
}

async function deleteMonitor() {
  actionError.value = ''
  try {
    await monitorsApi.deleteDomain(id)
    router.push({ name: 'domain-monitors' })
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : 'Delete failed'
    confirmDelete.value = false
  }
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

function fmtDate(iso: string | null) {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}
</script>

<template>
  <AppLayout>
    <div class="p-8 max-w-2xl mx-auto">
      <div class="flex items-center gap-3 mb-6 flex-wrap">
        <button
          class="text-sm transition-colors"
          style="color: var(--text-muted)"
          @click="router.push({ name: 'domain-monitors' })"
        >
          ← Back
        </button>
        <h1 class="text-2xl font-semibold" style="color: var(--text-strong)">
          {{ monitor?.name ?? 'Domain Monitor' }}
        </h1>
        <span
          v-if="monitor"
          class="inline-flex items-center gap-1.5 text-xs font-semibold px-2.5 py-1 rounded-full flex-shrink-0"
          :style="{ backgroundColor: 'var(--surface-raised)', color: statusColors[monitor.status] ?? 'var(--text-muted)' }"
        >
          <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: statusColors[monitor.status] }"></span>
          {{ statusLabel(monitor.status) }}
        </span>
      </div>

      <div v-if="loading" class="text-sm" style="color: var(--text-muted)">Loading…</div>
      <div v-else-if="error" class="text-sm" style="color: var(--status-down)">{{ error }}</div>

      <template v-else-if="monitor">
        <!-- Header card -->
        <div
          class="rounded-xl border p-5 mb-6"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div class="flex items-start justify-between gap-4 mb-4">
            <div>
              <p class="font-mono text-sm" style="color: var(--text-dim)">{{ monitor.domain }}</p>
            </div>
            <div class="flex items-center gap-2 flex-shrink-0">
              <Button
                variant="secondary"
                size="sm"
                @click="router.push({ name: 'domain-monitor-edit', params: { id } })"
              >
                Edit
              </Button>
              <Button variant="secondary" size="sm" @click="togglePause">
                {{ monitor.status === 'paused' ? 'Resume' : 'Pause' }}
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

          <div class="grid grid-cols-2 gap-x-8 gap-y-3 text-sm border-t pt-4" style="border-color: var(--border)">
            <div>
              <span class="block text-xs mb-0.5" style="color: var(--text-muted)">Expires</span>
              <span style="color: var(--text)">{{ fmtDate(monitor.expiresAt) }}</span>
            </div>
            <div>
              <span class="block text-xs mb-0.5" style="color: var(--text-muted)">Days remaining</span>
              <span
                :style="{
                  color: monitor.daysUntilExpiry !== null && monitor.daysUntilExpiry <= 14
                    ? 'var(--status-down)'
                    : monitor.daysUntilExpiry !== null && monitor.daysUntilExpiry <= 30
                    ? 'var(--status-degraded)'
                    : 'var(--text)',
                }"
              >
                {{ monitor.daysUntilExpiry !== null ? `${monitor.daysUntilExpiry} days` : '—' }}
              </span>
            </div>
            <div>
              <span class="block text-xs mb-0.5" style="color: var(--text-muted)">Registrar</span>
              <span style="color: var(--text)">{{ monitor.registrar ?? '—' }}</span>
            </div>
            <div>
              <span class="block text-xs mb-0.5" style="color: var(--text-muted)">Last checked</span>
              <span style="color: var(--text)">{{ relativeTime(monitor.lastCheckedAt) }}</span>
            </div>
          </div>

          <div v-if="monitor.errorMsg" class="mt-4 rounded-lg p-3 text-xs font-mono" style="background-color: var(--surface-raised); color: var(--status-down)">
            {{ monitor.errorMsg }}
          </div>
        </div>

        <UpgradePrompt v-if="limitReached" class="mb-4" :message="actionError" />
        <p v-else-if="actionError" class="text-sm" style="color: var(--status-down)">{{ actionError }}</p>

        <div
          v-if="!monitor.lastCheckedAt"
          class="rounded-xl border p-5 text-center"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <p class="text-sm" style="color: var(--text-muted)">
            First check runs within 24 hours of creation.
          </p>
        </div>
      </template>
    </div>
  </AppLayout>
</template>
