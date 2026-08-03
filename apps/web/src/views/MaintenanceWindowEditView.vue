<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import MonitorFormPage from '@/components/MonitorFormPage.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import MaintenanceMonitorPicker from '@/components/MaintenanceMonitorPicker.vue'
import { maintenanceApi } from '@/api/maintenance'
import { ApiError } from '@/api/client'
import { useMaintenanceWindow } from '@/composables/useMaintenanceWindows'

const router = useRouter()
const route = useRoute()
const id = route.params.id as string

const title = ref('')
const message = ref('')
const startsAt = ref('')
const noEnd = ref(false)
const endsAt = ref('')
const monitors = ref<{ monitorType: 'cron' | 'uptime' | 'ssl' | 'domain' | 'port'; monitorId: string; name: string }[]>([])
const submitting = ref(false)
const error = ref('')
const confirmDelete = ref(false)
const limitReached = ref(false)

function toLocalInputValue(iso: string): string {
  const d = new Date(iso)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function toIso(local: string): string {
  return new Date(local).toISOString()
}

const { data: win, isPending: loading, error: loadError } = useMaintenanceWindow(id)
watch(
  win,
  (w) => {
    if (!w) return
    title.value = w.title
    message.value = w.message
    startsAt.value = toLocalInputValue(w.startsAt)
    if (w.endsAt) {
      endsAt.value = toLocalInputValue(w.endsAt)
    } else {
      noEnd.value = true
    }
    monitors.value = (w.monitors ?? []).map((m) => ({
      monitorType: m.monitorType,
      monitorId: m.monitorId,
      name: m.name,
    }))
  },
  { immediate: true },
)
watch(loadError, (e) => {
  if (e) error.value = e.message
})

function validateMaintenanceWindowForm(): string {
  if (!title.value.trim()) return 'Title is required'
  if (!startsAt.value) return 'Start time is required'
  if (!noEnd.value && !endsAt.value) return 'End time is required, or check "no end date"'
  if (monitors.value.length === 0) return 'Select at least one monitor'
  return ''
}

function handleSubmitError(e: unknown): void {
  if (e instanceof ApiError && e.code === 'plan_limit_reached') {
    limitReached.value = true
    error.value = e.message
  } else {
    error.value = e instanceof Error ? e.message : 'Failed to update maintenance window'
  }
}

async function submit() {
  limitReached.value = false
  error.value = validateMaintenanceWindowForm()
  if (error.value) return

  submitting.value = true
  try {
    await maintenanceApi.update(id, {
      title: title.value.trim(),
      message: message.value.trim(),
      startsAt: toIso(startsAt.value),
      endsAt: noEnd.value ? null : toIso(endsAt.value),
      monitors: monitors.value.map((m) => ({ monitorType: m.monitorType, monitorId: m.monitorId })),
    })
    router.push({ name: 'maintenance' })
  } catch (e: unknown) {
    handleSubmitError(e)
  } finally {
    submitting.value = false
  }
}

async function deleteWindow() {
  error.value = ''
  try {
    await maintenanceApi.delete(id)
    router.push({ name: 'maintenance' })
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to delete maintenance window'
    confirmDelete.value = false
  }
}
</script>

<template>
  <MonitorFormPage
    title="Edit maintenance window"
    :back-to="{ name: 'maintenance' }"
    submit-label="Save changes"
    submitting-label="Saving…"
    :error="error"
    :submitting="submitting"
    :limit-reached="limitReached"
    :loading="loading"
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

    <template #actions>
      <div class="flex gap-2">
        <Button
          v-if="!confirmDelete"
          variant="secondary"
          type="button"
          style="color: var(--status-down)"
          @click="confirmDelete = true"
        >
          Delete
        </Button>
        <template v-else>
          <Button type="button" style="background-color: var(--status-down)" @click="deleteWindow">
            Confirm delete
          </Button>
          <Button variant="secondary" type="button" @click="confirmDelete = false">Cancel</Button>
        </template>
      </div>
    </template>
  </MonitorFormPage>
</template>
