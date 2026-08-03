<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import MonitorFormPage from '@/components/MonitorFormPage.vue'
import CronMonitorForm from '@/components/CronMonitorForm.vue'
import { monitorsApi } from '@/api/monitors'
import { ApiError } from '@/api/client'
import { useCronMonitor } from '@/composables/useCronMonitors'
import {
  createCronMonitorFormState,
  cronMonitorFormPayload,
  cronMonitorToFormState,
  validateCronMonitorForm,
} from '@/lib/cronMonitorForm'

const router = useRouter()
const route = useRoute()
const id = route.params.id as string

const form = ref(createCronMonitorFormState())
const submitting = ref(false)
const error = ref('')
const limitReached = ref(false)

const { data: detail, isPending: loading, error: loadError } = useCronMonitor(id)
watch(
  detail,
  (d) => {
    if (!d) return
    form.value = cronMonitorToFormState(d.monitor)
  },
  { immediate: true },
)
watch(loadError, (e) => {
  if (e) error.value = e.message
})

async function submit() {
  limitReached.value = false
  error.value = validateCronMonitorForm(form.value)
  if (error.value) return

  submitting.value = true
  try {
    await monitorsApi.updateCron(id, {
      ...cronMonitorFormPayload(form.value),
      alertsEnabled: form.value.alertsEnabled,
    })
    router.push({ name: 'cron-monitor-detail', params: { id } })
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
    :back-to="{ name: 'cron-monitor-detail', params: { id } }"
    submit-label="Save changes"
    submitting-label="Saving…"
    :error="error"
    :submitting="submitting"
    :limit-reached="limitReached"
    :loading="loading"
    @submit="submit"
  >
    <CronMonitorForm v-model="form" show-alerts-toggle>
      <template #scheduleHelp>
        <p class="text-xs mt-1" style="color: var(--text-muted)">
          The ping URL never changes when you edit the schedule.
        </p>
      </template>

      <template #maxDurationHelp>
        <p class="text-xs mt-1" style="color: var(--text-muted)">
          Alert if a run doesn't finish within this time — only applies if your job also pings
          <code>{{ detail?.monitor.pingUrl }}/start</code> when it begins.
        </p>
      </template>
    </CronMonitorForm>
  </MonitorFormPage>
</template>
