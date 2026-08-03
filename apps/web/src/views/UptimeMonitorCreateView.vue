<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { monitorsApi } from '@/api/monitors'
import { ApiError } from '@/api/client'
import MonitorFormPage from '@/components/MonitorFormPage.vue'
import UptimeMonitorForm from '@/components/UptimeMonitorForm.vue'
import { useBilling } from '@/composables/useBilling'
import {
  createUptimeMonitorFormState,
  uptimeMonitorFormPayload,
  validateUptimeMonitorForm,
} from '@/lib/uptimeMonitorForm'

const router = useRouter()

const form = ref(createUptimeMonitorFormState())
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
// Load failures stay silent (keep defaults), same as the original try/catch.

async function submit() {
  limitReached.value = false
  error.value = validateUptimeMonitorForm(form.value)
  if (error.value) return

  submitting.value = true
  try {
    const monitor = await monitorsApi.createUptime(uptimeMonitorFormPayload(form.value))
    router.push({ name: 'uptime-monitor-detail', params: { id: monitor.id } })
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
    title="New uptime monitor"
    :back-to="{ name: 'uptime-monitors' }"
    submit-label="Create monitor"
    submitting-label="Creating…"
    :error="error"
    :submitting="submitting"
    :limit-reached="limitReached"
    @submit="submit"
  >
    <UptimeMonitorForm
      v-model="form"
      :min-interval-mins="minIntervalMins"
      name-placeholder="Production API"
    />
  </MonitorFormPage>
</template>
