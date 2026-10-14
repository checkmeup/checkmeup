<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useQueryClient } from '@tanstack/vue-query'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import { maintenanceApi, type MaintenanceWindow } from '@/api/maintenance'
import { useMaintenanceWindows } from '@/composables/useMaintenanceWindows'

const router = useRouter()
const queryClient = useQueryClient()
const { data, isPending: loading, error: queryError, refetch } = useMaintenanceWindows()
const windows = computed<MaintenanceWindow[]>(() => data.value ?? [])
const actionError = ref('')
const error = computed(() => actionError.value || queryError.value?.message || '')
const endingId = ref('')

function load() {
  refetch()
}

const statusColors: Record<string, string> = {
  upcoming: 'var(--text-dim)',
  active: 'var(--status-paused)',
  ended: 'var(--text-muted)',
}

const statusLabels: Record<string, string> = {
  upcoming: 'Upcoming',
  active: 'Active',
  ended: 'Ended',
}

function formatDate(iso: string | null) {
  if (!iso) return 'until ended manually'
  return new Date(iso).toLocaleString(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  })
}

async function endNow(w: MaintenanceWindow) {
  endingId.value = w.id
  actionError.value = ''
  try {
    const updated = await maintenanceApi.endNow(w.id)
    queryClient.setQueryData<MaintenanceWindow[]>(['maintenance-windows'], (old) =>
      old?.map((x) => (x.id === updated.id ? updated : x)),
    )
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : 'Failed to end maintenance window'
  } finally {
    endingId.value = ''
  }
}
</script>

<template>
  <AppLayout>
    <div class="mx-auto max-w-[1200px] px-4 pt-8 pb-20 md:px-6">
      <div class="flex items-center justify-between mb-6">
        <h1 class="text-[30px] leading-[1.1] font-semibold tracking-[-0.035em] text-[var(--text-strong)] md:text-[34px]">Maintenance windows</h1>
        <Button @click="router.push({ name: 'maintenance-create' })">Schedule maintenance</Button>
      </div>

      <div v-if="loading" class="text-sm text-[var(--text-muted)]">Loading…</div>

      <div v-else-if="error" class="rounded-[14px] border border-[var(--border)] p-6 text-center">
        <p class="text-sm mb-4 text-[var(--status-down)]">{{ error }}</p>
        <Button variant="secondary" size="sm" @click="load">Try again</Button>
      </div>

      <div
        v-else-if="windows.length === 0"
        class="rounded-[14px] border border-[var(--border)] p-12 text-center"
      >
        <p class="text-sm mb-4 text-[var(--text-muted)]">
          No maintenance windows yet. Schedule one to suppress alerts during planned downtime.
        </p>
        <Button @click="router.push({ name: 'maintenance-create' })">Schedule your first window</Button>
      </div>

      <template v-else>
        <!-- Mobile cards -->
        <div class="md:hidden space-y-2">
          <div
            v-for="w in windows"
            :key="w.id"
            class="rounded-[14px] border border-[var(--border)] p-4 cursor-pointer"
            @click="router.push({ name: 'maintenance-edit', params: { id: w.id } })"
          >
            <div class="flex items-center justify-between mb-2">
              <span class="font-medium text-sm text-[var(--text-strong)]">{{ w.title }}</span>
              <span
                class="inline-flex items-center gap-1.5 text-xs font-medium"
                :style="{ color: statusColors[w.status] }"
              >
                <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: statusColors[w.status] }"></span>
                {{ statusLabels[w.status] }}
              </span>
            </div>
            <div class="flex items-center gap-4 text-xs text-[var(--text-dim)]">
              <span>{{ formatDate(w.startsAt) }} → {{ formatDate(w.endsAt) }}</span>
              <span>{{ w.monitorCount }} monitor{{ w.monitorCount === 1 ? '' : 's' }}</span>
            </div>
            <Button
              v-if="w.status !== 'ended'"
              variant="secondary"
              size="sm"
              class="mt-2"
              :disabled="endingId === w.id"
              @click.stop="endNow(w)"
            >
              {{ endingId === w.id ? 'Ending…' : 'End now' }}
            </Button>
          </div>
        </div>

        <!-- Desktop table -->
        <div class="hidden md:block rounded-xl border overflow-hidden border-[var(--border)]">
          <table class="w-full text-sm">
            <thead>
              <tr class="bg-[var(--surface)]">
                <th class="text-left px-4 py-2.5 font-medium text-[12.5px] text-[var(--text-dim)]">Title</th>
                <th class="text-left px-4 py-2.5 font-medium text-[12.5px] text-[var(--text-dim)]">Status</th>
                <th class="text-left px-4 py-2.5 font-medium text-[12.5px] text-[var(--text-dim)]">Starts</th>
                <th class="text-left px-4 py-2.5 font-medium text-[12.5px] text-[var(--text-dim)]">Ends</th>
                <th class="text-left px-4 py-2.5 font-medium text-[12.5px] text-[var(--text-dim)]">Monitors</th>
                <th class="px-4 py-3"></th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="w in windows"
                :key="w.id"
                class="cursor-pointer border-t border-[var(--border)] transition-colors hover:bg-[var(--surface)]"
                @click="router.push({ name: 'maintenance-edit', params: { id: w.id } })"
              >
                <td class="px-4 py-3 font-medium text-[var(--text-strong)]">{{ w.title }}</td>
                <td class="px-4 py-3">
                  <span
                    class="inline-flex items-center gap-1.5 text-xs font-medium"
                    :style="{ color: statusColors[w.status] }"
                  >
                    <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: statusColors[w.status] }"></span>
                    {{ statusLabels[w.status] }}
                  </span>
                </td>
                <td class="px-4 py-3 text-[var(--text-dim)]">{{ formatDate(w.startsAt) }}</td>
                <td class="px-4 py-3 text-[var(--text-dim)]">{{ formatDate(w.endsAt) }}</td>
                <td class="px-4 py-3 text-[var(--text-dim)]">{{ w.monitorCount }}</td>
                <td class="px-4 py-3 text-right">
                  <Button
                    v-if="w.status !== 'ended'"
                    variant="secondary"
                    size="sm"
                    :disabled="endingId === w.id"
                    @click.stop="endNow(w)"
                  >
                    {{ endingId === w.id ? 'Ending…' : 'End now' }}
                  </Button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </div>
  </AppLayout>
</template>
