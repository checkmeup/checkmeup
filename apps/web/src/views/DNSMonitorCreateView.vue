<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import MonitorFormPage from '@/components/MonitorFormPage.vue'
import DNSMonitorForm from '@/components/DNSMonitorForm.vue'
import { monitorsApi } from '@/api/monitors'
import { ApiError } from '@/api/client'
import { useBilling } from '@/composables/useBilling'
import {
  createDNSMonitorFormState,
  dnsMonitorFormPayload,
  validateDNSMonitorForm,
} from '@/lib/dnsMonitorForm'

const router = useRouter()

const form = ref(createDNSMonitorFormState())
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
  error.value = validateDNSMonitorForm(form.value)
  if (error.value) return

  submitting.value = true
  try {
    const monitor = await monitorsApi.createDns({
      ...dnsMonitorFormPayload(form.value),
      // Blank is omitted, not sent as '': the worker then captures whatever
      // the record resolves to on first check as the baseline.
      expectedValue: form.value.expectedValue.trim() || undefined,
    })
    router.push({ name: 'dns-monitor-detail', params: { id: monitor.id } })
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
    title="New DNS monitor"
    :back-to="{ name: 'dns-monitors' }"
    submit-label="Create monitor"
    submitting-label="Creating…"
    :error="error"
    :submitting="submitting"
    :limit-reached="limitReached"
    @submit="submit"
  >
    <DNSMonitorForm v-model="form" :min-interval-mins="minIntervalMins">
      <template #expectedValueHelp>
        <p class="text-xs mt-1" style="color: var(--text-muted)">
          Leave blank to auto-detect the current value on first check and alert on any later change. Set a value to pin it and alert on any mismatch from creation onward.
        </p>
      </template>
    </DNSMonitorForm>
  </MonitorFormPage>
</template>
