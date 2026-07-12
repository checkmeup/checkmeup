<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useQueryClient } from '@tanstack/vue-query'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import { incidentsApi, type Incident, type IncidentStatus } from '@/api/incidents'
import { useIncident } from '@/composables/useIncidents'

const router = useRouter()
const route = useRoute()
const id = route.params.id as string
const queryClient = useQueryClient()

const { data: incident, isPending: loading, error: loadError } = useIncident(id)
const error = ref('')
watch(loadError, (e) => {
  if (e) error.value = e.message
})

function setIncident(updated: Incident) {
  queryClient.setQueryData(['incident', id], updated)
}

const severityLabels: Record<string, string> = { minor: 'Minor', major: 'Major', critical: 'Critical' }
const severityColors: Record<string, string> = {
  minor: 'var(--status-paused)',
  major: 'var(--status-degraded)',
  critical: 'var(--status-down)',
}
const statusLabels: Record<string, string> = {
  investigating: 'Investigating',
  identified: 'Identified',
  monitoring: 'Monitoring',
  resolved: 'Resolved',
}
const statusOrder: IncidentStatus[] = ['investigating', 'identified', 'monitoring', 'resolved']

function nextStatus(current: IncidentStatus): IncidentStatus {
  const idx = statusOrder.indexOf(current)
  return statusOrder[Math.min(idx + 1, statusOrder.length - 1)]
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })
}

// ─── title editing (US-2404) ──────────────────────────────────────────────
const editingTitle = ref(false)
const titleDraft = ref('')
const savingTitle = ref(false)

function startEditTitle() {
  if (!incident.value) return
  titleDraft.value = incident.value.title
  editingTitle.value = true
}

async function saveTitle() {
  if (!titleDraft.value.trim()) return
  savingTitle.value = true
  try {
    const updated = await incidentsApi.updateTitle(id, titleDraft.value.trim())
    setIncident(updated)
    editingTitle.value = false
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to update title'
  } finally {
    savingTitle.value = false
  }
}

// ─── post a new update (US-2402) ──────────────────────────────────────────
const newMessage = ref('')
const newStatus = ref<IncidentStatus>('identified')
const posting = ref(false)

watch(
  incident,
  (inc) => {
    if (inc) newStatus.value = nextStatus(inc.status)
  },
  { immediate: true },
)

async function postUpdate() {
  if (!newMessage.value.trim()) {
    error.value = 'Update message is required'
    return
  }
  error.value = ''
  posting.value = true
  try {
    const updated = await incidentsApi.postUpdate(id, newMessage.value.trim(), newStatus.value)
    setIncident(updated)
    newMessage.value = ''
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to post update'
  } finally {
    posting.value = false
  }
}

// ─── edit an existing update's message (US-2404) ──────────────────────────
const editingUpdateId = ref('')
const editDraft = ref('')
const savingUpdate = ref(false)

function startEditUpdate(updateId: string, message: string) {
  editingUpdateId.value = updateId
  editDraft.value = message
}

async function saveUpdateMessage(updateId: string) {
  if (!editDraft.value.trim()) return
  savingUpdate.value = true
  try {
    const updated = await incidentsApi.updateUpdateMessage(id, updateId, editDraft.value.trim())
    setIncident(updated)
    editingUpdateId.value = ''
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to update message'
  } finally {
    savingUpdate.value = false
  }
}

// ─── delete (US-2404) ──────────────────────────────────────────────────────
const confirmDelete = ref(false)
const deleting = ref(false)

async function deleteIncident() {
  deleting.value = true
  try {
    await incidentsApi.delete(id)
    router.push({ name: 'incidents' })
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to delete incident'
    confirmDelete.value = false
  } finally {
    deleting.value = false
  }
}

const isResolved = computed(() => incident.value?.status === 'resolved')
</script>

