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

// ─── title editing (PRD incidents) ──────────────────────────────────────────────
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

// ─── post a new update (PRD incidents) ──────────────────────────────────────────
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

// ─── edit an existing update's message (PRD incidents) ──────────────────────────
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

// ─── delete (PRD incidents) ──────────────────────────────────────────────────────
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
    <div class="mx-auto max-w-[720px] px-4 pt-8 pb-20 md:px-6">
      <div class="mb-8 flex flex-col items-start gap-4">
        <button type="button" class="py-1 text-sm text-[var(--text-dim)] hover:text-[var(--text)]" @click="router.push({ name: 'incidents' })">
          ← Back
        </button>
      </div>

      <div v-if="loading" class="text-sm text-[var(--text-muted)]">Loading…</div>

      <template v-else-if="incident">
        <div class="mb-10">
          <div class="flex items-start justify-between gap-3 mb-3">
            <div v-if="!editingTitle" class="flex items-center gap-2 flex-1">
              <h1 class="text-[30px] leading-[1.1] font-semibold tracking-[-0.035em] text-[var(--text-strong)] md:text-[34px]">{{ incident.title }}</h1>
              <button type="button" class="text-sm text-[var(--text-dim)] hover:text-[var(--text)]" @click="startEditTitle">Edit</button>
            </div>
            <div v-else class="flex items-center gap-2 flex-1">
              <Input v-model="titleDraft" class="flex-1" />
              <Button size="sm" :disabled="savingTitle" @click="saveTitle">Save</Button>
              <Button size="sm" variant="secondary" @click="editingTitle = false">Cancel</Button>
            </div>
          </div>

          <div class="flex items-center gap-4 text-sm">
            <span class="inline-flex items-center gap-2 font-medium" :style="{ color: severityColors[incident.severity] }">
              <span class="h-2 w-2 rounded-full" :style="{ backgroundColor: severityColors[incident.severity] }"></span>
              {{ severityLabels[incident.severity] }}
            </span>
            <span class="text-[var(--text-dim)]">{{ statusLabels[incident.status] }}</span>
            <span class="text-[var(--text-muted)]">{{ incident.monitorCount }} monitor{{ incident.monitorCount === 1 ? '' : 's' }}</span>
          </div>

          <div v-if="incident.monitors?.length" class="mt-3 flex flex-wrap gap-1.5">
            <span
              v-for="m in incident.monitors"
              :key="`${m.monitorType}:${m.monitorId}`"
              class="text-xs px-2 py-1 rounded-full bg-[var(--surface-raised)] text-[var(--text-dim)]"
            >
              {{ m.name }}
            </span>
          </div>
        </div>

        <!-- Post a new update -->
        <div
          v-if="!isResolved"
          class="mb-10 rounded-[14px] bg-[var(--surface)] p-6"
        >
          <Label for="newMessage">Post an update</Label>
          <Input id="newMessage" v-model="newMessage" placeholder="What's changed?" class="mt-1" />
          <div class="flex items-center gap-3 mt-3">
            <select
              v-model="newStatus"
              class="h-10 rounded-lg border px-3 text-sm"
            >
              <option v-for="s in statusOrder" :key="s" :value="s">{{ statusLabels[s] }}</option>
            </select>
            <Button :disabled="posting" @click="postUpdate">{{ posting ? 'Posting…' : 'Post update' }}</Button>
          </div>
        </div>

        <p v-if="error" class="text-sm mb-4 text-[var(--status-down)]">{{ error }}</p>

        <!-- Updates feed -->
        <div class="mb-6 flex flex-col">
          <h2 class="mb-1 text-[22px] font-semibold tracking-[-0.02em] text-[var(--text-strong)]">Updates</h2>
          <div
            v-for="u in incident.updates"
            :key="u.id"
            class="border-t border-[var(--border)] py-4"
          >
            <div class="flex items-center justify-between mb-1.5">
              <span class="text-sm font-semibold text-[var(--text-strong)]">{{ statusLabels[u.status] }}</span>
              <div class="flex items-center gap-2">
                <span class="font-mono text-xs text-[var(--text-dim)]">{{ formatDate(u.createdAt) }}</span>
                <button
                  v-if="editingUpdateId !== u.id"
                  class="text-xs text-[var(--text-muted)]"
                  @click="startEditUpdate(u.id, u.message)"
                >
                  Edit
                </button>
              </div>
            </div>
            <p v-if="editingUpdateId !== u.id" class="text-[15px] text-[var(--text)]">{{ u.message }}</p>
            <div v-else class="flex items-center gap-2 mt-1">
              <Input v-model="editDraft" class="flex-1" />
              <Button size="sm" :disabled="savingUpdate" @click="saveUpdateMessage(u.id)">Save</Button>
              <Button size="sm" variant="secondary" @click="editingUpdateId = ''">Cancel</Button>
            </div>
          </div>
        </div>

        <div class="flex justify-end">
          <Button class="text-[var(--status-down)]"
            v-if="!confirmDelete"
            variant="secondary"
            @click="confirmDelete = true"
          >
            Delete incident
          </Button>
          <div v-else class="flex gap-2">
            <Button class="bg-[var(--status-down)]" :disabled="deleting" @click="deleteIncident">
              Confirm delete
            </Button>
            <Button variant="secondary" @click="confirmDelete = false">Cancel</Button>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>
