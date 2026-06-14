<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import { monitorsApi } from '@/api/monitors'

const router = useRouter()

const name = ref('')
const schedule = ref('')
const gracePeriodMins = ref(5)
const submitting = ref(false)
const error = ref('')

const scheduleExamples = [
  { label: 'Every hour', value: 'every 1h' },
  { label: 'Every 30 min', value: 'every 30m' },
  { label: 'Daily at midnight', value: '0 0 * * *' },
  { label: 'Daily at 9am', value: '0 9 * * *' },
  { label: 'Every weekday', value: '0 9 * * 1-5' },
]

const graceOptions = [
  { label: '1 min', value: 1 },
  { label: '5 min', value: 5 },
  { label: '10 min', value: 10 },
  { label: '30 min', value: 30 },
  { label: '1 hour', value: 60 },
]

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
    const monitor = await monitorsApi.createCron({
      name: name.value.trim(),
      schedule: schedule.value.trim(),
      gracePeriodMins: gracePeriodMins.value,
    })
    router.push({ name: 'cron-monitor-detail', params: { id: monitor.id } })
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to create monitor'
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
          @click="router.push({ name: 'cron-monitors' })"
        >
          ← Back
        </button>
        <h1 class="text-2xl font-semibold" style="color: var(--text-strong)">New cron monitor</h1>
      </div>

      <form
        class="rounded-xl border p-6 space-y-5"
        style="background-color: var(--surface); border-color: var(--border)"
        @submit.prevent="submit"
      >
        <div>
          <Label for="name">Name</Label>
          <Input
            id="name"
            v-model="name"
            placeholder="Daily backup"
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
          <div class="flex flex-wrap gap-2 mt-2">
            <button
              v-for="ex in scheduleExamples"
              :key="ex.value"
              type="button"
              class="text-xs px-2 py-1 rounded transition-colors"
              style="background-color: var(--surface-raised); color: var(--text-dim)"
              @click="schedule = ex.value"
            >
              {{ ex.label }}
            </button>
          </div>
          <p class="text-xs mt-2" style="color: var(--text-muted)">
            Use a cron expression or <code>every Xm / every Xh / every Xd</code>.
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
          <p class="text-xs mt-1" style="color: var(--text-muted)">
            How long to wait after the expected time before alerting.
          </p>
        </div>

        <p v-if="error" class="text-sm" style="color: var(--status-down)">{{ error }}</p>

        <div class="flex gap-3 pt-1">
          <Button type="submit" :disabled="submitting">
            {{ submitting ? 'Creating…' : 'Create monitor' }}
          </Button>
          <Button variant="secondary" type="button" @click="router.push({ name: 'cron-monitors' })">
            Cancel
          </Button>
        </div>
      </form>
    </div>
  </AppLayout>
</template>