<template>
  <AppLayout>
    <div class="p-8 max-w-2xl mx-auto">
      <div class="flex items-center gap-3 mb-6">
        <button class="text-sm" style="color: var(--text-muted)" @click="router.push({ name: 'incidents' })">
          ← Back
        </button>
      </div>

      <div v-if="loading" class="text-sm" style="color: var(--text-muted)">Loading…</div>

      <template v-else-if="incident">
        <div class="rounded-xl border p-6 mb-4" style="background-color: var(--surface); border-color: var(--border)">
          <div class="flex items-start justify-between gap-3 mb-3">
            <div v-if="!editingTitle" class="flex items-center gap-2 flex-1">
              <h1 class="text-xl font-semibold" style="color: var(--text-strong)">{{ incident.title }}</h1>
              <button class="text-xs" style="color: var(--text-muted)" @click="startEditTitle">Edit</button>
            </div>
            <div v-else class="flex items-center gap-2 flex-1">
              <Input v-model="titleDraft" class="flex-1" />
              <Button size="sm" :disabled="savingTitle" @click="saveTitle">Save</Button>
              <Button size="sm" variant="secondary" @click="editingTitle = false">Cancel</Button>
            </div>
          </div>

          <div class="flex items-center gap-4 text-sm">
            <span class="inline-flex items-center gap-1.5" :style="{ color: severityColors[incident.severity] }">
              <span class="w-1.5 h-1.5 rounded-full" :style="{ backgroundColor: severityColors[incident.severity] }"></span>
              {{ severityLabels[incident.severity] }}
            </span>
            <span style="color: var(--text-dim)">{{ statusLabels[incident.status] }}</span>
            <span style="color: var(--text-muted)">{{ incident.monitorCount }} monitor{{ incident.monitorCount === 1 ? '' : 's' }}</span>
          </div>

          <div v-if="incident.monitors?.length" class="mt-3 flex flex-wrap gap-1.5">
            <span
              v-for="m in incident.monitors"
              :key="`${m.monitorType}:${m.monitorId}`"
              class="text-xs px-2 py-1 rounded-full"
              style="background-color: var(--surface-raised); color: var(--text-dim)"
            >
              {{ m.name }}
            </span>
          </div>
        </div>

        <!-- Post a new update -->
        <div
          v-if="!isResolved"
          class="rounded-xl border p-6 mb-4"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <Label for="newMessage">Post an update</Label>
          <Input id="newMessage" v-model="newMessage" placeholder="What's changed?" class="mt-1" />
          <div class="flex items-center gap-3 mt-3">
            <select
              v-model="newStatus"
              class="rounded-md border px-3 py-2 text-sm"
              style="background-color: var(--surface-raised); border-color: var(--border); color: var(--text)"
            >
              <option v-for="s in statusOrder" :key="s" :value="s">{{ statusLabels[s] }}</option>
            </select>
            <Button :disabled="posting" @click="postUpdate">{{ posting ? 'Posting…' : 'Post update' }}</Button>
          </div>
        </div>

        <p v-if="error" class="text-sm mb-4" style="color: var(--status-down)">{{ error }}</p>

        <!-- Updates feed -->
        <div class="space-y-3 mb-6">
          <div
            v-for="u in incident.updates"
            :key="u.id"
            class="rounded-xl border p-4"
            style="background-color: var(--surface); border-color: var(--border)"
          >
            <div class="flex items-center justify-between mb-1.5">
              <span class="text-xs font-medium" style="color: var(--text-strong)">{{ statusLabels[u.status] }}</span>
              <div class="flex items-center gap-2">
                <span class="text-xs" style="color: var(--text-muted)">{{ formatDate(u.createdAt) }}</span>
                <button
                  v-if="editingUpdateId !== u.id"
                  class="text-xs"
                  style="color: var(--text-muted)"
                  @click="startEditUpdate(u.id, u.message)"
                >
                  Edit
                </button>
              </div>
            </div>
            <p v-if="editingUpdateId !== u.id" class="text-sm" style="color: var(--text-dim)">{{ u.message }}</p>
            <div v-else class="flex items-center gap-2 mt-1">
              <Input v-model="editDraft" class="flex-1" />
              <Button size="sm" :disabled="savingUpdate" @click="saveUpdateMessage(u.id)">Save</Button>
              <Button size="sm" variant="secondary" @click="editingUpdateId = ''">Cancel</Button>
            </div>
          </div>
        </div>

        <div class="flex justify-end">
          <Button
            v-if="!confirmDelete"
            variant="secondary"
            style="color: var(--status-down)"
            @click="confirmDelete = true"
          >
            Delete incident
          </Button>
          <div v-else class="flex gap-2">
            <Button style="background-color: var(--status-down)" :disabled="deleting" @click="deleteIncident">
              Confirm delete
            </Button>
            <Button variant="secondary" @click="confirmDelete = false">Cancel</Button>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>
