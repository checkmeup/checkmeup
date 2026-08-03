<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import MonitorFormPage from '@/components/MonitorFormPage.vue'
import PortMonitorForm from '@/components/PortMonitorForm.vue'
import { monitorsApi } from '@/api/monitors'
import { ApiError } from '@/api/client'
import { useBilling } from '@/composables/useBilling'
import {
  createPortMonitorFormState,
  portMonitorFormPayload,
  validatePortMonitorForm,
} from '@/lib/portMonitorForm'

const router = useRouter()

const form = ref(createPortMonitorFormState())
const submitting = ref(false)
const error = ref('')
const limitReached = ref(false)
const minIntervalMins = ref(5)

const { data: billingInfo } = useBilling()
watch(
  billingInfo,
  (info) => {
    if (!info) return
    minIntervalMins.value = info.minIntervalMins
  },
  { immediate: true },
)

async function submit() {
  limitReached.value = false
  error.value = validatePortMonitorForm(form.value)
  if (error.value) return

  submitting.value = true
  try {
    const monitor = await monitorsApi.createPort(portMonitorFormPayload(form.value))
    router.push({ name: 'port-monitor-detail', params: { id: monitor.id } })
  } catch (e: unknown) {
    if (e instanceof ApiError && e.code === 'plan_limit_reached') {
      limitReached.value = true
      error.value = e.message
    } else {
      error.value = e instanceof Error ? e.message : 'Failed to create monitor'
    }
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <MonitorFormPage
    title="New port monitor"
    :back-to="{ name: 'port-monitors' }"
    submit-label="Create monitor"
    submitting-label="Creating…"
    :error="error"
    :submitting="submitting"
    :limit-reached="limitReached"
    @submit="submit"
  >
    <PortMonitorForm v-model="form" :min-interval-mins="minIntervalMins" />
  </MonitorFormPage>
</template>
