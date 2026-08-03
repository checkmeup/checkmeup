<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import MonitorFormPage from '@/components/MonitorFormPage.vue'
import PortMonitorForm from '@/components/PortMonitorForm.vue'
import { monitorsApi } from '@/api/monitors'
import { ApiError } from '@/api/client'
import { usePortMonitor } from '@/composables/usePortMonitors'
import { useBilling } from '@/composables/useBilling'
import {
  createPortMonitorFormState,
  portMonitorFormPayload,
  portMonitorToFormState,
  validatePortMonitorForm,
} from '@/lib/portMonitorForm'

const router = useRouter()
const route = useRoute()
const id = route.params.id as string

const form = ref(createPortMonitorFormState())
const submitting = ref(false)
const error = ref('')
const limitReached = ref(false)
const minIntervalMins = ref(5)

const { data: detail, isPending: monitorLoading, error: loadError } = usePortMonitor(id)
const { data: billingInfo, isPending: billingLoading } = useBilling()
const loading = computed(() => monitorLoading.value || billingLoading.value)

let formPopulated = false
watch(
  detail,
  (d) => {
    if (!d || formPopulated) return
    formPopulated = true
    form.value = portMonitorToFormState(d.monitor)
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
  error.value = validatePortMonitorForm(form.value)
  if (error.value) return

  submitting.value = true
  try {
    await monitorsApi.updatePort(id, {
      ...portMonitorFormPayload(form.value),
      alertsEnabled: form.value.alertsEnabled,
    })
    router.push({ name: 'port-monitor-detail', params: { id } })
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
    :back-to="{ name: 'port-monitor-detail', params: { id } }"
    submit-label="Save changes"
    submitting-label="Saving…"
    :error="error"
    :submitting="submitting"
    :limit-reached="limitReached"
    :loading="loading"
    @submit="submit"
  >
    <PortMonitorForm v-model="form" :min-interval-mins="minIntervalMins" show-alerts-toggle />
  </MonitorFormPage>
</template>
