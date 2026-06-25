<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import { monitorsApi } from '@/api/monitors'
import { ApiError } from '@/api/client'
import UpgradePrompt from '@/components/UpgradePrompt.vue'
import NotificationChannelPicker from '@/components/NotificationChannelPicker.vue'
import { useSSLMonitor } from '@/composables/useSSLMonitors'

const router = useRouter()
const route = useRoute()
const id = route.params.id as string

const name = ref('')
const hostname = ref('') // read-only; passed through on save
const alertsEnabled = ref(true)
const alertAfterNFailures = ref(0)
const maxAlertsPerIncident = ref(3)
const channelIds = ref<string[]>([])
const submitting = ref(false)
const error = ref('')
const limitReached = ref(false)

const alertLimitOptions = [
  { label: '1 time', value: 1 },
  { label: '2 times', value: 2 },
  { label: '3 times (default)', value: 3 },
  { label: '5 times', value: 5 },
  { label: '10 times', value: 10 },
]

const alertFilterOptions = [
  { label: 'Alert immediately (default)', value: 0 },
  { label: 'Skip first 1 failure', value: 1 },
  { label: 'Skip first 2 failures', value: 2 },
  { label: 'Skip first 3 failures', value: 3 },
  { label: 'Skip first 5 failures', value: 5 },
]

const { data: monitor, isPending: loading, error: loadError } = useSSLMonitor(id)
watch(
  monitor,
  (m) => {
    if (!m) return
    name.value = m.name
    hostname.value = m.hostname
    alertsEnabled.value = m.alertsEnabled
    alertAfterNFailures.value = m.alertAfterNFailures
    maxAlertsPerIncident.value = m.maxAlertsPerIncident
    channelIds.value = m.channelIds ?? []
  },
  { immediate: true },
)
watch(loadError, (e) => {
  if (e) error.value = e.message
})

async function submit() {
  error.value = ''
  limitReached.value = false
  if (!name.value.trim()) {
    error.value = 'Name is required'
    return
  }

  submitting.value = true
  try {
    await monitorsApi.updateSSL(id, {
      name: name.value.trim(),
      hostname: hostname.value,
      alertsEnabled: alertsEnabled.value,
      alertAfterNFailures: alertAfterNFailures.value,
      maxAlertsPerIncident: maxAlertsPerIncident.value,
      channelIds: channelIds.value,
    })
    router.push({ name: 'ssl-monitor-detail', params: { id } })
  } catch (e: unknown) {
    if (e instanceof ApiError && e.code === 'plan_limit_reached') {
      limitReached.value = true
      error.value = e.message
    } else {
      error.value = e instanceof Error ? e.message : 'Failed to update monitor'
    }
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <AppLayout>
    <div class="p-8 max-w-xl mx-auto">
      <div class="flex items-center gap-3 mb-6">
        <button
          class="text-sm transition-colors"
          style="color: var(--text-muted)"
          @click="router.push({ name: 'ssl-monitor-detail', params: { id } })"
        >
          ← Back
        </button>
        <h1 class="text-2xl font-semibold" style="color: var(--text-strong)">Edit SSL monitor</h1>
      </div>

      <div v-if="loading" class="text-sm" style="color: var(--text-muted)">Loading…</div>

      <form
        v-else
        class="rounded-xl border p-6 space-y-5"
        style="background-color: var(--surface); border-color: var(--border)"
        @submit.prevent="submit"
      >
        <div>
          <Label for="name">Name</Label>
          <Input id="name" v-model="name" class="mt-1" required />
        </div>

        <div>
          <Label>Hostname</Label>
          <p class="mt-1 text-sm font-mono" style="color: var(--text-dim)">{{ hostname }}</p>
          <p class="text-xs mt-0.5" style="color: var(--text-muted)">
            To change the domain, delete this monitor and create a new one.
          </p>
        </div>

        <div class="flex items-center gap-3">
          <input id="alerts" v-model="alertsEnabled" type="checkbox" class="rounded" />
          <Label for="alerts" class="cursor-pointer">Send alerts</Label>
        </div>

        <div>
          <Label for="alertLimit">Alert limit per expiry cycle</Label>
          <select
            id="alertLimit"
            v-model="maxAlertsPerIncident"
            class="mt-1 w-full rounded-md border px-3 py-2 text-sm"
            style="background-color: var(--surface-raised); border-color: var(--border); color: var(--text)"
          >
            <option v-for="opt in alertLimitOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </option>
          </select>
          <p class="text-xs mt-1" style="color: var(--text-muted)">
            Stop alerting after this many notifications. Resets on renewal.
          </p>
        </div>

        <div>
          <Label for="alertFilter">Alert filter</Label>
          <select
            id="alertFilter"
            v-model="alertAfterNFailures"
            class="mt-1 w-full rounded-md border px-3 py-2 text-sm"
            style="background-color: var(--surface-raised); border-color: var(--border); color: var(--text)"
          >
            <option v-for="opt in alertFilterOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </option>
          </select>
          <p class="text-xs mt-1" style="color: var(--text-muted)">
            Suppress alerts until N consecutive check cycles detect the issue. Resets on renewal.
          </p>
        </div>

        <NotificationChannelPicker v-model="channelIds" />

        <UpgradePrompt v-if="limitReached" :message="error" />
        <p v-else-if="error" class="text-sm" style="color: var(--status-down)">{{ error }}</p>

        <div class="flex gap-3 pt-1">
          <Button type="submit" :disabled="submitting">
            {{ submitting ? 'Saving…' : 'Save changes' }}
          </Button>
          <Button
            variant="secondary"
            type="button"
            @click="router.push({ name: 'ssl-monitor-detail', params: { id } })"
          >
            Cancel
          </Button>
        </div>
      </form>
    </div>
  </AppLayout>
</template>
