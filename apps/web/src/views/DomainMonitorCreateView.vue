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
const domain = ref('')
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
  if (!domain.value.trim()) {
    error.value = 'Domain is required'
    return
  }

  submitting.value = true
  try {
    const monitor = await monitorsApi.createDomain({
      name: name.value.trim(),
      domain: domain.value.trim(),
      channelIds: channelIds.value,
    })
    router.push({ name: 'domain-monitor-detail', params: { id: monitor.id } })
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
    title="New domain monitor"
    :back-to="{ name: 'domain-monitors' }"
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
        placeholder="Production domain"
        class="mt-1"
        required
      />
    </div>

    <div>
      <Label for="domain">Domain</Label>
      <Input
        id="domain"
        v-model="domain"
        placeholder="example.com"
        class="mt-1"
      />
      <p class="text-xs mt-1" style="color: var(--text-muted)">
        Apex domain only — no https:// or path. Registration expiry checked daily.
      </p>
    </div>

    <NotificationChannelPicker v-model="channelIds" />
  </MonitorFormPage>
</template>
