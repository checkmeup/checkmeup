<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import { type Incident } from '@/api/incidents'
import { useIncidents } from '@/composables/useIncidents'

const router = useRouter()
const { data, isPending: loading, error: queryError, refetch } = useIncidents()
const incidents = computed<Incident[]>(() => data.value ?? [])
const error = computed(() => queryError.value?.message || '')

function load() {
  refetch()
}

const severityColors: Record<string, string> = {
  minor: 'var(--status-paused)',
  major: 'var(--status-degraded)',
  critical: 'var(--status-down)',
}

const severityLabels: Record<string, string> = {
  minor: 'Minor',
  major: 'Major',
  critical: 'Critical',
}

const statusLabels: Record<string, string> = {
  investigating: 'Investigating',
  identified: 'Identified',
  monitoring: 'Monitoring',
  resolved: 'Resolved',
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })
}
</script>

<template>
  <AppLayout>
    <div class="mx-auto max-w-[1200px] px-4 pt-8 pb-20 md:px-6">
      <div class="flex items-center justify-between mb-6">
        <h1 class="text-[30px] leading-[1.1] font-semibold tracking-[-0.035em] text-[var(--text-strong)] md:text-[34px]">Incidents</h1>
        <Button @click="router.push({ name: 'incidents-create' })">Declare incident</Button>
      </div>

      <div v-if="loading" class="text-sm text-[var(--text-muted)]">Loading…</div>

      <div v-else-if="error" class="rounded-[14px] border border-[var(--border)] p-6 text-center">
        <p class="text-sm mb-4 text-[var(--status-down)]">{{ error }}</p>
        <Button variant="secondary" size="sm" @click="load">Try again</Button>
      </div>

      <div
        v-else-if="incidents.length === 0"
        class="rounded-[14px] border border-[var(--border)] p-12 text-center"
      >
        <p class="text-sm mb-4 text-[var(--text-muted)]">
          No incidents yet. Declare one when something's affecting your service that isn't a plain monitor-down.
        </p>
        <Button @click="router.push({ name: 'incidents-create' })">Declare your first incident</Button>
      </div>

      <template v-else>
        <!-- Mobile cards -->
        <div class="md:hidden space-y-2">
          <div
            v-for="inc in incidents"
            :key="inc.id"
            class="rounded-[14px] border border-[var(--border)] p-4 cursor-pointer"
            @click="router.push({ name: 'incident-detail', params: { id: inc.id } })"
          >
            <div class="flex items-center justify-between mb-2">
              <span class="font-medium text-sm text-[var(--text-strong)]">{{ inc.title }}</span>
              <span
                class="inline-flex items-center gap-1.5 text-xs font-medium"
                :style="{ color: severityColors[inc.severity] }"
              >
                <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: severityColors[inc.severity] }"></span>
                {{ severityLabels[inc.severity] }}
              </span>
            </div>
            <div class="flex items-center gap-4 text-xs text-[var(--text-dim)]">
              <span>{{ statusLabels[inc.status] }}</span>
              <span>{{ formatDate(inc.createdAt) }}</span>
              <span>{{ inc.monitorCount }} monitor{{ inc.monitorCount === 1 ? '' : 's' }}</span>
            </div>
          </div>
        </div>

        <!-- Desktop table -->
        <div class="hidden md:block rounded-xl border overflow-hidden border-[var(--border)]">
          <table class="w-full text-sm">
            <thead>
              <tr class="bg-[var(--surface)]">
                <th class="text-left px-4 py-2.5 font-medium text-[12.5px] text-[var(--text-dim)]">Title</th>
                <th class="text-left px-4 py-2.5 font-medium text-[12.5px] text-[var(--text-dim)]">Severity</th>
                <th class="text-left px-4 py-2.5 font-medium text-[12.5px] text-[var(--text-dim)]">Status</th>
                <th class="text-left px-4 py-2.5 font-medium text-[12.5px] text-[var(--text-dim)]">Monitors</th>
                <th class="text-left px-4 py-2.5 font-medium text-[12.5px] text-[var(--text-dim)]">Created</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="inc in incidents"
                :key="inc.id"
                class="cursor-pointer border-t border-[var(--border)] transition-colors hover:bg-[var(--surface)]"
                @click="router.push({ name: 'incident-detail', params: { id: inc.id } })"
              >
                <td class="px-4 py-3 font-medium text-[var(--text-strong)]">{{ inc.title }}</td>
                <td class="px-4 py-3">
                  <span
                    class="inline-flex items-center gap-1.5 text-xs font-medium"
                    :style="{ color: severityColors[inc.severity] }"
                  >
                    <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: severityColors[inc.severity] }"></span>
                    {{ severityLabels[inc.severity] }}
                  </span>
                </td>
                <td class="px-4 py-3 text-[var(--text-dim)]">{{ statusLabels[inc.status] }}</td>
                <td class="px-4 py-3 text-[var(--text-dim)]">{{ inc.monitorCount }}</td>
                <td class="px-4 py-3 text-[var(--text-dim)]">{{ formatDate(inc.createdAt) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </div>
  </AppLayout>
</template>
