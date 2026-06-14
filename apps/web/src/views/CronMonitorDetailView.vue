<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import { monitorsApi, type CronMonitorDetail } from '@/api/monitors'

const router = useRouter()
const route = useRoute()
const id = route.params.id as string

const detail = ref<CronMonitorDetail | null>(null)
const loading = ref(true)
const error = ref('')
const actionPending = ref(false)
const showDeleteConfirm = ref(false)

onMounted(async () => {
  await load()
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    detail.value = await monitorsApi.getCron(id)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load monitor'
  } finally {
    loading.value = false
  }
}

const monitor = computed(() => detail.value?.monitor)

const statusColors: Record<string, string> = {
  up: 'var(--status-up)',
  down: 'var(--status-down)',
  waiting: 'var(--text-muted)',
  paused: 'var(--status-paused)',
}

function fmt(iso: string | null | undefined) {
  if (!iso) return '—'
  return new Date(iso).toLocaleString()
}

function duration(startIso: string, endIso: string | null) {
  const start = new Date(startIso).getTime()
  const end = endIso ? new Date(endIso).getTime() : Date.now()
  const mins = Math.floor((end - start) / 60000)
  if (mins < 60) return `${mins}m`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ${mins % 60}m`
  return `${Math.floor(hrs / 24)}d ${hrs % 24}h`
}

async function pause() {
  actionPending.value = true
  try {
    await monitorsApi.pauseCron(id)
    await load()
  } finally {
    actionPending.value = false
  }
}

async function resume() {
  actionPending.value = true
  try {
    await monitorsApi.resumeCron(id)
    await load()
  } finally {
    actionPending.value = false
  }
}

async function confirmDelete() {
  actionPending.value = true
  try {
    await monitorsApi.deleteCron(id)
    router.push({ name: 'cron-monitors' })
  } finally {
    actionPending.value = false
    showDeleteConfirm.value = false
  }
}

function copyPingUrl() {
  if (monitor.value?.pingUrl) {
    navigator.clipboard.writeText(monitor.value.pingUrl)
  }
}
</script>

<template>
  <AppLayout>
    <div class="p-8 max-w-4xl mx-auto">
      <div v-if="loading" class="text-sm" style="color: var(--text-muted)">Loading…</div>
      <div v-else-if="error" class="text-sm" style="color: var(--status-down)">{{ error }}</div>

      <template v-else-if="monitor">
        <!-- Header -->
        <div class="flex items-start justify-between mb-6">
          <div>
            <button
              class="text-sm mb-2 block transition-colors"
              style="color: var(--text-muted)"
              @click="router.push({ name: 'cron-monitors' })"
            >
              ← Cron monitors
            </button>
            <div class="flex items-center gap-3">
              <h1 class="text-2xl font-semibold" style="color: var(--text-strong)">
                {{ monitor.name }}
              </h1>
              <span
                class="inline-flex items-center gap-1.5 text-xs font-medium px-2 py-1 rounded-full"
                :style="{ color: statusColors[monitor.status], backgroundColor: 'var(--surface-raised)' }"
              >
                <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: statusColors[monitor.status] }"></span>
                {{ monitor.status.charAt(0).toUpperCase() + monitor.status.slice(1) }}
              </span>
            </div>
          </div>
          <div class="flex gap-2">
            <Button
              variant="secondary"
              size="sm"
              @click="router.push({ name: 'cron-monitor-edit', params: { id } })"
            >
              Edit
            </Button>
            <Button
              v-if="monitor.status !== 'paused'"
              variant="secondary"
              size="sm"
              :disabled="actionPending"
              @click="pause"
            >
              Pause
            </Button>
            <Button
              v-else
              variant="secondary"
              size="sm"
              :disabled="actionPending"
              @click="resume"
            >
              Resume
            </Button>
            <Button
              variant="destructive"
              size="sm"
              :disabled="actionPending"
              @click="showDeleteConfirm = true"
            >
              Delete
            </Button>
          </div>
        </div>

        <!-- Config card -->
        <div class="rounded-xl border p-5 mb-5" style="background-color: var(--surface); border-color: var(--border)">
          <h2 class="text-sm font-medium mb-4" style="color: var(--text-muted)">Configuration</h2>
          <div class="grid grid-cols-2 gap-4 text-sm">
            <div>
              <div style="color: var(--text-muted)">Schedule</div>
              <div class="font-mono mt-0.5" style="color: var(--text)">{{ monitor.schedule }}</div>
            </div>
            <div>
              <div style="color: var(--text-muted)">Grace period</div>
              <div class="mt-0.5" style="color: var(--text)">{{ monitor.gracePeriodMins }} min</div>
            </div>
            <div>
              <div style="color: var(--text-muted)">Last ping</div>
              <div class="mt-0.5" style="color: var(--text)">{{ fmt(monitor.lastPingAt) }}</div>
            </div>
            <div>
              <div style="color: var(--text-muted)">Next expected</div>
              <div class="mt-0.5" style="color: var(--text)">{{ fmt(monitor.nextPingAt) }}</div>
            </div>
          </div>

          <div class="mt-4 pt-4 border-t" style="border-color: var(--border)">
            <div class="text-xs mb-1" style="color: var(--text-muted)">Ping URL</div>
            <div class="flex items-center gap-2">
              <code
                class="flex-1 text-xs px-3 py-2 rounded-md truncate"
                style="background-color: var(--surface-raised); color: var(--text-dim)"
              >
                {{ monitor.pingUrl }}
              </code>
              <Button variant="secondary" size="sm" @click="copyPingUrl">Copy</Button>
            </div>
          </div>
        </div>

        <!-- Incidents -->
        <div v-if="detail!.incidents.length > 0" class="rounded-xl border p-5 mb-5" style="background-color: var(--surface); border-color: var(--border)">
          <h2 class="text-sm font-medium mb-4" style="color: var(--text-muted)">Incidents</h2>
          <div class="space-y-2">
            <div
              v-for="inc in detail!.incidents"
              :key="inc.id"
              class="flex items-center justify-between text-sm py-2 border-b last:border-0"
              style="border-color: var(--border)"
            >
              <div style="color: var(--text)">{{ fmt(inc.startedAt) }}</div>
              <div class="text-xs" style="color: var(--text-muted)">
                {{ inc.resolvedAt ? `Resolved after ${duration(inc.startedAt, inc.resolvedAt)}` : 'Ongoing' }}
              </div>
            </div>
          </div>
        </div>

        <!-- Ping log -->
        <div class="rounded-xl border p-5" style="background-color: var(--surface); border-color: var(--border)">
          <h2 class="text-sm font-medium mb-4" style="color: var(--text-muted)">Execution log</h2>

          <div v-if="detail!.pings.length === 0" class="text-sm" style="color: var(--text-muted)">
            No pings received yet. Add the ping URL to your cron job.
          </div>

          <div v-else class="space-y-0">
            <div
              v-for="ping in detail!.pings"
              :key="ping.id"
              class="flex items-center justify-between text-sm py-2 border-b last:border-0"
              style="border-color: var(--border)"
            >
              <div class="flex items-center gap-2">
                <span class="w-1.5 h-1.5 rounded-full" style="background-color: var(--status-up)"></span>
                <span style="color: var(--text)">{{ fmt(ping.receivedAt) }}</span>
              </div>
              <div class="text-xs font-mono" style="color: var(--text-muted)">{{ ping.sourceIp || '—' }}</div>
            </div>
          </div>
        </div>
      </template>

      <!-- Delete confirm dialog -->
      <div
        v-if="showDeleteConfirm"
        class="fixed inset-0 flex items-center justify-center z-50"
        style="background-color: rgba(0,0,0,0.6)"
        @click.self="showDeleteConfirm = false"
      >
        <div class="rounded-xl border p-6 max-w-sm w-full mx-4" style="background-color: var(--surface); border-color: var(--border)">
          <h3 class="font-medium mb-2" style="color: var(--text-strong)">Delete monitor?</h3>
          <p class="text-sm mb-5" style="color: var(--text-muted)">
            This will permanently delete <strong style="color: var(--text)">{{ monitor?.name }}</strong> and all its ping history. The ping URL will stop working immediately.
          </p>
          <div class="flex gap-3">
            <Button variant="destructive" :disabled="actionPending" @click="confirmDelete">
              {{ actionPending ? 'Deleting…' : 'Delete' }}
            </Button>
            <Button variant="secondary" @click="showDeleteConfirm = false">Cancel</Button>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>
