<script setup lang="ts">
// The project a monitor belongs to (PRD projects), on every monitor form.
// "New project…" turns the select into a name field, so a project can be
// created without leaving the form; the new project is then selected.
import { computed, nextTick, ref } from 'vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import type { Project } from '@/api/projects'
import { useAddProject, useProjects } from '@/composables/useProjects'

const projectId = defineModel<string | null>({ required: true })

const NEW = '__new__'

const addProject = useAddProject()
const { data } = useProjects()
const projects = computed<Project[]>(() => data.value ?? [])

const creating = ref(false)
const newName = ref('')
const saving = ref(false)
const error = ref('')

const selectValue = computed({
  get: () => projectId.value ?? '',
  set(v: string) {
    if (v === NEW) {
      creating.value = true
      error.value = ''
      nextTick(() => document.getElementById('project')?.focus())
      return
    }
    projectId.value = v || null
  },
})

function cancel() {
  creating.value = false
  newName.value = ''
  error.value = ''
}

async function create() {
  const name = newName.value.trim()
  if (!name) {
    error.value = 'Name is required'
    return
  }
  saving.value = true
  error.value = ''
  try {
    const p = await addProject(name)
    projectId.value = p.id
    cancel()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : 'Failed to create project'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div>
    <Label for="project">Project</Label>
    <div v-if="creating" class="mt-1 flex flex-wrap gap-2">
      <Input
        id="project"
        v-model="newName"
        placeholder="Client or project name"
        class="min-w-0 flex-1"
        @keydown.enter.prevent="create"
      />
      <Button type="button" size="sm" :disabled="saving" @click="create">
        {{ saving ? 'Creating…' : 'Create' }}
      </Button>
      <Button type="button" size="sm" variant="secondary" @click="cancel">Cancel</Button>
    </div>
    <select
      v-else
      id="project"
      v-model="selectValue"
      class="mt-1 w-full h-10 rounded-lg border pl-3 text-sm"
    >
      <option value="">No project</option>
      <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
      <option :value="NEW">New project…</option>
    </select>
    <p v-if="error" class="text-xs mt-1 text-[var(--status-down)]">{{ error }}</p>
    <p v-else class="text-xs mt-1 text-[var(--text-muted)]">
      Groups this monitor with others, for example one client's. Its alerts start with the project's
      name.
    </p>
  </div>
</template>
