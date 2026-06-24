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

const typeLabel: Record<NotificationChannelType, string> = {
  telegram: 'Telegram',
  email: 'Email',
  webhook: 'Webhook',
  slack: 'Slack',
}
const configKey: Record<NotificationChannelType, string> = {
  telegram: 'chatId',
  email: 'email',
  webhook: 'url',
  slack: 'url',
}
const valueLabel: Record<NotificationChannelType, string> = {
  telegram: 'Chat ID',
  email: 'Email address',
  webhook: 'Webhook URL',
  slack: 'Incoming Webhook URL',
}
const valuePlaceholder: Record<NotificationChannelType, string> = {
  telegram: '-1001234567890',
  email: 'alerts@yourteam.com',
  webhook: 'https://example.com/hooks/checkmeup',
  slack: 'https://hooks.slack.com/services/...',
}

const showForm = ref(false)
const editingId = ref<string | null>(null)
const type = ref<NotificationChannelType>('telegram')
const name = ref('')
const value = ref('')
const enabled = ref(true)
// Only populated when editing an existing webhook channel — the secret
// doesn't exist until the channel is first saved (US-1401).
const secret = ref('')

const saving = ref(false)
const testing = ref(false)
const deletingId = ref('')
const regeneratingSecret = ref(false)
const regenerateError = ref('')
const formError = ref('')
const testSuccess = ref(false)
const testError = ref('')

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
  if ((c.type !== 'webhook' && c.type !== 'slack') || !c.lastDeliveryStatus) return ''
  const parts = [c.lastDeliveryStatus, c.lastDeliveryDetail, relativeTime(c.lastDeliveryAt)].filter(
    Boolean,
  )
  return `Last delivery: ${parts.join(', ')}`
}

function startAdd() {
  editingId.value = null
  type.value = 'telegram'
  name.value = ''
  value.value = ''
  secret.value = ''
  enabled.value = true
  formError.value = ''
  testSuccess.value = false
  testError.value = ''
  regenerateError.value = ''
  showForm.value = true
}

function startEdit(c: NotificationChannel) {
  editingId.value = c.id
  type.value = c.type
  name.value = c.name
  value.value = c.config[configKey[c.type]] ?? ''
  secret.value = c.type === 'webhook' ? (c.config.secret ?? '') : ''
  enabled.value = c.enabled
  formError.value = ''
  testSuccess.value = false
  testError.value = ''
  regenerateError.value = ''
  showForm.value = true
}

async function regenerateSecret() {
  if (!editingId.value) return
  regeneratingSecret.value = true
  regenerateError.value = ''
  try {
    const updated = await notificationChannelsApi.regenerateWebhookSecret(editingId.value)
    secret.value = updated.config.secret ?? ''
    await refetch()
  } catch (e: unknown) {
    regenerateError.value = e instanceof Error ? e.message : 'Failed to regenerate secret'
  } finally {
    regeneratingSecret.value = false
  }
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
    formError.value = `${valueLabel[type.value]} is required`
    return
  }
  if (type.value === 'webhook' && !value.value.trim().startsWith('https://')) {
    formError.value = 'Webhook URL must start with https://'
    return
  }
  if (type.value === 'slack' && !value.value.trim().startsWith('https://hooks.slack.com/')) {
    formError.value = 'Must be a Slack Incoming Webhook URL (https://hooks.slack.com/...)'
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
  <div
    class="rounded-xl border p-6"
    style="background-color: var(--surface); border-color: var(--border)"
  >
    <h2 class="font-medium mb-1" style="color: var(--text-strong)">Notification channels</h2>
    <p class="text-sm mb-5" style="color: var(--text-muted)">
      Connect Telegram, email, Slack, and webhook destinations, then choose which channels each
      monitor alerts on. A monitor with no channels attached falls back to your account email.
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
          <option value="webhook">Webhook</option>
          <option value="slack">Slack</option>
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
          <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
            >/start</code
          >
          — the bot will reply with your Chat ID
        </li>
        <li>Paste the Chat ID below and click <strong>Send test message</strong> to verify</li>
      </ol>

      <p v-else-if="type === 'webhook'" class="text-sm" style="color: var(--text-dim)">
        checkmeup will POST a JSON payload to this URL on every down/recovery event, signed with
        <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
          >X-Checkmeup-Signature</code
        >. Must be
        <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
          >https://</code
        >.
      </p>

      <ol
        v-else-if="type === 'slack'"
        class="text-sm space-y-2 list-decimal list-inside"
        style="color: var(--text-dim)"
      >
        <li>
          In Slack, go to <strong>Apps → Incoming Webhooks</strong> and create a new webhook for
          your target channel
        </li>
        <li>
          Copy the <strong>Webhook URL</strong> (starts with
          <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
            >https://hooks.slack.com/services/</code
          >) and paste it below
        </li>
        <li>Click <strong>Send test message</strong> to verify the connection before saving</li>
      </ol>

      <div>
        <Label for="channel-name">Name</Label>
        <Input id="channel-name" v-model="name" placeholder="e.g. Ops Telegram" class="mt-1" />
      </div>

      <div>
        <Label for="channel-value">{{ valueLabel[type] }}</Label>
        <Input
          id="channel-value"
          v-model="value"
          :type="type === 'email' ? 'email' : type === 'webhook' ? 'url' : 'text'"
          :placeholder="valuePlaceholder[type]"
          class="mt-1"
        />
      </div>

      <div v-if="type === 'webhook' && editingId" class="space-y-2">
        <Label for="channel-secret">Signing secret</Label>
        <div class="flex items-center gap-3">
          <Input
            id="channel-secret"
            :model-value="secret"
            disabled
            class="mt-1 font-mono text-xs"
          />
          <Button
            variant="secondary"
            size="sm"
            :disabled="regeneratingSecret"
            @click="regenerateSecret"
          >
            {{ regeneratingSecret ? 'Regenerating…' : 'Regenerate' }}
          </Button>
        </div>
        <p class="text-xs" style="color: var(--text-muted)">
          Verify a request by computing HMAC-SHA256 of the raw request body using this secret as the
          key, hex-encoding it, and comparing it to the
          <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
            >X-Checkmeup-Signature</code
          >
          header. Regenerating invalidates the signature for future sends only — already-delivered
          requests aren't affected.
        </p>
        <p v-if="regenerateError" class="text-xs" style="color: var(--status-down)">
          {{ regenerateError }}
        </p>
      </div>
      <p v-else-if="type === 'webhook'" class="text-xs" style="color: var(--text-muted)">
        A signing secret is generated automatically once you save this channel.
      </p>

      <div class="flex items-center gap-3">
        <Button variant="secondary" :disabled="!value.trim() || testing" @click="test">
          {{
            testing ? 'Sending…' : type === 'webhook' ? 'Send test webhook' : 'Send test message'
          }}
        </Button>
        <Button :disabled="saving" @click="save">
          {{ saving ? 'Saving…' : editingId ? 'Save changes' : 'Add channel' }}
        </Button>
        <Button variant="secondary" type="button" @click="cancelForm">Cancel</Button>
      </div>

      <p v-if="testSuccess" class="text-sm" style="color: var(--status-up)">
        {{ type === 'webhook' ? 'Test webhook sent!' : 'Test message sent!' }}
      </p>
      <p v-if="testError" class="text-sm" style="color: var(--status-down)">{{ testError }}</p>
      <p v-if="formError" class="text-sm" style="color: var(--status-down)">{{ formError }}</p>
    </div>
  </div>
</template>
