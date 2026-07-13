<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import MaintenanceMonitorPicker from '@/components/MaintenanceMonitorPicker.vue'
import { maintenanceApi, type MaintenanceMonitorRef } from '@/api/maintenance'
import { ApiError } from '@/api/client'
import UpgradePrompt from '@/components/UpgradePrompt.vue'
import { useMaintenanceWindow } from '@/composables/useMaintenanceWindows'
import { validateMaintenanceWindowForm } from '@/lib/maintenanceWindowValidation'

const router = useRouter()
const route = useRoute()
const id = route.params.id as string

const title = ref('')
const message = ref('')
const startsAt = ref('')
const noEnd = ref(false)
const endsAt = ref('')
const monitors = ref<MaintenanceMonitorRef[]>([])
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

async function submit() {
  error.value = ''
  limitReached.value = false
  const validationError = validateMaintenanceWindowForm(
    title.value,
    startsAt.value,
    noEnd.value,
    endsAt.value,
    monitors.value.length,
  )
  if (validationError) {
    error.value = validationError
    return
  }

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
    if (e instanceof ApiError && e.code === 'plan_limit_reached') {
      limitReached.value = true
      error.value = e.message
    } else {
      error.value = e instanceof Error ? e.message : 'Failed to update maintenance window'
    }
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
  <AppLayout>
    <div class="p-8 max-w-xl mx-auto">
      <div class="flex items-center gap-3 mb-6">
        <button
          class="text-sm"
          style="color: var(--text-muted)"
          @click="router.push({ name: 'maintenance' })"
        >
          ← Back
        </button>
        <h1 class="text-2xl font-semibold" style="color: var(--text-strong)">Edit maintenance window</h1>
      </div>

      <div v-if="loading" class="text-sm" style="color: var(--text-muted)">Loading…</div>

      <form
        v-else
        class="rounded-xl border p-6 space-y-5"
        style="background-color: var(--surface); border-color: var(--border)"
        @submit.prevent="submit"
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

        <UpgradePrompt v-if="limitReached" :message="error" />
        <p v-else-if="error" class="text-sm" style="color: var(--status-down)">{{ error }}</p>

        <div class="flex items-center justify-between pt-1">
          <div class="flex gap-3">
            <Button type="submit" :disabled="submitting">
              {{ submitting ? 'Saving…' : 'Save changes' }}
            </Button>
            <Button variant="secondary" type="button" @click="router.push({ name: 'maintenance' })">
              Cancel
            </Button>
          </div>
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
        </div>
      </form>
    </div>
  </AppLayout>
</template>
