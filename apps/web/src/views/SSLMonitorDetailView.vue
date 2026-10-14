<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import UpgradePrompt from '@/components/UpgradePrompt.vue'
import MonitorDetailHeader from '@/components/MonitorDetailHeader.vue'
import StatGrid from '@/components/StatGrid.vue'
import { monitorsApi } from '@/api/monitors'
import { useSSLMonitor } from '@/composables/useSSLMonitors'
import { useMonitorActions } from '@/composables/useMonitorActions'
import {
  fmtDaysLeft,
  monitorsListRoute,
  relativeTime,
  statusLabel,
  toneOf,
} from '@/lib/monitorRows'

const router = useRouter()
const route = useRoute()
const id = route.params.id as string

const { data: monitor, isPending: loading, error: queryError, refetch } = useSSLMonitor(id)
const error = computed(() => queryError.value?.message ?? '')

const { actionError, limitReached, confirmDelete, togglePause, deleteMonitor } = useMonitorActions({
  isPaused: () => monitor.value?.status === 'paused',
  pause: () => monitorsApi.pauseSSL(id),
  resume: () => monitorsApi.resumeSSL(id),
  remove: () => monitorsApi.deleteSSL(id),
  refetch,
  afterDelete: () => router.push({ name: 'ssl-monitors' }),
})

function fmtDate(iso: string | null) {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

const daysColor = computed(() => {
  const d = monitor.value?.daysUntilExpiry
  if (d === null || d === undefined) return undefined
  if (d <= 14) return 'var(--status-down)'
  if (d <= 30) return 'var(--status-degraded)'
  return undefined
})

const summary = computed(() => {
  const m = monitor.value
  if (!m) return ''
  const date = fmtDate(m.expiresAt)
  switch (m.status) {
    case 'up':
      return `The certificate is valid for another ${fmtDaysLeft(m.daysUntilExpiry)}, until ${date}.`
    case 'expiring_soon':
      return `The certificate expires in ${fmtDaysLeft(m.daysUntilExpiry)} — renew it before ${date}.`
    case 'expired':
      return `The certificate expired on ${date}.`
    case 'error':
      return 'The last check could not read the certificate.'
    case 'paused':
      return 'Paused — no checks run until you resume it.'
    default:
      return 'Waiting for the first check.'
  }
})
</script>

<template>
  <AppLayout>
    <div class="mx-auto flex max-w-[1200px] flex-col gap-11 px-4 pt-8 pb-20 md:px-6">
      <p v-if="loading" class="text-sm text-[var(--text-dim)]">Loading…</p>
      <p v-else-if="error" class="text-sm text-[var(--status-down)]">{{ error }}</p>

      <template v-else-if="monitor">
        <MonitorDetailHeader
          v-model:confirm-delete="confirmDelete"
          :back-to="monitorsListRoute('ssl')"
          back-label="SSL monitors"
          :name="monitor.name"
          :tone="toneOf(monitor.status)"
          :status-label="statusLabel('ssl', monitor.status)"
          :paused="monitor.status === 'paused'"
          @edit="router.push({ name: 'ssl-monitor-edit', params: { id } })"
          @toggle-pause="togglePause"
          @delete="deleteMonitor"
        >
          <p class="text-lg text-[var(--text-dim)]">{{ summary }}</p>
          <p class="mt-1 font-mono text-[13px] text-[var(--text-dim)]">{{ monitor.hostname }}</p>
        </MonitorDetailHeader>

        <UpgradePrompt v-if="limitReached" :message="actionError" />
        <p v-else-if="actionError" class="text-sm text-[var(--status-down)]">{{ actionError }}</p>

        <p
          v-if="monitor.errorMsg"
          class="rounded-[14px] bg-[var(--status-down-wash)] p-4 font-mono text-[13px] text-[var(--status-down)]"
        >
          {{ monitor.errorMsg }}
        </p>

        <p
          v-if="!monitor.lastCheckedAt"
          class="border-y border-[var(--border)] py-6 text-center text-sm text-[var(--text-dim)]"
        >
          First check runs within 24 hours of creation.
        </p>

        <StatGrid
          :items="[
            {
              label: 'Days remaining',
              value: monitor.daysUntilExpiry !== null ? fmtDaysLeft(monitor.daysUntilExpiry) : '—',
              color: daysColor,
            },
            { label: 'Expires', value: fmtDate(monitor.expiresAt) },
            { label: 'Issuer', value: monitor.issuer ?? '—' },
            { label: 'Last checked', value: relativeTime(monitor.lastCheckedAt) },
          ]"
        />
      </template>
    </div>
  </AppLayout>
</template>
