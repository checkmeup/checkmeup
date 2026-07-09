<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import Button from '@/components/ui/Button.vue'
import UpgradePrompt from '@/components/UpgradePrompt.vue'
import NotificationChannelForm from '@/components/NotificationChannelForm.vue'
import { ApiError } from '@/api/client'
import { notificationChannelsApi, type NotificationChannel } from '@/api/notificationChannels'
import { useNotificationChannels } from '@/composables/useNotificationChannels'
import { useBilling } from '@/composables/useBilling'
import {
  useNotificationChannelForm,
  typeLabel,
  typeIconPath,
  configKey,
} from '@/composables/useNotificationChannelForm'

const { data, isPending: loading, refetch } = useNotificationChannels()
const channels = computed(() => data.value ?? [])

const { data: billingInfo } = useBilling()

const form = reactive(useNotificationChannelForm({ billingInfo, refetch }))

const deletingId = ref('')

// relativeTime mirrors the small per-view helper used elsewhere (e.g.
// CronMonitorListView.vue) rather than a new shared util, matching this
// codebase's existing pattern for this exact bit of formatting.
function relativeTime(iso: string | undefined) {
  if (!iso) return ''
  const diff = Date.now() - new Date(iso).getTime()
  const m = Math.floor(diff / 60000)
  const h = Math.floor(m / 60)
  if (m < 1) return 'just now'
  if (m < 60) return `${m} min ago`
  if (h < 24) return `${h}h ago`
  return `${Math.floor(h / 24)}d ago`
}

function deliverySummary(c: NotificationChannel) {
  if (!['webhook', 'slack', 'sms'].includes(c.type) || !c.lastDeliveryStatus) return ''
  const parts = [c.lastDeliveryStatus, c.lastDeliveryDetail, relativeTime(c.lastDeliveryAt)].filter(
    Boolean,
  )
  return `Last delivery: ${parts.join(', ')}`
}

async function remove(c: NotificationChannel) {
  deletingId.value = c.id
  try {
    await notificationChannelsApi.delete(c.id)
    await refetch()
  } finally {
    deletingId.value = ''
  }
}

const toggleError = ref('')
const toggleLimitReached = ref(false)

async function toggleEnabled(c: NotificationChannel) {
  toggleError.value = ''
  toggleLimitReached.value = false
  try {
    await notificationChannelsApi.update(c.id, {
      type: c.type,
      name: c.name,
      config: c.config,
      enabled: !c.enabled,
    })
    await refetch()
  } catch (e: unknown) {
    if (e instanceof ApiError && e.code === 'plan_limit_reached') {
      toggleLimitReached.value = true
      toggleError.value = e.message
    } else {
      toggleError.value = e instanceof Error ? e.message : 'Failed to update channel'
    }
  }
}
</script>

<template>
  <div
    class="rounded-xl border p-6"
    style="background-color: var(--surface); border-color: var(--border)"
  >
    <h2 class="font-medium mb-1" style="color: var(--text-strong)">Notification channels</h2>
    <p class="text-sm mb-5" style="color: var(--text-muted)">
      Connect Telegram, email, Slack, SMS, and webhook destinations, then choose which channels
      each monitor alerts on. A monitor with no channels attached falls back to your account email.
    </p>

    <div v-if="loading" class="text-sm" style="color: var(--text-muted)">Loading…</div>

    <ul v-else-if="channels.length > 0" class="space-y-2 mb-5">
      <li
        v-for="c in channels"
        :key="c.id"
        class="flex items-center gap-3 rounded-lg border px-4 py-3"
        style="border-color: var(--border)"
      >
        <input
          type="checkbox"
          class="rounded flex-shrink-0"
          :checked="c.enabled"
          :title="c.enabled ? 'Disable channel' : 'Enable channel'"
          @change="toggleEnabled(c)"
        />
        <div
          class="w-[30px] h-[30px] rounded-lg flex items-center justify-center flex-shrink-0"
          style="background-color: var(--surface-raised); color: var(--text-dim)"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path :d="typeIconPath[c.type]" />
          </svg>
        </div>
        <div class="flex-1 min-w-0">
          <p class="text-sm truncate" style="color: var(--text)">{{ c.name }}</p>
          <p class="text-xs truncate font-mono" style="color: var(--text-muted)">
            {{ typeLabel[c.type] }} · {{ c.config[configKey[c.type]] }}
          </p>
          <p
            v-if="deliverySummary(c)"
            class="text-xs truncate"
            :style="{
              color: c.lastDeliveryStatus === 'success' ? 'var(--status-up)' : 'var(--status-down)',
            }"
          >
            {{ deliverySummary(c) }}
          </p>
        </div>
        <Button variant="secondary" size="sm" @click="form.startEdit(c)">Edit</Button>
        <Button variant="secondary" size="sm" :disabled="deletingId === c.id" @click="remove(c)">
          {{ deletingId === c.id ? 'Removing…' : 'Remove' }}
        </Button>
      </li>
    </ul>
    <p v-else class="text-sm mb-5" style="color: var(--text-muted)">No channels connected yet.</p>

    <UpgradePrompt v-if="toggleLimitReached" class="mb-5" :message="toggleError" />
    <p v-else-if="toggleError" class="text-sm mb-5" style="color: var(--status-down)">{{ toggleError }}</p>

    <Button v-if="!form.showForm" variant="secondary" @click="form.startAdd">+ Add channel</Button>

    <NotificationChannelForm v-else :form="form" />
  </div>
</template>
