<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import MonitorFormPage from '@/components/MonitorFormPage.vue'
import MaintenanceWindowForm from '@/components/MaintenanceWindowForm.vue'
import Button from '@/components/ui/Button.vue'
import { maintenanceApi } from '@/api/maintenance'
import { ApiError } from '@/api/client'
import { useMaintenanceWindow } from '@/composables/useMaintenanceWindows'
import {
  createMaintenanceWindowFormState,
  maintenanceWindowFormPayload,
  maintenanceWindowToFormState,
  validateMaintenanceWindowForm,
} from '@/lib/maintenanceWindowForm'

const router = useRouter()
const route = useRoute()
const id = route.params.id as string

const form = ref(createMaintenanceWindowFormState())
const submitting = ref(false)
const error = ref('')
const confirmDelete = ref(false)
const limitReached = ref(false)

const { data: win, isPending: loading, error: loadError } = useMaintenanceWindow(id)
watch(
  win,
  (w) => {
    if (!w) return
    form.value = maintenanceWindowToFormState(w)
  },
  { immediate: true },
)
watch(loadError, (e) => {
  if (e) error.value = e.message
})

async function submit() {
  limitReached.value = false
  error.value = validateMaintenanceWindowForm(form.value)
  if (error.value) return

  submitting.value = true
  try {
    await maintenanceApi.update(id, maintenanceWindowFormPayload(form.value))
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
    <MaintenanceWindowForm v-model="form" />

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
