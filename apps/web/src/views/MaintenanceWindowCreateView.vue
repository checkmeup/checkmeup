<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import MonitorFormPage from '@/components/MonitorFormPage.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import MaintenanceMonitorPicker from '@/components/MaintenanceMonitorPicker.vue'
import { maintenanceApi } from '@/api/maintenance'

const router = useRouter()

const title = ref('')
const message = ref('')
const startsAt = ref('')
const noEnd = ref(false)
const endsAt = ref('')
const monitors = ref<{ monitorType: 'cron' | 'uptime' | 'ssl' | 'domain' | 'port'; monitorId: string; name: string }[]>([])
const submitting = ref(false)
const error = ref('')

function toIso(local: string): string {
  return new Date(local).toISOString()
}

function validateMaintenanceWindowForm(): string {
  if (!title.value.trim()) return 'Title is required'
  if (!startsAt.value) return 'Start time is required'
  if (!noEnd.value && !endsAt.value) return 'End time is required, or check "no end date"'
  if (monitors.value.length === 0) return 'Select at least one monitor'
  return ''
}

async function submit() {
  error.value = validateMaintenanceWindowForm()
  if (error.value) return

  submitting.value = true
  try {
    await maintenanceApi.create({
      title: title.value.trim(),
      message: message.value.trim(),
      startsAt: toIso(startsAt.value),
      endsAt: noEnd.value ? null : toIso(endsAt.value),
      monitors: monitors.value.map((m) => ({ monitorType: m.monitorType, monitorId: m.monitorId })),
    })
    router.push({ name: 'maintenance' })
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to create maintenance window'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <MonitorFormPage
    title="Schedule maintenance"
    :back-to="{ name: 'maintenance' }"
    submit-label="Schedule maintenance"
    submitting-label="Scheduling…"
    :error="error"
    :submitting="submitting"
    @submit="submit"
  >
    <div>
      <Label for="title">Title</Label>
      <Input id="title" v-model="title" placeholder="Database migration" class="mt-1" required />
    </div>

    <div>
      <Label for="message">
        Message <span style="color: var(--text-muted)">(optional, shown on status page)</span>
      </Label>
      <Input
        id="message"
        v-model="message"
        placeholder="We're upgrading our database, expect brief downtime."
        class="mt-1"
      />
    </div>

    <div class="grid grid-cols-2 gap-4">
      <div>
        <Label for="startsAt">Starts</Label>
        <input
          id="startsAt"
          v-model="startsAt"
          type="datetime-local"
          class="mt-1 flex h-10 w-full rounded-md border px-3 py-2 text-sm focus:outline-none focus:ring-2"
          style="background-color: var(--surface); border-color: var(--border); color: var(--text)"
          required
        />
      </div>
      <div>
        <Label for="endsAt">Ends</Label>
        <input
          id="endsAt"
          v-model="endsAt"
          type="datetime-local"
          :disabled="noEnd"
          class="mt-1 flex h-10 w-full rounded-md border px-3 py-2 text-sm focus:outline-none focus:ring-2 disabled:opacity-50"
          style="background-color: var(--surface); border-color: var(--border); color: var(--text)"
        />
        <label class="flex items-center gap-2 mt-1.5 text-xs" style="color: var(--text-muted)">
          <input v-model="noEnd" type="checkbox" />
          No end date — I'll end it manually
        </label>
      </div>
    </div>

    <div>
      <Label>Monitors</Label>
      <MaintenanceMonitorPicker v-model="monitors" class="mt-1" />
    </div>
  </MonitorFormPage>
</template>
