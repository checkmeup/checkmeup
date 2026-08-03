<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import MonitorFormPage from '@/components/MonitorFormPage.vue'
import DNSMonitorForm from '@/components/DNSMonitorForm.vue'
import { monitorsApi, type DNSRecordType } from '@/api/monitors'
import { ApiError } from '@/api/client'
import { useDNSMonitor } from '@/composables/useDNSMonitors'
import { useBilling } from '@/composables/useBilling'
import {
  createDNSMonitorFormState,
  dnsMonitorFormPayload,
  dnsMonitorToFormState,
  validateDNSMonitorForm,
} from '@/lib/dnsMonitorForm'

const router = useRouter()
const route = useRoute()
const id = route.params.id as string

const form = ref(createDNSMonitorFormState())
const submitting = ref(false)
const error = ref('')
const limitReached = ref(false)
const minIntervalMins = ref(5)

const { data: detail, isPending: monitorLoading, error: loadError } = useDNSMonitor(id)
const { data: billingInfo, isPending: billingLoading } = useBilling()
const loading = computed(() => monitorLoading.value || billingLoading.value)

// recordTypeAtLoad tracks the value fetched from the server so changing the
// record type away from it clears the now-meaningless expected value — an A
// record's address means nothing for a TXT lookup. Only clears on a real
// change by the user, not on the initial populate.
let formPopulated = false
let recordTypeAtLoad: DNSRecordType | null = null
watch(
  detail,
  (d) => {
    if (!d || formPopulated) return
    formPopulated = true
    recordTypeAtLoad = d.monitor.recordType
    form.value = dnsMonitorToFormState(d.monitor)
  },
  { immediate: true },
)
watch(
  () => form.value.recordType,
  (newType) => {
    if (formPopulated && recordTypeAtLoad !== null && newType !== recordTypeAtLoad) {
      form.value.expectedValue = ''
    }
  },
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
  error.value = validateDNSMonitorForm(form.value)
  if (error.value) return

  submitting.value = true
  try {
    await monitorsApi.updateDns(id, {
      ...dnsMonitorFormPayload(form.value),
      // Sent even when blank — '' actively re-arms baseline mode, unlike
      // create, where a blank field is omitted.
      expectedValue: form.value.expectedValue.trim(),
      alertsEnabled: form.value.alertsEnabled,
    })
    router.push({ name: 'dns-monitor-detail', params: { id } })
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
    :back-to="{ name: 'dns-monitor-detail', params: { id } }"
    submit-label="Save changes"
    submitting-label="Saving…"
    :error="error"
    :submitting="submitting"
    :limit-reached="limitReached"
    :loading="loading"
    @submit="submit"
  >
    <DNSMonitorForm v-model="form" :min-interval-mins="minIntervalMins" show-alerts-toggle>
      <template #expectedValueHelp>
        <p class="text-xs mt-1" style="color: var(--text-muted)">
          Clear this to re-arm baseline mode — the next check re-captures whatever the record currently resolves to. Save a new value to acknowledge an intentional change.
        </p>
      </template>
    </DNSMonitorForm>
  </MonitorFormPage>
</template>
