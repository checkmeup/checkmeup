<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import { monitorsApi, type DNSRecordType } from '@/api/monitors'
import { ApiError } from '@/api/client'
import UpgradePrompt from '@/components/UpgradePrompt.vue'
import NotificationChannelPicker from '@/components/NotificationChannelPicker.vue'
import { useDNSMonitor } from '@/composables/useDNSMonitors'
import { useBilling } from '@/composables/useBilling'

const router = useRouter()
const route = useRoute()
const id = route.params.id as string

const name = ref('')
const hostname = ref('')
const recordType = ref<DNSRecordType>('A')
const expectedValue = ref('')
const intervalMins = ref(10)
const alertsEnabled = ref(true)
const maxAlertsPerIncident = ref(3)
const alertAfterNFailures = ref(0)
const channelIds = ref<string[]>([])
const submitting = ref(false)
const error = ref('')
const minIntervalMins = ref(5)
const limitReached = ref(false)

const recordTypeOptions: { label: string; value: DNSRecordType }[] = [
  { label: 'A — IPv4 address', value: 'A' },
  { label: 'AAAA — IPv6 address', value: 'AAAA' },
  { label: 'CNAME — canonical name', value: 'CNAME' },
  { label: 'MX — mail exchange', value: 'MX' },
  { label: 'TXT — text record', value: 'TXT' },
  { label: 'NS — nameserver', value: 'NS' },
]

const intervalOptions = computed(() => [
  ...(minIntervalMins.value === 1 ? [{ label: '1 minute', value: 1 }] : []),
  { label: '5 minutes', value: 5 },
  { label: '10 minutes', value: 10 },
  { label: '30 minutes', value: 30 },
])

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

const { data: detail, isPending: monitorLoading, error: loadError } = useDNSMonitor(id)
const { data: billingInfo, isPending: billingLoading } = useBilling()
const loading = computed(() => monitorLoading.value || billingLoading.value)

// recordTypeAtLoad tracks the value fetched from the server so changing the
// record type in the form (a different, no-longer-comparable value shape)
// can clear the stale expected value client-side without also clobbering it
// on the initial population watch below.
let recordTypeAtLoad: DNSRecordType | null = null

let formPopulated = false
watch(
  detail,
  (d) => {
    if (!d || formPopulated) return
    formPopulated = true
    const m = d.monitor
    name.value = m.name
    hostname.value = m.hostname
    recordType.value = m.recordType
    recordTypeAtLoad = m.recordType
    expectedValue.value = m.expectedValue ?? ''
    intervalMins.value = m.intervalMins
    alertsEnabled.value = m.alertsEnabled
    maxAlertsPerIncident.value = m.maxAlertsPerIncident
    alertAfterNFailures.value = m.alertAfterNFailures
    channelIds.value = m.channelIds ?? []
  },
  { immediate: true },
)
watch(recordType, (newType) => {
  if (formPopulated && recordTypeAtLoad !== null && newType !== recordTypeAtLoad) {
    expectedValue.value = ''
  }
})
watch(
  billingInfo,
  (info) => {
    if (!info) return
    minIntervalMins.value = info.minIntervalMins
  },
  { immediate: true },
)
watch(loadError, (e) => {
  if (e) error.value = e.message
})

// validateDNSForm returns the first validation error, or null if the form
// is ready to submit — kept separate from submit() to keep its complexity down.
function validateDNSForm(): string | null {
  if (!name.value.trim()) return 'Name is required'
  if (!hostname.value.trim()) return 'Hostname is required'
  return null
}

async function submit() {
  error.value = ''
  limitReached.value = false
  const validationError = validateDNSForm()
  if (validationError) {
    error.value = validationError
    return
  }

  submitting.value = true
  try {
    await monitorsApi.updateDns(id, {
      name: name.value.trim(),
      hostname: hostname.value.trim(),
      recordType: recordType.value,
      expectedValue: expectedValue.value.trim(),
      intervalMins: intervalMins.value,
      alertsEnabled: alertsEnabled.value,
      maxAlertsPerIncident: maxAlertsPerIncident.value,
      alertAfterNFailures: alertAfterNFailures.value,
      channelIds: channelIds.value,
    })
    router.push({ name: 'dns-monitor-detail', params: { id } })
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
          @click="router.push({ name: 'dns-monitor-detail', params: { id } })"
        >
          ← Back
        </button>
        <h1 class="text-2xl font-semibold" style="color: var(--text-strong)">Edit monitor</h1>
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

        <div class="grid grid-cols-3 gap-3">
          <div class="col-span-2">
            <Label for="hostname">Hostname</Label>
            <Input id="hostname" v-model="hostname" class="mt-1" />
          </div>
          <div>
            <Label for="recordType">Record type</Label>
            <select
              id="recordType"
              v-model="recordType"
              class="mt-1 w-full rounded-md border px-3 py-2 text-sm"
              style="background-color: var(--surface-raised); border-color: var(--border); color: var(--text)"
            >
              <option v-for="opt in recordTypeOptions" :key="opt.value" :value="opt.value">
                {{ opt.value }}
              </option>
            </select>
          </div>
        </div>

        <div>
          <Label for="expectedValue">Expected value (optional)</Label>
          <Input
            id="expectedValue"
            v-model="expectedValue"
            placeholder="e.g. 203.0.113.10"
            class="mt-1"
          />
          <p class="text-xs mt-1" style="color: var(--text-muted)">
            Clear this to re-arm baseline mode — the next check re-captures whatever the record currently resolves to. Save a new value to acknowledge an intentional change.
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

        <div class="flex items-center gap-3">
          <input id="alerts" v-model="alertsEnabled" type="checkbox" class="rounded" />
          <Label for="alerts" class="cursor-pointer">Send alerts</Label>
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
            {{ submitting ? 'Saving…' : 'Save changes' }}
          </Button>
          <Button
            variant="secondary"
            type="button"
            @click="router.push({ name: 'dns-monitor-detail', params: { id } })"
          >
            Cancel
          </Button>
        </div>
      </form>
    </div>
  </AppLayout>
</template>
