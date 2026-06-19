<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import { monitorsApi, type KeywordMode } from '@/api/monitors'
import { billingApi } from '@/api/billing'
import { ApiError } from '@/api/client'
import UpgradePrompt from '@/components/UpgradePrompt.vue'

const router = useRouter()
const route = useRoute()
const id = route.params.id as string

const name = ref('')
const url = ref('')
const intervalMins = ref(10)
const alertsEnabled = ref(true)
const maxAlertsPerIncident = ref(3)
const keyword = ref('')
const keywordMode = ref<KeywordMode>('contains')
const keywordCaseSensitive = ref(false)
const loading = ref(true)
const submitting = ref(false)
const error = ref('')
const minIntervalMins = ref(5)
const keywordMonitoringEnabled = ref(false)
const limitReached = ref(false)

const keywordModeOptions: { label: string; value: KeywordMode }[] = [
  { label: 'Contains', value: 'contains' },
  { label: 'Does not contain', value: 'not_contains' },
]

// A keyword set while on a paid plan stays editable-to-clear after a
// downgrade, but can't be changed to new text until re-upgrading — same
// "keep what you have, can't add more" policy as monitor/status-page limits.
const keywordLocked = computed(() => !keywordMonitoringEnabled.value && !!keyword.value)

const intervalOptions = computed(() => [
  ...(minIntervalMins.value === 1 ? [{ label: '1 minute', value: 1 }] : []),
  { label: '5 minutes', value: 5 },
  { label: '10 minutes', value: 10 },
  { label: '30 minutes', value: 30 },
])

const alertLimitOptions = [
  { label: 'Always alert', value: 0 },
  { label: '1 time', value: 1 },
  { label: '2 times', value: 2 },
  { label: '3 times (default)', value: 3 },
  { label: '5 times', value: 5 },
  { label: '10 times', value: 10 },
]

onMounted(async () => {
  try {
    const [detail, info] = await Promise.all([monitorsApi.getUptime(id), billingApi.getInfo()])
    const m = detail.monitor
    name.value = m.name
    url.value = m.url
    intervalMins.value = m.intervalMins
    alertsEnabled.value = m.alertsEnabled
    maxAlertsPerIncident.value = m.maxAlertsPerIncident
    keyword.value = m.keyword ?? ''
    keywordMode.value = m.keywordMode
    keywordCaseSensitive.value = m.keywordCaseSensitive
    minIntervalMins.value = info.minIntervalMins
    keywordMonitoringEnabled.value = info.keywordMonitoringEnabled
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load monitor'
  } finally {
    loading.value = false
  }
})

async function submit() {
  error.value = ''
  limitReached.value = false
  if (!name.value.trim()) {
    error.value = 'Name is required'
    return
  }
  if (!url.value.trim() || !url.value.match(/^https?:\/\//)) {
    error.value = 'URL must start with http:// or https://'
    return
  }
  if (keyword.value.trim().length > 500) {
    error.value = 'Keyword must be 500 characters or fewer'
    return
  }

  submitting.value = true
  try {
    await monitorsApi.updateUptime(id, {
      name: name.value.trim(),
      url: url.value.trim(),
      intervalMins: intervalMins.value,
      alertsEnabled: alertsEnabled.value,
      maxAlertsPerIncident: maxAlertsPerIncident.value,
      keyword: keyword.value.trim(),
      keywordMode: keywordMode.value,
      keywordCaseSensitive: keywordCaseSensitive.value,
    })
    router.push({ name: 'uptime-monitor-detail', params: { id } })
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
          @click="router.push({ name: 'uptime-monitor-detail', params: { id } })"
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

        <div>
          <Label for="url">URL</Label>
          <Input id="url" v-model="url" placeholder="https://example.com/health" class="mt-1" />
          <p class="text-xs mt-1" style="color: var(--text-muted)">
            Must return HTTP 200. GET request, 10-second timeout.
          </p>
        </div>

        <UpgradePrompt
          v-if="!keywordMonitoringEnabled"
          :message="
            keywordLocked
              ? 'Your keyword check is paused — keyword monitoring is available on paid plans. You can still clear it below.'
              : 'Keyword monitoring is available on paid plans.'
          "
        />

        <div>
          <Label for="keyword">Keyword (optional)</Label>
          <div class="flex items-start gap-2 mt-1">
            <Input
              id="keyword"
              v-model="keyword"
              placeholder="e.g. Welcome back"
              class="flex-1"
              maxlength="500"
              :disabled="keywordLocked"
            />
            <Button v-if="keywordLocked" type="button" variant="secondary" @click="keyword = ''">
              Clear
            </Button>
          </div>
          <p class="text-xs mt-1" style="color: var(--text-muted)">
            Leave blank to check status code only. Searches the first 512 KB of the response body.
          </p>
        </div>

        <div v-if="keyword.trim() && keywordMonitoringEnabled" class="space-y-4 pl-4 border-l-2" style="border-color: var(--border)">
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

        <UpgradePrompt v-if="limitReached" :message="error" />
        <p v-else-if="error" class="text-sm" style="color: var(--status-down)">{{ error }}</p>

        <div class="flex gap-3 pt-1">
          <Button type="submit" :disabled="submitting">
            {{ submitting ? 'Saving…' : 'Save changes' }}
          </Button>
          <Button
            variant="secondary"
            type="button"
            @click="router.push({ name: 'uptime-monitor-detail', params: { id } })"
          >
            Cancel
          </Button>
        </div>
      </form>
    </div>
  </AppLayout>
</template>
