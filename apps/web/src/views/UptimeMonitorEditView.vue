<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import { monitorsApi } from '@/api/monitors'
import { ApiError } from '@/api/client'
import UpgradePrompt from '@/components/UpgradePrompt.vue'
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
  <AppLayout>
    <div class="p-8 max-w-xl mx-auto">
      <div class="flex items-center gap-3 mb-6">
        <button
          class="text-sm transition-colors"
          style="color: var(--text-muted)"
          @click="router.push({ name: 'uptime-monitor-detail', params: { id } })"
        >
          ← Back
        </button>
        <h1 class="text-2xl font-semibold" style="color: var(--text-strong)">Edit monitor</h1>
      </div>

      <div v-if="loading" class="text-sm" style="color: var(--text-muted)">Loading…</div>

      <form
        v-else
        class="rounded-xl border p-6 space-y-5"
        style="background-color: var(--surface); border-color: var(--border)"
        @submit.prevent="submit"
      >
        <UptimeMonitorForm v-model="form" :min-interval-mins="minIntervalMins" show-alerts-toggle />

        <UpgradePrompt v-if="limitReached" :message="error" />
        <p v-else-if="error" class="text-sm" style="color: var(--status-down)">{{ error }}</p>

        <div class="flex gap-3 pt-1">
          <Button type="submit" :disabled="submitting">
            {{ submitting ? 'Saving…' : 'Save changes' }}
          </Button>
          <Button
            variant="secondary"
            type="button"
            @click="router.push({ name: 'uptime-monitor-detail', params: { id } })"
          >
            Cancel
          </Button>
        </div>
      </form>
    </div>
  </AppLayout>
</template>
