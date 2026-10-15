<script setup lang="ts">
// Projects (PRD projects): create, rename and delete them, and choose each
// one's monitors. A project's monitors are listed on /monitors?project=<id>.
import { computed, ref } from 'vue'
import { useQueryClient } from '@tanstack/vue-query'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import MaintenanceMonitorPicker from '@/components/MaintenanceMonitorPicker.vue'
import { projectsApi, type Project } from '@/api/projects'
import { useAddProject, useProjects } from '@/composables/useProjects'
import { useMonitorRows } from '@/composables/useMonitorRows'
import { type MonitorType, monitorCountLabel, monitorTypes } from '@/lib/monitorRows'

type PickedMonitor = { monitorType: MonitorType; monitorId: string; name: string }

const queryClient = useQueryClient()
const { data, isPending: loading, error: queryError, refetch } = useProjects()
const projects = computed<Project[]>(() => data.value ?? [])
const { rows, isComplete: monitorsLoaded } = useMonitorRows()
const addProject = useAddProject()

const actionError = ref('')

// ─── create ───────────────────────────────────────────────────────────────
const newName = ref('')
const creating = ref(false)

async function create() {
  const name = newName.value.trim()
  if (!name) {
    actionError.value = 'Name is required'
    return
  }
  creating.value = true
  actionError.value = ''
  try {
    await addProject(name)
    newName.value = ''
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : 'Failed to create project'
  } finally {
    creating.value = false
  }
}

// ─── one open action per page: rename, delete or monitors ────────────────
type Mode = 'rename' | 'delete' | 'monitors'
const openId = ref('')
const mode = ref<Mode | null>(null)
const renameValue = ref('')
const picked = ref<PickedMonitor[]>([])
const busy = ref(false)

function open(p: Project, m: Mode) {
  actionError.value = ''
  openId.value = p.id
  mode.value = m
  if (m === 'rename') renameValue.value = p.name
  if (m === 'monitors') {
    if (!monitorsLoaded.value) {
      actionError.value = 'Not every monitor list has loaded yet — try again in a moment.'
      close()
      return
    }
    picked.value = rows.value
      .filter((r) => r.projectId === p.id)
      .map((r) => ({ monitorType: r.type, monitorId: r.id, name: r.name }))
  }
}

function close() {
  openId.value = ''
  mode.value = null
}

function isOpen(p: Project, m: Mode) {
  return openId.value === p.id && mode.value === m
}

// Monitors show their project in lists and on their edit forms, so
// assigning or deleting refreshes every type's list and single-monitor
// queries as well as the projects.
async function refreshAfterChange() {
  await Promise.all([
    queryClient.invalidateQueries({ queryKey: ['projects'] }),
    ...monitorTypes.flatMap((t) => [
      queryClient.invalidateQueries({ queryKey: [`${t}-monitors`] }),
      queryClient.invalidateQueries({ queryKey: [`${t}-monitor`] }),
    ]),
  ])
}

async function run(action: () => Promise<unknown>, failure: string, refreshMonitors: boolean) {
  busy.value = true
  actionError.value = ''
  try {
    await action()
    if (refreshMonitors) await refreshAfterChange()
    else await queryClient.invalidateQueries({ queryKey: ['projects'] })
    close()
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : failure
  } finally {
    busy.value = false
  }
}

function rename(p: Project) {
  const name = renameValue.value.trim()
  if (!name) {
    actionError.value = 'Name is required'
    return
  }
  return run(() => projectsApi.rename(p.id, name), 'Failed to rename project', false)
}

function remove(p: Project) {
  return run(() => projectsApi.delete(p.id), 'Failed to delete project', true)
}

// Saving replaces the project's monitors of every type, so it waits for
// every type's list: a list that hasn't loaded would read as "none of
// these", and save would take that type's monitors out of the project.
function saveMonitors(p: Project) {
  if (!monitorsLoaded.value) {
    actionError.value = 'Not every monitor list has loaded yet — try again in a moment.'
    return
  }
  const monitors = picked.value.map((m) => ({ type: m.monitorType, id: m.monitorId }))
  return run(() => projectsApi.setMonitors(p.id, monitors), 'Failed to save monitors', true)
}
</script>

