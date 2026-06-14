<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import { monitorsApi } from '@/api/monitors'

const router = useRouter()
const route = useRoute()
const id = route.params.id as string

const name = ref('')
const schedule = ref('')
const gracePeriodMins = ref(5)
const alertsEnabled = ref(true)
const loading = ref(true)
const submitting = ref(false)
const error = ref('')

const graceOptions = [
  { label: '1 min', value: 1 },
  { label: '5 min', value: 5 },
  { label: '10 min', value: 10 },
  { label: '30 min', value: 30 },
  { label: '1 hour', value: 60 },
]

onMounted(async () => {
  try {
    const detail = await monitorsApi.getCron(id)
    const m = detail.monitor
    name.value = m.name
    schedule.value = m.schedule
    gracePeriodMins.value = m.gracePeriodMins
    alertsEnabled.value = m.alertsEnabled
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to load monitor'
  } finally {
    loading.value = false
  }
})

async function submit() {
  error.value = ''
  if (!name.value.trim()) {
    error.value = 'Name is required'
    return
  }
  if (!schedule.value.trim()) {
    error.value = 'Schedule is required'
    return
  }

  submitting.value = true
  try {
    await monitorsApi.updateCron(id, {
      name: name.value.trim(),
      schedule: schedule.value.trim(),
      gracePeriodMins: gracePeriodMins.value,
      alertsEnabled: alertsEnabled.value,
    })
    router.push({ name: 'cron-monitor-detail', params: { id } })
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to update monitor'
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
          @click="router.push({ name: 'cron-monitor-detail', params: { id } })"
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
        <div>
          <Label for="name">Name</Label>
          <Input
            id="name"
            v-model="name"
            class="mt-1"
            required
          />
        </div>

        <div>
          <Label for="schedule">Schedule</Label>
          <Input
            id="schedule"
            v-model="schedule"
            placeholder="every 1h  or  0 * * * *"
            class="mt-1"
          />
          <p class="text-xs mt-1" style="color: var(--text-muted)">
            The ping URL never changes when you edit the schedule.
          </p>
        </div>

        <div>
          <Label for="grace">Grace period</Label>
          <select
            id="grace"
            v-model="gracePeriodMins"
            class="mt-1 w-full rounded-md border px-3 py-2 text-sm"
            style="background-color: var(--surface-raised); border-color: var(--border); color: var(--text)"
          >
            <option v-for="opt in graceOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </option>
          </select>
        </div>

        <div class="flex items-center gap-3">
          <input
            id="alerts"
            v-model="alertsEnabled"
            type="checkbox"
            class="rounded"
          />
          <Label for="alerts" class="cursor-pointer">Send Telegram alerts</Label>
        </div>

        <p v-if="error" class="text-sm" style="color: var(--status-down)">{{ error }}</p>

        <div class="flex gap-3 pt-1">
          <Button type="submit" :disabled="submitting">
            {{ submitting ? 'Saving…' : 'Save changes' }}
          </Button>
          <Button
            variant="secondary"
            type="button"
            @click="router.push({ name: 'cron-monitor-detail', params: { id } })"
          >
            Cancel
          </Button>
        </div>
      </form>
    </div>
  </AppLayout>
</template>
