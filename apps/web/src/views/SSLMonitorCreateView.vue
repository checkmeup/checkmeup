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
const hostname = ref('')
const submitting = ref(false)
const error = ref('')

async function submit() {
  error.value = ''
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
    })
    router.push({ name: 'ssl-monitor-detail', params: { id: monitor.id } })
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
          @click="router.push({ name: 'ssl-monitors' })"
        >
          ← Back
        </button>
        <h1 class="text-2xl font-semibold" style="color: var(--text-strong)">New SSL monitor</h1>
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

        <p v-if="error" class="text-sm" style="color: var(--status-down)">{{ error }}</p>

        <div class="flex gap-3 pt-1">
          <Button type="submit" :disabled="submitting">
            {{ submitting ? 'Creating…' : 'Create monitor' }}
          </Button>
          <Button variant="secondary" type="button" @click="router.push({ name: 'ssl-monitors' })">
            Cancel
          </Button>
        </div>
      </form>
    </div>
  </AppLayout>
</template>
