<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import MaintenanceMonitorPicker from '@/components/MaintenanceMonitorPicker.vue'
import { incidentsApi, type IncidentSeverity } from '@/api/incidents'
import { ApiError } from '@/api/client'

const router = useRouter()

const title = ref('')
const message = ref('')
const severity = ref<IncidentSeverity>('minor')
const monitors = ref<{ monitorType: 'cron' | 'uptime' | 'ssl' | 'domain' | 'port'; monitorId: string; name: string }[]>([])
const submitting = ref(false)
const error = ref('')
const overlapWarning = ref('')

const severityOptions: { value: IncidentSeverity; label: string }[] = [
  { value: 'minor', label: 'Minor — informational, no outage' },
  { value: 'major', label: 'Major — partial outage' },
  { value: 'critical', label: 'Critical — full outage' },
]

function validate(): boolean {
  if (!title.value.trim()) {
    error.value = 'Title is required'
    return false
  }
  if (!message.value.trim()) {
    error.value = 'An initial message is required'
    return false
  }
  if (monitors.value.length === 0) {
    error.value = 'Select at least one affected monitor'
    return false
  }
  return true
}

async function submit(confirmOverlap = false) {
  error.value = ''
  overlapWarning.value = ''
  if (!validate()) return

  submitting.value = true
  try {
    await incidentsApi.create({
      title: title.value.trim(),
      message: message.value.trim(),
      severity: severity.value,
      monitors: monitors.value.map((m) => ({ monitorType: m.monitorType, monitorId: m.monitorId })),
      confirmOverlap,
    })
    router.push({ name: 'incidents' })
  } catch (e: unknown) {
    if (e instanceof ApiError && e.code === 'maintenance_overlap') {
      overlapWarning.value = e.message
    } else {
      error.value = e instanceof Error ? e.message : 'Failed to declare incident'
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
          class="text-sm"
          style="color: var(--text-muted)"
          @click="router.push({ name: 'incidents' })"
        >
          ← Back
        </button>
        <h1 class="text-2xl font-semibold" style="color: var(--text-strong)">Declare incident</h1>
      </div>

      <form
        class="rounded-xl border p-6 space-y-5"
        style="background-color: var(--surface); border-color: var(--border)"
        @submit.prevent="submit(false)"
      >
        <div>
          <Label for="title">Title</Label>
          <Input id="title" v-model="title" placeholder="Elevated API latency" class="mt-1" required />
        </div>

        <div>
          <Label for="message">Initial message</Label>
          <Input
            id="message"
            v-model="message"
            placeholder="We're investigating reports of slow API responses."
            class="mt-1"
          />
        </div>

        <div>
          <Label for="severity">Severity</Label>
          <select
            id="severity"
            v-model="severity"
            class="mt-1 w-full rounded-md border px-3 py-2 text-sm"
            style="background-color: var(--surface-raised); border-color: var(--border); color: var(--text)"
          >
            <option v-for="opt in severityOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
          </select>
        </div>

        <div>
          <Label>Affected monitors</Label>
          <MaintenanceMonitorPicker v-model="monitors" class="mt-1" />
        </div>

        <div
          v-if="overlapWarning"
          class="rounded-lg border p-4 text-sm space-y-3"
          style="background-color: var(--surface-raised); border-color: var(--status-degraded)"
        >
          <p style="color: var(--status-degraded)">{{ overlapWarning }}</p>
          <div class="flex gap-2">
            <Button type="button" :disabled="submitting" @click="submit(true)">
              {{ submitting ? 'Declaring…' : 'Declare anyway' }}
            </Button>
            <Button variant="secondary" type="button" @click="overlapWarning = ''">Cancel</Button>
          </div>
        </div>
        <p v-else-if="error" class="text-sm" style="color: var(--status-down)">{{ error }}</p>

        <div v-if="!overlapWarning" class="flex gap-3 pt-1">
          <Button type="submit" :disabled="submitting">
            {{ submitting ? 'Declaring…' : 'Declare incident' }}
          </Button>
          <Button variant="secondary" type="button" @click="router.push({ name: 'incidents' })">
            Cancel
          </Button>
        </div>
      </form>
    </div>
  </AppLayout>
</template>
