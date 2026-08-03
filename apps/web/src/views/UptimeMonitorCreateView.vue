<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import { monitorsApi } from '@/api/monitors'
import { ApiError } from '@/api/client'
import UpgradePrompt from '@/components/UpgradePrompt.vue'
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
  <AppLayout>
    <div class="p-8 max-w-xl mx-auto">
      <div class="flex items-center gap-3 mb-6">
        <button
          class="text-sm transition-colors"
          style="color: var(--text-muted)"
          @click="router.push({ name: 'uptime-monitors' })"
        >
          ← Back
        </button>
        <h1 class="text-2xl font-semibold" style="color: var(--text-strong)">New uptime monitor</h1>
      </div>

      <form
        class="rounded-xl border p-6 space-y-5"
        style="background-color: var(--surface); border-color: var(--border)"
        @submit.prevent="submit"
      >
        <UptimeMonitorForm
          v-model="form"
          :min-interval-mins="minIntervalMins"
          name-placeholder="Production API"
        />

        <UpgradePrompt v-if="limitReached" :message="error" />
        <p v-else-if="error" class="text-sm" style="color: var(--status-down)">{{ error }}</p>

        <div class="flex gap-3 pt-1">
          <Button type="submit" :disabled="submitting">
            {{ submitting ? 'Creating…' : 'Create monitor' }}
          </Button>
          <Button variant="secondary" type="button" @click="router.push({ name: 'uptime-monitors' })">
            Cancel
          </Button>
        </div>
      </form>
    </div>
  </AppLayout>
</template>
