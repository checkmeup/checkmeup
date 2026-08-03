<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { monitorsApi } from '@/api/monitors'
import { ApiError } from '@/api/client'
import MonitorFormPage from '@/components/MonitorFormPage.vue'
import UptimeMonitorForm from '@/components/UptimeMonitorForm.vue'
import { useUptimeMonitor } from '@/composables/useUptimeMonitors'
import { useBilling } from '@/composables/useBilling'
import {
  createUptimeMonitorFormState,
  uptimeMonitorFormPayload,
  uptimeMonitorToFormState,
  validateUptimeMonitorForm,
} from '@/lib/uptimeMonitorForm'

const router = useRouter()
const route = useRoute()
const id = route.params.id as string

const form = ref(createUptimeMonitorFormState())
const submitting = ref(false)
const error = ref('')
const limitReached = ref(false)
const minIntervalMins = ref(5)

const { data: detail, isPending: monitorLoading, error: loadError } = useUptimeMonitor(id)
const { data: billingInfo, isPending: billingLoading } = useBilling()
const loading = computed(() => monitorLoading.value || billingLoading.value)

let formPopulated = false
watch(
  detail,
  (d) => {
    if (!d || formPopulated) return
    formPopulated = true
    form.value = uptimeMonitorToFormState(d.monitor)
  },
  { immediate: true },
)
watch(
  billingInfo,
  (info) => {
    if (!info) return
    minIntervalMins.value = info.minIntervalMins
  },
  { immediate: true },
)
watch(loadError, (e) => {
  if (e) error.value = e.message
})

async function submit() {
  limitReached.value = false
  error.value = validateUptimeMonitorForm(form.value)
  if (error.value) return

  submitting.value = true
  try {
    await monitorsApi.updateUptime(id, {
      ...uptimeMonitorFormPayload(form.value),
      alertsEnabled: form.value.alertsEnabled,
    })
    router.push({ name: 'uptime-monitor-detail', params: { id } })
  } catch (e: unknown) {
    if (e instanceof ApiError && e.code === 'plan_limit_reached') {
      limitReached.value = true
      error.value = e.message
    } else {
      error.value = e instanceof Error ? e.message : 'Failed to update monitor'
    }
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <MonitorFormPage
    title="Edit monitor"
    :back-to="{ name: 'uptime-monitor-detail', params: { id } }"
    submit-label="Save changes"
    submitting-label="Saving…"
    :error="error"
    :submitting="submitting"
    :limit-reached="limitReached"
    :loading="loading"
    @submit="submit"
  >
    <UptimeMonitorForm v-model="form" :min-interval-mins="minIntervalMins" show-alerts-toggle />
  </MonitorFormPage>
</template>
