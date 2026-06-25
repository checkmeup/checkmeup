<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import { monitorsApi, type KeywordMode, type JsonAssertion, type AssertionComparator } from '@/api/monitors'
import { ApiError } from '@/api/client'
import UpgradePrompt from '@/components/UpgradePrompt.vue'
import NotificationChannelPicker from '@/components/NotificationChannelPicker.vue'
import { useBilling } from '@/composables/useBilling'

const router = useRouter()

const name = ref('')
const url = ref('')
const intervalMins = ref(10)
const maxAlertsPerIncident = ref(3)
const alertAfterNFailures = ref(0)
const channelIds = ref<string[]>([])
const keyword = ref('')
const keywordMode = ref<KeywordMode>('contains')
const keywordCaseSensitive = ref(false)
const jsonAssertions = ref<JsonAssertion[]>([])
const maxResponseTimeMs = ref<number | null>(null)
const submitting = ref(false)
const error = ref('')
const limitReached = ref(false)
const minIntervalMins = ref(5)

const keywordModeOptions: { label: string; value: KeywordMode }[] = [
  { label: 'Contains', value: 'contains' },
  { label: 'Does not contain', value: 'not_contains' },
]

const comparatorOptions: { label: string; value: AssertionComparator }[] = [
  { label: 'equals', value: 'equals' },
  { label: 'not equals', value: 'not_equals' },
  { label: 'contains', value: 'contains' },
  { label: '>', value: 'greater_than' },
  { label: '<', value: 'less_than' },
]

function addAssertion() {
  jsonAssertions.value.push({ path: '', comparator: 'equals', expected: '' })
}

function removeAssertion(i: number) {
  jsonAssertions.value.splice(i, 1)
}

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
// Load failures stay silent (keep defaults), same as the original try/catch.

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

async function submit() {
  error.value = ''
  limitReached.value = false
  if (!name.value.trim()) {
    error.value = 'Name is required'
    return
  }
  if (!url.value.trim()) {
    error.value = 'URL is required'
    return
  }
  if (!url.value.match(/^https?:\/\//)) {
    error.value = 'URL must start with http:// or https://'
    return
  }
  if (keyword.value.trim().length > 500) {
    error.value = 'Keyword must be 500 characters or fewer'
    return
  }

  submitting.value = true
  try {
    const monitor = await monitorsApi.createUptime({
      name: name.value.trim(),
      url: url.value.trim(),
      intervalMins: intervalMins.value,
      maxAlertsPerIncident: maxAlertsPerIncident.value,
      alertAfterNFailures: alertAfterNFailures.value,
      keyword: keyword.value.trim(),
      keywordMode: keywordMode.value,
      keywordCaseSensitive: keywordCaseSensitive.value,
      jsonAssertions: jsonAssertions.value,
      maxResponseTimeMs: maxResponseTimeMs.value,
      channelIds: channelIds.value,
    })
    router.push({ name: 'uptime-monitor-detail', params: { id: monitor.id } })
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
          @click="router.push({ name: 'uptime-monitors' })"
        >
          ← Back
        </button>
        <h1 class="text-2xl font-semibold" style="color: var(--text-strong)">New uptime monitor</h1>
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
            placeholder="Production API"
            class="mt-1"
            required
          />
        </div>

        <div>
          <Label for="url">URL</Label>
          <Input
            id="url"
            v-model="url"
            placeholder="https://example.com/health"
            class="mt-1"
          />
          <p class="text-xs mt-1" style="color: var(--text-muted)">
            Must return HTTP 200. GET request, 10-second timeout.
          </p>
        </div>

        <div>
          <Label for="keyword">Keyword (optional)</Label>
          <Input
            id="keyword"
            v-model="keyword"
            placeholder="e.g. Welcome back"
            class="mt-1"
            maxlength="500"
          />
          <p class="text-xs mt-1" style="color: var(--text-muted)">
            Leave blank to check status code only. Searches the first 512 KB of the response body.
          </p>
        </div>

        <div v-if="keyword.trim()" class="space-y-4 pl-4 border-l-2" style="border-color: var(--border)">
          <div>
            <Label for="keywordMode">Mode</Label>
            <select
              id="keywordMode"
              v-model="keywordMode"
              class="mt-1 w-full rounded-md border px-3 py-2 text-sm"
              style="background-color: var(--surface-raised); border-color: var(--border); color: var(--text)"
            >
              <option v-for="opt in keywordModeOptions" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </option>
            </select>
          </div>

          <div class="flex items-center gap-3">
            <input
              id="keywordCaseSensitive"
              v-model="keywordCaseSensitive"
              type="checkbox"
              class="rounded"
            />
            <Label for="keywordCaseSensitive" class="cursor-pointer">Case-sensitive</Label>
          </div>
        </div>

        <div>
          <div class="flex items-center justify-between mb-2">
            <Label>JSON assertions (optional)</Label>
            <button
              type="button"
              class="text-xs px-2 py-1 rounded"
              style="color: var(--text-dim); background-color: var(--surface-raised)"
              @click="addAssertion"
            >
              + Add
            </button>
          </div>
          <div v-if="jsonAssertions.length === 0" class="text-xs" style="color: var(--text-muted)">
            Assert on JSON response fields, e.g. <code>data.status</code> equals <code>ok</code>.
          </div>
          <div v-for="(a, i) in jsonAssertions" :key="i" class="flex items-center gap-2 mt-2">
            <Input v-model="a.path" placeholder="$.status" class="flex-1 min-w-0" />
            <select
              v-model="a.comparator"
              class="rounded-md border px-2 py-2 text-sm"
              style="background-color: var(--surface-raised); border-color: var(--border); color: var(--text)"
            >
              <option v-for="opt in comparatorOptions" :key="opt.value" :value="opt.value">
                {{ opt.label }}
              </option>
            </select>
            <Input v-model="a.expected" placeholder="ok" class="flex-1 min-w-0" />
            <button
              type="button"
              class="text-xs px-1.5 py-1 rounded flex-shrink-0"
              style="color: var(--text-muted); background-color: var(--surface-raised)"
              @click="removeAssertion(i)"
            >
              ✕
            </button>
          </div>
        </div>

        <div>
          <Label for="maxResponseTimeMs">Max response time (optional)</Label>
          <div class="flex items-center gap-2 mt-1">
            <Input
              id="maxResponseTimeMs"
              v-model.number="maxResponseTimeMs"
              type="number"
              min="1"
              placeholder="e.g. 2000"
              class="w-40"
            />
            <span class="text-sm" style="color: var(--text-muted)">ms</span>
          </div>
          <p class="text-xs mt-1" style="color: var(--text-muted)">
            Fail the check if response takes longer than this, regardless of status code.
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
          <Button variant="secondary" type="button" @click="router.push({ name: 'uptime-monitors' })">
            Cancel
          </Button>
        </div>
      </form>
    </div>
  </AppLayout>
</template>
