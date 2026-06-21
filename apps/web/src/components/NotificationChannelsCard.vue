<script setup lang="ts">
import { computed, ref } from 'vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import {
  notificationChannelsApi,
  type NotificationChannel,
  type NotificationChannelType,
} from '@/api/notificationChannels'
import { useNotificationChannels } from '@/composables/useNotificationChannels'

const { data, isPending: loading, refetch } = useNotificationChannels()
const channels = computed(() => data.value ?? [])

const typeLabel: Record<NotificationChannelType, string> = { telegram: 'Telegram', email: 'Email' }
const configKey: Record<NotificationChannelType, string> = { telegram: 'chatId', email: 'email' }

const showForm = ref(false)
const editingId = ref<string | null>(null)
const type = ref<NotificationChannelType>('telegram')
const name = ref('')
const value = ref('')
const enabled = ref(true)

const saving = ref(false)
const testing = ref(false)
const deletingId = ref('')
const formError = ref('')
const testSuccess = ref(false)
const testError = ref('')

function startAdd() {
  editingId.value = null
  type.value = 'telegram'
  name.value = ''
  value.value = ''
  enabled.value = true
  formError.value = ''
  testSuccess.value = false
  testError.value = ''
  showForm.value = true
}

function startEdit(c: NotificationChannel) {
  editingId.value = c.id
  type.value = c.type
  name.value = c.name
  value.value = c.config[configKey[c.type]] ?? ''
  enabled.value = c.enabled
  formError.value = ''
  testSuccess.value = false
  testError.value = ''
  showForm.value = true
}

function cancelForm() {
  showForm.value = false
  editingId.value = null
}

function buildConfig(): Record<string, string> {
  return { [configKey[type.value]]: value.value.trim() }
}

async function test() {
  testing.value = true
  testError.value = ''
  testSuccess.value = false
  try {
    await notificationChannelsApi.test({ type: type.value, config: buildConfig() })
    testSuccess.value = true
  } catch (e: unknown) {
    testError.value = e instanceof Error ? e.message : 'Failed to send test message'
  } finally {
    testing.value = false
  }
}

async function save() {
  formError.value = ''
  if (!name.value.trim()) {
    formError.value = 'Name is required'
    return
  }
  if (!value.value.trim()) {
    formError.value = type.value === 'telegram' ? 'Chat ID is required' : 'Email is required'
    return
  }

  saving.value = true
  try {
    if (editingId.value) {
      await notificationChannelsApi.update(editingId.value, {
        type: type.value,
        name: name.value.trim(),
        config: buildConfig(),
        enabled: enabled.value,
      })
    } else {
      await notificationChannelsApi.create({
        type: type.value,
        name: name.value.trim(),
        config: buildConfig(),
      })
    }
    await refetch()
    cancelForm()
  } catch (e: unknown) {
    formError.value = e instanceof Error ? e.message : 'Failed to save channel'
  } finally {
    saving.value = false
  }
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

async function toggleEnabled(c: NotificationChannel) {
  await notificationChannelsApi.update(c.id, {
    type: c.type,
    name: c.name,
    config: c.config,
    enabled: !c.enabled,
  })
  await refetch()
}
</script>

<template>
  <div class="rounded-xl border p-6" style="background-color: var(--surface); border-color: var(--border)">
    <h2 class="font-medium mb-1" style="color: var(--text-strong)">Notification channels</h2>
    <p class="text-sm mb-5" style="color: var(--text-muted)">
      Connect Telegram and email destinations, then choose which channels each monitor alerts on. A
      monitor with no channels attached falls back to your account email.
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
        <div class="flex-1 min-w-0">
          <p class="text-sm truncate" style="color: var(--text)">{{ c.name }}</p>
          <p class="text-xs truncate" style="color: var(--text-muted)">
            {{ typeLabel[c.type] }} · {{ c.config[configKey[c.type]] }}
          </p>
        </div>
        <Button variant="secondary" size="sm" @click="startEdit(c)">Edit</Button>
        <Button variant="secondary" size="sm" :disabled="deletingId === c.id" @click="remove(c)">
          {{ deletingId === c.id ? 'Removing…' : 'Remove' }}
        </Button>
      </li>
    </ul>
    <p v-else class="text-sm mb-5" style="color: var(--text-muted)">No channels connected yet.</p>

    <Button v-if="!showForm" variant="secondary" @click="startAdd">+ Add channel</Button>

    <div
      v-else
      class="rounded-lg border p-4 space-y-4"
      style="border-color: var(--border); background-color: var(--surface-raised)"
    >
      <div>
        <Label for="channel-type">Type</Label>
        <select
          id="channel-type"
          v-model="type"
          class="mt-1 w-full rounded-md border px-3 py-2 text-sm"
          style="background-color: var(--surface); border-color: var(--border); color: var(--text)"
          :disabled="!!editingId"
        >
          <option value="telegram">Telegram</option>
          <option value="email">Email</option>
        </select>
      </div>

      <ol
        v-if="type === 'telegram'"
        class="text-sm space-y-2 list-decimal list-inside"
        style="color: var(--text-dim)"
      >
        <li>
          Open
          <a
            href="https://t.me/checkmeupnet_bot"
            target="_blank"
            rel="noopener"
            class="underline"
            style="color: var(--color-green-500)"
            >@checkmeupnet_bot</a
          >
          in Telegram and send
          <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)">/start</code>
          — the bot will reply with your Chat ID
        </li>
        <li>Paste the Chat ID below and click <strong>Send test message</strong> to verify</li>
      </ol>

      <div>
        <Label for="channel-name">Name</Label>
        <Input id="channel-name" v-model="name" placeholder="e.g. Ops Telegram" class="mt-1" />
      </div>

      <div>
        <Label for="channel-value">{{ type === 'telegram' ? 'Chat ID' : 'Email address' }}</Label>
        <Input
          id="channel-value"
          v-model="value"
          :type="type === 'email' ? 'email' : 'text'"
          :placeholder="type === 'telegram' ? '-1001234567890' : 'alerts@yourteam.com'"
          class="mt-1"
        />
      </div>

      <div class="flex items-center gap-3">
        <Button variant="secondary" :disabled="!value.trim() || testing" @click="test">
          {{ testing ? 'Sending…' : 'Send test message' }}
        </Button>
        <Button :disabled="saving" @click="save">
          {{ saving ? 'Saving…' : editingId ? 'Save changes' : 'Add channel' }}
        </Button>
        <Button variant="secondary" type="button" @click="cancelForm">Cancel</Button>
      </div>

      <p v-if="testSuccess" class="text-sm" style="color: var(--status-up)">Test message sent!</p>
      <p v-if="testError" class="text-sm" style="color: var(--status-down)">{{ testError }}</p>
      <p v-if="formError" class="text-sm" style="color: var(--status-down)">{{ formError }}</p>
    </div>
  </div>
</template>