<template>
  <AppLayout>
    <div class="mx-auto flex max-w-[1200px] flex-col gap-8 px-4 pt-8 pb-20 md:px-6">
      <div class="flex flex-col gap-2">
        <h1
          class="text-[30px] leading-[1.1] font-semibold tracking-[-0.035em] text-[var(--text-strong)] md:text-[34px]"
        >
          Projects
        </h1>
        <p class="max-w-[640px] text-sm text-[var(--text-dim)]">
          Group monitors by client or project. The monitor list shows each project as its own
          section, and every alert for a project's monitor starts with the project's name.
        </p>
      </div>

      <form class="flex max-w-[560px] flex-wrap gap-2" @submit.prevent="create">
        <label for="new-project" class="sr-only">New project name</label>
        <Input
          id="new-project"
          v-model="newName"
          placeholder="Client or project name"
          class="min-w-0 flex-1"
        />
        <Button type="submit" :disabled="creating">{{
          creating ? 'Adding…' : 'Add project'
        }}</Button>
      </form>

      <p v-if="actionError" class="text-sm text-[var(--status-down)]">{{ actionError }}</p>

      <div v-if="loading" class="text-sm text-[var(--text-muted)]">Loading…</div>

      <div v-else-if="queryError" class="border-y border-[var(--border)] py-12 text-center">
        <p class="mb-4 text-sm text-[var(--status-down)]">{{ queryError.message }}</p>
        <Button variant="secondary" size="sm" @click="refetch()">Try again</Button>
      </div>

      <div
        v-else-if="projects.length === 0"
        class="border-y border-[var(--border)] py-16 text-center text-sm text-[var(--text-muted)]"
      >
        No projects yet. Add one above, then choose its monitors, or pick a project on a monitor's
        form.
      </div>

      <ul
        v-else
        class="divide-y divide-[var(--border)] rounded-[14px] border border-[var(--border)]"
      >
        <li v-for="p in projects" :key="p.id" class="flex flex-col gap-3 px-4 py-3.5">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <form
              v-if="isOpen(p, 'rename')"
              class="flex min-w-0 flex-1 flex-wrap gap-2"
              @submit.prevent="rename(p)"
            >
              <label :for="`rename-${p.id}`" class="sr-only">Project name</label>
              <Input :id="`rename-${p.id}`" v-model="renameValue" class="min-w-0 flex-1" />
              <Button type="submit" size="sm" :disabled="busy">Save</Button>
              <Button type="button" size="sm" variant="secondary" @click="close">Cancel</Button>
            </form>
            <div v-else class="flex min-w-0 flex-col">
              <span class="truncate font-medium text-[var(--text-strong)]">{{ p.name }}</span>
              <RouterLink
                :to="{ name: 'monitors', query: { project: p.id } }"
                class="text-xs text-[var(--text-dim)] hover:text-[var(--text)] hover:underline"
              >
                {{ monitorCountLabel(p.monitorCount) }}
              </RouterLink>
            </div>

            <div v-if="!isOpen(p, 'rename')" class="flex flex-wrap gap-2">
              <Button size="sm" variant="secondary" @click="open(p, 'monitors')">Monitors</Button>
              <Button size="sm" variant="secondary" @click="open(p, 'rename')">Rename</Button>
              <Button size="sm" variant="secondary" @click="open(p, 'delete')">Delete</Button>
            </div>
          </div>

          <div
            v-if="isOpen(p, 'delete')"
            class="flex flex-wrap items-center gap-3 rounded-[10px] bg-[var(--surface)] px-3 py-2.5"
          >
            <span class="text-sm text-[var(--text)]">
              Delete {{ p.name }}? Its monitors keep running, without a project.
            </span>
            <div class="flex gap-2">
              <Button size="sm" variant="destructive" :disabled="busy" @click="remove(p)">
                Confirm delete
              </Button>
              <Button size="sm" variant="secondary" @click="close">Cancel</Button>
            </div>
          </div>

          <div v-if="isOpen(p, 'monitors')" class="flex flex-col gap-3">
            <MaintenanceMonitorPicker v-model="picked" />
            <p class="text-xs text-[var(--text-muted)]">
              A monitor is in one project at a time: choosing one that's in another project moves it
              here.
            </p>
            <div class="flex gap-2">
              <Button size="sm" :disabled="busy" @click="saveMonitors(p)">
                {{ busy ? 'Saving…' : 'Save monitors' }}
              </Button>
              <Button size="sm" variant="secondary" @click="close">Cancel</Button>
            </div>
          </div>
        </li>
      </ul>
    </div>
  </AppLayout>
</template>
