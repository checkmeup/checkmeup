<script setup lang="ts">
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import NotificationChannelPicker from '@/components/NotificationChannelPicker.vue'
import {
  alertFilterOptions,
  alertLimitOptions,
  graceOptions,
  maxDurationOptions,
  type CronMonitorFormState,
} from '@/lib/cronMonitorForm'

const form = defineModel<CronMonitorFormState>({ required: true })

withDefaults(
  defineProps<{
    // The alerts on/off toggle only exists once a monitor does — create has
    // nothing to silence yet, so the API defaults it to enabled.
    showAlertsToggle?: boolean
    namePlaceholder?: string
  }>(),
  { showAlertsToggle: false, namePlaceholder: '' },
)

// Two slots rather than props, because what differs between create and edit
// here is genuinely different content, not a string swap: create offers
// clickable schedule examples, edit explains that the ping URL is stable and
// renders the monitor's real /start URL.
defineSlots<{
  scheduleHelp?: () => unknown
  maxDurationHelp?: () => unknown
}>()
</script>

<template>
  <div class="space-y-5">
    <div>
      <Label for="name">Name</Label>
      <Input id="name" v-model="form.name" :placeholder="namePlaceholder" class="mt-1" required />
    </div>

    <div>
      <Label for="schedule">Schedule</Label>
      <Input
        id="schedule"
        v-model="form.schedule"
        placeholder="every 1h  or  0 * * * *"
        class="mt-1"
      />
      <slot name="scheduleHelp" />
    </div>

    <div>
      <Label for="grace">Grace period</Label>
      <select
        id="grace"
        v-model="form.gracePeriodMins"
        class="mt-1 w-full rounded-md border px-3 py-2 text-sm"
        style="background-color: var(--surface-raised); border-color: var(--border); color: var(--text)"
      >
        <option v-for="opt in graceOptions" :key="opt.value" :value="opt.value">
          {{ opt.label }}
        </option>
      </select>
      <p class="text-xs mt-1" style="color: var(--text-muted)">
        How long to wait after the expected time before alerting.
      </p>
    </div>

    <div v-if="showAlertsToggle" class="flex items-center gap-3">
      <input id="alerts" v-model="form.alertsEnabled" type="checkbox" class="rounded" />
      <Label for="alerts" class="cursor-pointer">Send alerts</Label>
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

    <div>
      <Label for="maxDuration">Max run duration</Label>
      <select
        id="maxDuration"
        v-model="form.maxDurationMins"
        class="mt-1 w-full rounded-md border px-3 py-2 text-sm"
        style="background-color: var(--surface-raised); border-color: var(--border); color: var(--text)"
      >
        <option v-for="opt in maxDurationOptions" :key="opt.label" :value="opt.value">
          {{ opt.label }}
        </option>
      </select>
      <slot name="maxDurationHelp" />
    </div>

    <NotificationChannelPicker v-model="form.channelIds" />
  </div>
</template>
