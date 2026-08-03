<script setup lang="ts">
import { computed } from 'vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import NotificationChannelPicker from '@/components/NotificationChannelPicker.vue'
import UptimeAdvancedSettings from '@/components/UptimeAdvancedSettings.vue'
import UptimeJsonAssertionsField from '@/components/UptimeJsonAssertionsField.vue'
import {
  alertFilterOptions,
  alertLimitOptions,
  intervalOptionsFor,
  keywordModeOptions,
  type UptimeMonitorFormState,
} from '@/lib/uptimeMonitorForm'

const form = defineModel<UptimeMonitorFormState>({ required: true })

const props = withDefaults(
  defineProps<{
    minIntervalMins?: number
    // The alerts on/off toggle only exists once a monitor does — create has
    // nothing to silence yet, so the API defaults it to enabled.
    showAlertsToggle?: boolean
    namePlaceholder?: string
  }>(),
  { minIntervalMins: 5, showAlertsToggle: false, namePlaceholder: '' },
)

const intervalOptions = computed(() => intervalOptionsFor(props.minIntervalMins))
</script>

<template>
  <div class="space-y-5">
    <div>
      <Label for="name">Name</Label>
      <Input id="name" v-model="form.name" :placeholder="namePlaceholder" class="mt-1" required />
    </div>

    <div>
      <Label for="url">URL</Label>
      <Input id="url" v-model="form.url" placeholder="https://example.com/health" class="mt-1" />
      <p class="text-xs mt-1" style="color: var(--text-muted)">
        Defaults to a GET request, 10-second timeout, HTTP 200 — customize under Advanced below.
      </p>
    </div>

    <div>
      <Label for="keyword">Keyword (optional)</Label>
      <Input
        id="keyword"
        v-model="form.keyword"
        placeholder="e.g. Welcome back"
        class="mt-1"
        maxlength="500"
      />
      <p class="text-xs mt-1" style="color: var(--text-muted)">
        Leave blank to check status code only. Searches the first 512 KB of the response body.
      </p>
    </div>

    <div
      v-if="form.keyword.trim()"
      class="space-y-4 pl-4 border-l-2"
      style="border-color: var(--border)"
    >
      <div>
        <Label for="keywordMode">Mode</Label>
        <select
          id="keywordMode"
          v-model="form.keywordMode"
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
          v-model="form.keywordCaseSensitive"
          type="checkbox"
          class="rounded"
        />
        <Label for="keywordCaseSensitive" class="cursor-pointer">Case-sensitive</Label>
      </div>
    </div>

    <UptimeJsonAssertionsField v-model="form.jsonAssertions" />

    <UptimeAdvancedSettings
      v-model:http-method="form.httpMethod"
      v-model:accepted-status-codes="form.acceptedStatusCodes"
      v-model:max-response-time-ms="form.maxResponseTimeMs"
    />

    <div>
      <Label for="interval">Check interval</Label>
      <select
        id="interval"
        v-model="form.intervalMins"
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
        v-model="form.maxAlertsPerIncident"
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
        v-model="form.alertAfterNFailures"
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

    <div v-if="showAlertsToggle" class="flex items-center gap-3">
      <input id="alerts" v-model="form.alertsEnabled" type="checkbox" class="rounded" />
      <Label for="alerts" class="cursor-pointer">Send alerts</Label>
    </div>

    <NotificationChannelPicker v-model="form.channelIds" />
  </div>
</template>
