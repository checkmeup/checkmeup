<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import MonitorFormPage from '@/components/MonitorFormPage.vue'
import MaintenanceWindowForm from '@/components/MaintenanceWindowForm.vue'
import { maintenanceApi } from '@/api/maintenance'
import {
  createMaintenanceWindowFormState,
  maintenanceWindowFormPayload,
  validateMaintenanceWindowForm,
} from '@/lib/maintenanceWindowForm'

const router = useRouter()

const form = ref(createMaintenanceWindowFormState())
const submitting = ref(false)
const error = ref('')

async function submit() {
  error.value = validateMaintenanceWindowForm(form.value)
  if (error.value) return

  submitting.value = true
  try {
    await maintenanceApi.create(maintenanceWindowFormPayload(form.value))
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
    <MaintenanceWindowForm v-model="form" />
  </MonitorFormPage>
</template>
