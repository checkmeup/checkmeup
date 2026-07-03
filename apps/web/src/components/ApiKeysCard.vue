<script setup lang="ts">
import { computed, ref } from 'vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import { apiKeysApi, type CreatedApiKey } from '@/api/apiKeys'
import { useApiKeys } from '@/composables/useApiKeys'

const { data, isPending: loading, refetch } = useApiKeys()
const keys = computed(() => data.value ?? [])

const showForm = ref(false)
const label = ref('')
const creating = ref(false)
const formError = ref('')
const revokingId = ref('')

// Shown once, right after creation — the raw key is never retrievable again.
const createdKey = ref<CreatedApiKey | null>(null)
const copied = ref(false)

// Mirrors the small per-view relative-time helper used elsewhere (e.g.
// NotificationChannelsCard.vue) rather than a new shared util.
function relativeTime(iso: string | null): string {
  if (!iso) return 'Never used'
  const diff = Date.now() - new Date(iso).getTime()
  const m = Math.floor(diff / 60000)
  const h = Math.floor(m / 60)
  if (m < 1) return 'used just now'
  if (m < 60) return `used ${m} min ago`
  if (h < 24) return `used ${h}h ago`
  return `used ${Math.floor(h / 24)}d ago`
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

function startCreate() {
  label.value = ''
  formError.value = ''
  showForm.value = true
}

function cancelForm() {
  showForm.value = false
}

async function create() {
  creating.value = true
  formError.value = ''
  try {
    createdKey.value = await apiKeysApi.create(label.value.trim())
    showForm.value = false
    await refetch()
  } catch (e: unknown) {
    formError.value = e instanceof Error ? e.message : 'Failed to generate key'
  } finally {
    creating.value = false
  }
}

function copyKey() {
  if (!createdKey.value) return
  navigator.clipboard.writeText(createdKey.value.key)
  copied.value = true
}

function dismissCreatedKey() {
  createdKey.value = null
  copied.value = false
}

async function revoke(id: string) {
  revokingId.value = id
  try {
    await apiKeysApi.revoke(id)
    await refetch()
  } finally {
    revokingId.value = ''
  }
}
</script>

<template>
  <div
    class="rounded-xl border p-6"
    style="background-color: var(--surface); border-color: var(--border)"
  >
    <h2 class="font-medium mb-1" style="color: var(--text-strong)">API keys</h2>
    <p class="text-sm mb-5" style="color: var(--text-muted)">
      Generate a key to read monitor status from scripts, CI pipelines, or third-party integrations
      via
      <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
        >X-API-Key</code
      >. Keys are read-only for now.
    </p>

    <div
      v-if="createdKey"
      class="rounded-lg border p-4 space-y-3 mb-5"
      style="border-color: var(--status-up); background-color: var(--surface-raised)"
    >
      <p class="text-sm font-medium" style="color: var(--text-strong)">
        Copy this key now — you won't be able to see it again.
      </p>
      <div class="flex items-center gap-2">
        <code
          class="flex-1 text-xs px-3 py-2 rounded-md truncate"
          style="background-color: var(--surface); color: var(--text-dim)"
        >
          {{ createdKey.key }}
        </code>
        <Button variant="secondary" size="sm" @click="copyKey">{{
          copied ? 'Copied!' : 'Copy'
        }}</Button>
      </div>
      <Button size="sm" @click="dismissCreatedKey">Done</Button>
    </div>

    <div v-if="loading" class="text-sm" style="color: var(--text-muted)">Loading…</div>

    <ul v-else-if="keys.length > 0" class="space-y-2 mb-5">
      <li
        v-for="k in keys"
        :key="k.id"
        class="flex items-center gap-3 rounded-lg border px-4 py-3"
        style="border-color: var(--border)"
      >
        <div class="flex-1 min-w-0">
          <p class="text-sm truncate" style="color: var(--text)">{{ k.label || '(no label)' }}</p>
          <p class="text-xs truncate font-mono" style="color: var(--text-muted)">
            {{ k.keyPrefix }}… · created {{ formatDate(k.createdAt) }} ·
            {{ relativeTime(k.lastUsedAt) }}
          </p>
        </div>
        <Button variant="secondary" size="sm" :disabled="revokingId === k.id" @click="revoke(k.id)">
          {{ revokingId === k.id ? 'Revoking…' : 'Revoke' }}
        </Button>
      </li>
    </ul>
    <p v-else class="text-sm mb-5" style="color: var(--text-muted)">No API keys yet.</p>

    <Button v-if="!showForm" variant="secondary" @click="startCreate">+ Generate key</Button>

    <div
      v-else
      class="rounded-lg border p-4 space-y-4"
      style="border-color: var(--border); background-color: var(--surface-raised)"
    >
      <div>
        <Label for="key-label">Label (optional)</Label>
        <Input id="key-label" v-model="label" placeholder="e.g. CI integration" class="mt-1" />
      </div>

      <div class="flex items-center gap-3">
        <Button :disabled="creating" @click="create">
          {{ creating ? 'Generating…' : 'Generate key' }}
        </Button>
        <Button variant="secondary" type="button" @click="cancelForm">Cancel</Button>
      </div>

      <p v-if="formError" class="text-sm" style="color: var(--status-down)">{{ formError }}</p>
    </div>
  </div>
</template>
