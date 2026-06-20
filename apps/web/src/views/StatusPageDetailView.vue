<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import { statusPagesApi } from '@/api/statusPages'
import { useStatusPage } from '@/composables/useStatusPages'
import { useCronMonitors } from '@/composables/useCronMonitors'
import { useUptimeMonitors } from '@/composables/useUptimeMonitors'
import { useSSLMonitors } from '@/composables/useSSLMonitors'

const router = useRouter()
const route = useRoute()
const id = route.params.id as string

const { data: detail, isPending: pageLoading, error: pageError, refetch: refetchPage } = useStatusPage(id)
const { data: cronData, isPending: cronLoading } = useCronMonitors()
const { data: uptimeData, isPending: uptimeLoading } = useUptimeMonitors()
const { data: sslData, isPending: sslLoading } = useSSLMonitors()

const loading = computed(
  () => pageLoading.value || cronLoading.value || uptimeLoading.value || sslLoading.value,
)
const error = computed(() => pageError.value?.message ?? '')
const actionError = ref('')
const confirmDelete = ref(false)
const savingMonitors = ref(false)

// All monitors from all types
const cronMonitors = computed(() => cronData.value ?? [])
const uptimeMonitors = computed(() => uptimeData.value ?? [])
const sslMonitors = computed(() => sslData.value ?? [])

// Local editable list of selected monitors (copy from detail)
interface MonitorEntry {
  monitorType: 'cron' | 'uptime' | 'ssl'
  monitorId: string
  displayName: string
  displayOrder: number
  // UI helper
  key: string
}

const monitorEntries = ref<MonitorEntry[]>([])

watch(
  detail,
  (page) => {
    if (!page) return
    monitorEntries.value = page.monitors.map((m, i) => ({
      ...m,
      key: `${m.monitorType}:${m.monitorId}`,
      displayOrder: i,
    }))
  },
  { immediate: true },
)

// All available monitors across all types
const allMonitors = computed(() => {
  const result: { key: string; type: 'cron' | 'uptime' | 'ssl'; id: string; name: string }[] = []
  cronMonitors.value.forEach((m) =>
    result.push({ key: `cron:${m.id}`, type: 'cron', id: m.id, name: m.name }),
  )
  uptimeMonitors.value.forEach((m) =>
    result.push({ key: `uptime:${m.id}`, type: 'uptime', id: m.id, name: m.name }),
  )
  sslMonitors.value.forEach((m) =>
    result.push({ key: `ssl:${m.id}`, type: 'ssl', id: m.id, name: m.name }),
  )
  return result
})

const selectedKeys = computed(() => new Set(monitorEntries.value.map((e) => e.key)))

function toggleMonitor(m: { key: string; type: 'cron' | 'uptime' | 'ssl'; id: string; name: string }) {
  if (selectedKeys.value.has(m.key)) {
    monitorEntries.value = monitorEntries.value.filter((e) => e.key !== m.key)
  } else {
    monitorEntries.value = [
      ...monitorEntries.value,
      {
        monitorType: m.type,
        monitorId: m.id,
        displayName: m.name,
        displayOrder: monitorEntries.value.length,
        key: m.key,
      },
    ]
  }
}

function moveUp(i: number) {
  if (i === 0) return
  const arr = [...monitorEntries.value]
  ;[arr[i - 1], arr[i]] = [arr[i], arr[i - 1]]
  monitorEntries.value = arr.map((e, j) => ({ ...e, displayOrder: j }))
}

function moveDown(i: number) {
  if (i === monitorEntries.value.length - 1) return
  const arr = [...monitorEntries.value]
  ;[arr[i], arr[i + 1]] = [arr[i + 1], arr[i]]
  monitorEntries.value = arr.map((e, j) => ({ ...e, displayOrder: j }))
}

async function saveMonitors() {
  savingMonitors.value = true
  actionError.value = ''
  try {
    const monitors = monitorEntries.value.map((e, i) => ({
      monitorType: e.monitorType,
      monitorId: e.monitorId,
      displayName: e.displayName,
      displayOrder: i,
    }))
    await statusPagesApi.setMonitors(id, { monitors })
    await refetchPage()
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : 'Failed to save monitors'
  } finally {
    savingMonitors.value = false
  }
}

async function deletePage() {
  actionError.value = ''
  try {
    await statusPagesApi.delete(id)
    router.push({ name: 'status-pages' })
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : 'Delete failed'
    confirmDelete.value = false
  }
}

const typeLabel: Record<string, string> = { cron: 'Cron', uptime: 'Uptime', ssl: 'SSL' }
</script>

