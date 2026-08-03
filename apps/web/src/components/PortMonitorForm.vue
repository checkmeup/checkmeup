<script setup lang="ts">
import { computed } from 'vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import NotificationChannelPicker from '@/components/NotificationChannelPicker.vue'
import {
  alertFilterOptions,
  alertLimitOptions,
  expectedStateOptions,
  intervalOptionsFor,
  type PortMonitorFormState,
} from '@/lib/portMonitorForm'

const form = defineModel<PortMonitorFormState>({ required: true })

const props = withDefaults(
  defineProps<{
    minIntervalMins?: number
    // The alerts on/off toggle only exists once a monitor does — create has
    // nothing to silence yet, so the API defaults it to enabled.
    showAlertsToggle?: boolean
  }>(),
  { minIntervalMins: 5, showAlertsToggle: false },
)

const intervalOptions = computed(() => intervalOptionsFor(props.minIntervalMins))

// Input's modelValue is string-typed (it doesn't implement Vue's
// modelModifiers convention, so a bare `v-model.number` silently does no
// numeric conversion) — bridge it to the numeric field explicitly.
const portInput = computed({
  get: () => form.value.port?.toString() ?? '',
  set: (v: string) => {
    form.value.port = v === '' ? undefined : Number(v)
  },
})
</script>

<template>
  <div class="space-y-5">
    <div>
      <Label for="name">Name</Label>
      <Input id="name" v-model="form.name" placeholder="Mail server" class="mt-1" required />
    </div>

    <div class="grid grid-cols-3 gap-3">
      <div class="col-span-2">
        <Label for="host">Host</Label>
        <Input id="host" v-model="form.host" placeholder="mail.example.com" class="mt-1" />
      </div>
      <div>
        <Label for="port">Port</Label>
        <Input
          id="port"
          v-model="portInput"
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
        v-model="form.expectedState"
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
