<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import MonitorFormPage from '@/components/MonitorFormPage.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import { monitorsApi } from '@/api/monitors'
import { ApiError } from '@/api/client'
import NotificationChannelPicker from '@/components/NotificationChannelPicker.vue'

const router = useRouter()

const name = ref('')
const hostname = ref('')
const channelIds = ref<string[]>([])
const submitting = ref(false)
const error = ref('')
const limitReached = ref(false)

async function submit() {
  error.value = ''
  limitReached.value = false
  if (!name.value.trim()) {
    error.value = 'Name is required'
    return
  }
  if (!hostname.value.trim()) {
    error.value = 'Hostname is required'
    return
  }

  submitting.value = true
  try {
    const monitor = await monitorsApi.createSSL({
      name: name.value.trim(),
      hostname: hostname.value.trim(),
      channelIds: channelIds.value,
    })
    router.push({ name: 'ssl-monitor-detail', params: { id: monitor.id } })
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
    title="New SSL monitor"
    :back-to="{ name: 'ssl-monitors' }"
    submit-label="Create monitor"
    submitting-label="Creating…"
    :error="error"
    :submitting="submitting"
    :limit-reached="limitReached"
    @submit="submit"
  >
    <div>
      <Label for="name">Name</Label>
      <Input
        id="name"
        v-model="name"
        placeholder="Production API"
        class="mt-1"
        required
      />
    </div>

    <div>
      <Label for="hostname">Hostname</Label>
      <Input
        id="hostname"
        v-model="hostname"
        placeholder="example.com"
        class="mt-1"
      />
      <p class="text-xs mt-1" style="color: var(--text-muted)">
        Domain name only — no https:// or path. Checked daily on port 443.
      </p>
    </div>

    <NotificationChannelPicker v-model="channelIds" />
  </MonitorFormPage>
</template>