<template>
  <AppLayout>
    <div class="p-8 max-w-3xl mx-auto">
      <div class="flex items-center gap-3 mb-6">
        <button
          class="text-sm"
          style="color: var(--text-muted)"
          @click="router.push({ name: 'status-pages' })"
        >
          ← Back
        </button>
        <h1 class="text-2xl font-semibold" style="color: var(--text-strong)">
          {{ detail?.title ?? 'Status page' }}
        </h1>
      </div>

      <div v-if="loading" class="text-sm" style="color: var(--text-muted)">Loading…</div>
      <div v-else-if="error" class="text-sm" style="color: var(--status-down)">{{ error }}</div>

      <template v-else-if="detail">
        <!-- Page info card -->
        <div
          class="rounded-xl border p-5 mb-6"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div class="flex items-start justify-between gap-4 mb-4">
            <div>
              <p class="text-sm font-mono" style="color: var(--text-dim)">/status/{{ detail.slug }}</p>
              <p v-if="detail.description" class="text-xs mt-1" style="color: var(--text-muted)">
                {{ detail.description }}
              </p>
            </div>
            <div class="flex items-center gap-2 flex-shrink-0">
              <a
                :href="detail.publicUrl"
                target="_blank"
                rel="noopener"
                class="text-xs px-3 py-1.5 rounded-md border transition-colors"
                style="color: var(--text-dim); border-color: var(--border)"
              >
                View public page →
              </a>
              <Button
                variant="secondary"
                size="sm"
                @click="router.push({ name: 'status-page-edit', params: { id } })"
              >
                Edit
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
                <Button size="sm" style="background-color: var(--status-down)" @click="deletePage">
                  Confirm delete
                </Button>
                <Button variant="secondary" size="sm" @click="confirmDelete = false">Cancel</Button>
              </template>
            </div>
          </div>
        </div>

        <p v-if="actionError" class="text-sm mb-4" style="color: var(--status-down)">{{ actionError }}</p>

        <!-- Monitor management -->
        <div class="grid grid-cols-2 gap-6">
          <!-- Available monitors -->
          <div
            class="rounded-xl border"
            style="background-color: var(--surface); border-color: var(--border)"
          >
            <div class="px-4 py-3 border-b text-sm font-medium" style="border-color: var(--border); color: var(--text-strong)">
              Available monitors
            </div>
            <div v-if="allMonitors.length === 0" class="px-4 py-3 text-xs" style="color: var(--text-muted)">
              No monitors yet.
            </div>
            <ul v-else>
              <li
                v-for="m in allMonitors"
                :key="m.key"
                class="flex items-center gap-3 px-4 py-2.5 cursor-pointer transition-colors border-b last:border-0"
                style="border-color: var(--border)"
                @click="toggleMonitor(m)"
              >
                <div
                  class="w-4 h-4 rounded border flex-shrink-0 flex items-center justify-center text-xs font-bold"
                  :style="{
                    backgroundColor: selectedKeys.has(m.key) ? 'var(--accent)' : 'transparent',
                    borderColor: selectedKeys.has(m.key) ? 'var(--accent)' : 'var(--border)',
                    color: '#fff',
                  }"
                >
                  {{ selectedKeys.has(m.key) ? '✓' : '' }}
                </div>
                <span class="text-sm flex-1 truncate" style="color: var(--text)">{{ m.name }}</span>
                <span
                  class="text-xs flex-shrink-0"
                  style="color: var(--text-muted)"
                >{{ typeLabel[m.type] }}</span>
              </li>
            </ul>
          </div>

          <!-- Selected monitors + order -->
          <div
            class="rounded-xl border"
            style="background-color: var(--surface); border-color: var(--border)"
          >
            <div class="px-4 py-3 border-b text-sm font-medium" style="border-color: var(--border); color: var(--text-strong)">
              On this page ({{ monitorEntries.length }})
            </div>
            <div v-if="monitorEntries.length === 0" class="px-4 py-3 text-xs" style="color: var(--text-muted)">
              Select monitors from the left.
            </div>
            <ul v-else>
              <li
                v-for="(m, i) in monitorEntries"
                :key="m.key"
                class="flex items-center gap-2 px-4 py-2.5 border-b last:border-0"
                style="border-color: var(--border)"
              >
                <div class="flex flex-col gap-0.5">
                  <button
                    class="text-xs leading-none px-1"
                    style="color: var(--text-muted)"
                    :disabled="i === 0"
                    @click="moveUp(i)"
                  >▲</button>
                  <button
                    class="text-xs leading-none px-1"
                    style="color: var(--text-muted)"
                    :disabled="i === monitorEntries.length - 1"
                    @click="moveDown(i)"
                  >▼</button>
                </div>
                <input
                  v-model="m.displayName"
                  class="flex-1 text-sm bg-transparent border-0 focus:outline-none min-w-0"
                  style="color: var(--text)"
                  :placeholder="m.displayName"
                />
                <span class="text-xs flex-shrink-0" style="color: var(--text-muted)">
                  {{ typeLabel[m.monitorType] }}
                </span>
                <button
                  class="text-xs ml-1 flex-shrink-0"
                  style="color: var(--text-muted)"
                  @click="toggleMonitor({ key: m.key, type: m.monitorType, id: m.monitorId, name: m.displayName })"
                >✕</button>
              </li>
            </ul>
            <div class="px-4 py-3 border-t" style="border-color: var(--border)">
              <Button size="sm" :disabled="savingMonitors" @click="saveMonitors">
                {{ savingMonitors ? 'Saving…' : 'Save changes' }}
              </Button>
            </div>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>
