<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import { monitorsApi, type ExpectedState } from '@/api/monitors'
import { ApiError } from '@/api/client'
import UpgradePrompt from '@/components/UpgradePrompt.vue'
import NotificationChannelPicker from '@/components/NotificationChannelPicker.vue'
import { useBilling } from '@/composables/useBilling'

const router = useRouter()

const name = ref('')
const host = ref('')
const port = ref<number | null>(null)
const expectedState = ref<ExpectedState>('open')
const intervalMins = ref(10)
const maxAlertsPerIncident = ref(3)
const alertAfterNFailures = ref(0)
const channelIds = ref<string[]>([])
const submitting = ref(false)
const error = ref('')
const limitReached = ref(false)
const minIntervalMins = ref(5)

const expectedStateOptions: { label: string; value: ExpectedState }[] = [
  { label: 'Open — alert if it stops accepting connections', value: 'open' },
  { label: 'Closed — alert if it becomes reachable', value: 'closed' },
]

const intervalOptions = computed(() => [
  ...(minIntervalMins.value === 1 ? [{ label: '1 minute', value: 1 }] : []),
  { label: '5 minutes', value: 5 },
  { label: '10 minutes', value: 10 },
  { label: '30 minutes', value: 30 },
])

const { data: billingInfo } = useBilling()
watch(
  billingInfo,
  (info) => {
    if (!info) return
    minIntervalMins.value = info.minIntervalMins
  },
  { immediate: true },
)

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

// validatePortForm returns the first validation error, or null if the form
// is ready to submit — kept separate from submit() to keep its complexity down.
function validatePortForm(): string | null {
  if (!name.value.trim()) return 'Name is required'
  if (!host.value.trim()) return 'Host is required'
  if (!port.value || port.value < 1 || port.value > 65535) return 'Port must be between 1 and 65535'
  return null
}

async function submit() {
  error.value = ''
  limitReached.value = false
  const validationError = validatePortForm()
  if (validationError) {
    error.value = validationError
    return
  }

  submitting.value = true
  try {
    const monitor = await monitorsApi.createPort({
      name: name.value.trim(),
      host: host.value.trim(),
      port: port.value,
      expectedState: expectedState.value,
      intervalMins: intervalMins.value,
      maxAlertsPerIncident: maxAlertsPerIncident.value,
      alertAfterNFailures: alertAfterNFailures.value,
      channelIds: channelIds.value,
    })
    router.push({ name: 'port-monitor-detail', params: { id: monitor.id } })
  } catch (e: unknown) {
    if (e instanceof ApiError && e.code === 'plan_limit_reached') {
      limitReached.value = true
      error.value = e.message
    } else {
      error.value = e instanceof Error ? e.message : 'Failed to create monitor'
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
          @click="router.push({ name: 'port-monitors' })"
        >
          ← Back
        </button>
        <h1 class="text-2xl font-semibold" style="color: var(--text-strong)">New port monitor</h1>
      </div>

      <form
        class="rounded-xl border p-6 space-y-5"
        style="background-color: var(--surface); border-color: var(--border)"
        @submit.prevent="submit"
      >
        <div>
          <Label for="name">Name</Label>
          <Input
            id="name"
            v-model="name"
            placeholder="Mail server"
            class="mt-1"
            required
          />
        </div>

        <div class="grid grid-cols-3 gap-3">
          <div class="col-span-2">
            <Label for="host">Host</Label>
            <Input
              id="host"
              v-model="host"
              placeholder="mail.example.com"
              class="mt-1"
            />
          </div>
          <div>
            <Label for="port">Port</Label>
            <Input
              id="port"
              v-model.number="port"
              type="number"
              min="1"
              max="65535"
              placeholder="25"
              class="mt-1"
            />
          </div>
        </div>
        <p class="text-xs -mt-3" style="color: var(--text-muted)">
          Raw TCP connect — no data sent or received. 10-second timeout.
        </p>

        <div>
          <Label for="expectedState">Expected state</Label>
          <select
            id="expectedState"
            v-model="expectedState"
            class="mt-1 w-full rounded-md border px-3 py-2 text-sm"
            style="background-color: var(--surface-raised); border-color: var(--border); color: var(--text)"
          >
            <option v-for="opt in expectedStateOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </option>
          </select>
          <p class="text-xs mt-1" style="color: var(--text-muted)">
            "Closed" is a security check — confirm a port that should be firewalled off stays unreachable.
          </p>
        </div>

        <div>
          <Label for="interval">Check interval</Label>
          <select
            id="interval"
            v-model="intervalMins"
            class="mt-1 w-full rounded-md border px-3 py-2 text-sm"
            style="background-color: var(--surface-raised); border-color: var(--border); color: var(--text)"
          >
            <option v-for="opt in intervalOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </option>
          </select>
        </div>

        <div>
          <Label for="alertLimit">Alert limit per incident</Label>
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
            Stop alerting after this many notifications per incident.
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
            Suppress alerts until N consecutive failures are detected. Resets on success.
          </p>
        </div>

        <NotificationChannelPicker v-model="channelIds" />

        <UpgradePrompt v-if="limitReached" :message="error" />
        <p v-else-if="error" class="text-sm" style="color: var(--status-down)">{{ error }}</p>

        <div class="flex gap-3 pt-1">
          <Button type="submit" :disabled="submitting">
            {{ submitting ? 'Creating…' : 'Create monitor' }}
          </Button>
          <Button variant="secondary" type="button" @click="router.push({ name: 'port-monitors' })">
            Cancel
          </Button>
        </div>
      </form>
    </div>
  </AppLayout>
</template>
