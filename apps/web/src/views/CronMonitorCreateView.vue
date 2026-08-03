<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import MonitorFormPage from '@/components/MonitorFormPage.vue'
import CronMonitorForm from '@/components/CronMonitorForm.vue'
import { monitorsApi } from '@/api/monitors'
import { ApiError } from '@/api/client'
import {
  createCronMonitorFormState,
  cronMonitorFormPayload,
  scheduleExamples,
  validateCronMonitorForm,
} from '@/lib/cronMonitorForm'

const router = useRouter()

const form = ref(createCronMonitorFormState())
const submitting = ref(false)
const error = ref('')
const limitReached = ref(false)

async function submit() {
  limitReached.value = false
  error.value = validateCronMonitorForm(form.value)
  if (error.value) return

  submitting.value = true
  try {
    const monitor = await monitorsApi.createCron(cronMonitorFormPayload(form.value))
    router.push({ name: 'cron-monitor-detail', params: { id: monitor.id } })
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
    title="New cron monitor"
    :back-to="{ name: 'cron-monitors' }"
    submit-label="Create monitor"
    submitting-label="Creating…"
    :error="error"
    :submitting="submitting"
    :limit-reached="limitReached"
    @submit="submit"
  >
    <CronMonitorForm v-model="form" name-placeholder="Daily backup">
      <template #scheduleHelp>
        <div class="flex flex-wrap gap-2 mt-2">
          <button
            v-for="ex in scheduleExamples"
            :key="ex.value"
            type="button"
            class="text-xs px-2 py-1 rounded transition-colors"
            style="background-color: var(--surface-raised); color: var(--text-dim)"
            @click="form.schedule = ex.value"
          >
            {{ ex.label }}
          </button>
        </div>
        <p class="text-xs mt-2" style="color: var(--text-muted)">
          Use a cron expression or <code>every Xm / every Xh / every Xd</code>.
        </p>
      </template>

      <template #maxDurationHelp>
        <p class="text-xs mt-1" style="color: var(--text-muted)">
          Alert if a run doesn't finish within this time — only applies if your job also pings
          this monitor's <code>/start</code> URL when it begins (shown after creating this
          monitor).
        </p>
      </template>
    </CronMonitorForm>
  </MonitorFormPage>
</template>
