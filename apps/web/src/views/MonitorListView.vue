<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import AddMonitorMenu from '@/components/AddMonitorMenu.vue'
import MonitorsTable, { type MonitorFilter } from '@/components/MonitorsTable.vue'
import { useMonitorRows } from '@/composables/useMonitorRows'
import { useProjects } from '@/composables/useProjects'
import { monitorTypes } from '@/lib/monitorRows'

const route = useRoute()
const router = useRouter()
const { rows, isPending, error, refetch } = useMonitorRows()
const { data: projectsData } = useProjects()
const projects = computed(() => projectsData.value ?? [])

// Replaces one query parameter, keeping the other filter's.
function setQuery(key: 'type' | 'project', value: string | null) {
  const query = { ...route.query }
  if (value === null) delete query[key]
  else query[key] = value
  router.replace({ query })
}

// The filter lives in ?type= so the old per-type list URLs (redirected here)
// and shared links land on the right tab.
const filter = computed<MonitorFilter>({
  get() {
    const t = route.query.type
    if (t === 'problems') return 'problems'
    return typeof t === 'string' && (monitorTypes as string[]).includes(t)
      ? (t as MonitorFilter)
      : 'all'
  },
  set(value) {
    setQuery('type', value === 'all' ? null : value)
  },
})

// ?project= holds a project's id, or "none" for monitors without one
// (PRD projects), so a project's monitors have a link of their own.
const project = computed<string | null>({
  get() {
    const p = route.query.project
    return typeof p === 'string' && p ? p : null
  },
  set(value) {
    setQuery('project', value)
  },
})
</script>

<template>
  <AppLayout>
    <div class="mx-auto flex max-w-[1200px] flex-col gap-8 px-4 pt-10 pb-20 md:px-6">
      <div class="flex flex-wrap items-end justify-between gap-4">
        <h1
          class="text-[34px] leading-[1.1] font-semibold tracking-[-0.035em] text-[var(--text-strong)]"
        >
          Monitors
        </h1>
        <AddMonitorMenu />
      </div>

      <p v-if="isPending" class="text-sm text-[var(--text-dim)]">Loading…</p>

      <div
        v-else-if="error && rows.length === 0"
        class="flex flex-col items-center gap-4 border-y border-[var(--border)] py-12 text-center"
      >
        <p class="text-sm text-[var(--status-down)]">{{ error }}</p>
        <Button variant="secondary" size="sm" @click="refetch">Try again</Button>
      </div>

      <div
        v-else-if="rows.length === 0"
        class="flex flex-col items-center gap-4 border-y border-[var(--border)] py-16 text-center"
      >
        <p class="text-[var(--text-dim)]">
          No monitors yet. Add a cron, uptime, SSL, domain, port, or DNS monitor to get started.
        </p>
        <AddMonitorMenu />
      </div>

      <MonitorsTable
        v-else
        v-model:filter="filter"
        v-model:project="project"
        :rows="rows"
        :projects="projects"
      />
    </div>
  </AppLayout>
</template>
