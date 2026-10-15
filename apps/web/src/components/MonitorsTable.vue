<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import StatusLabel from '@/components/StatusLabel.vue'
import CheckStrip from '@/components/CheckStrip.vue'
import {
  type MonitorGroup,
  type MonitorRow,
  type MonitorType,
  detailRouteName,
  groupByProject,
  matchesSearch,
  monitorTypes,
  needsAttention,
  typeLabel,
} from '@/lib/monitorRows'

export type MonitorFilter = 'all' | 'problems' | MonitorType

const props = withDefaults(
  defineProps<{
    rows: MonitorRow[]
    // The org's projects (PRD projects). With any, rows are grouped by
    // project and a project filter appears; with none, the list is flat.
    projects?: { id: string; name: string }[]
  }>(),
  { projects: () => [] },
)

const filter = defineModel<MonitorFilter>('filter', { default: 'all' })
// One project's id, "none" for monitors without a project, or null for all.
const project = defineModel<string | null>('project', { default: null })
const search = ref('')
const router = useRouter()

const isProblem = needsAttention

// "All" and "Problems" always show; a type tab only once it has a monitor.
const tabs = computed(() => [
  { id: 'all' as const, label: 'All', count: props.rows.length },
  { id: 'problems' as const, label: 'Problems', count: props.rows.filter(isProblem).length },
  ...monitorTypes
    .map((t) => ({
      id: t,
      label: typeLabel[t],
      count: props.rows.filter((r) => r.type === t).length,
    }))
    .filter((t) => t.count > 0 || filter.value === t.id),
])

function inProject(r: MonitorRow) {
  if (project.value === null) return true
  if (project.value === 'none') return !r.projectId
  return r.projectId === project.value
}

const visibleRows = computed(() =>
  props.rows
    .filter((r) =>
      filter.value === 'all'
        ? true
        : filter.value === 'problems'
          ? isProblem(r)
          : r.type === filter.value,
    )
    .filter(inProject)
    .filter((r) => matchesSearch(r, search.value)),
)

// Grouped only while the org has projects and no single project is picked;
// otherwise one group without a header.
const grouped = computed(() => props.projects.length > 0 && project.value === null)
const groups = computed<MonitorGroup[]>(() =>
  grouped.value
    ? groupByProject(visibleRows.value, props.projects)
    : [{ key: 'all', name: '', rows: visibleRows.value, attention: 0 }],
)

const collapsed = ref(new Set<string>())
function toggleGroup(key: string) {
  const next = new Set(collapsed.value)
  if (next.has(key)) next.delete(key)
  else next.add(key)
  collapsed.value = next
}

function monitorCount(n: number) {
  return n === 1 ? '1 monitor' : `${n} monitors`
}

function open(row: MonitorRow) {
  router.push({ name: detailRouteName[row.type], params: { id: row.id } })
}
</script>

<template>
  <div class="flex flex-col gap-3.5">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div
        role="tablist"
        aria-label="Filter monitors"
        class="flex flex-wrap gap-0.5 rounded-[10px] bg-[var(--surface-raised)] p-[3px]"
      >
        <button
          v-for="tab in tabs"
          :key="tab.id"
          type="button"
          role="tab"
          :aria-selected="filter === tab.id"
          class="min-h-[34px] rounded-lg px-3 text-[13px] font-medium transition-colors focus-visible:ring-2 focus-visible:ring-[var(--accent)] focus-visible:outline-none"
          :class="
            filter === tab.id
              ? 'bg-[var(--bg)] text-[var(--text-strong)] shadow-[0_1px_2px_rgb(0_0_0/12%)]'
              : 'text-[var(--text-dim)] hover:text-[var(--text)]'
          "
          @click="filter = tab.id"
        >
          {{ tab.label }} <span class="text-[var(--text-muted)]">{{ tab.count }}</span>
        </button>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <label v-if="projects.length > 0" class="flex min-h-10 items-center">
          <span class="sr-only">Project</span>
          <select
            :value="project ?? ''"
            class="h-10 rounded-[10px] border border-[var(--border-strong)] bg-transparent pr-8 pl-3 text-sm text-[var(--text)] focus-visible:ring-2 focus-visible:ring-[var(--accent)] focus-visible:outline-none"
            @change="project = ($event.target as HTMLSelectElement).value || null"
          >
            <option value="">All projects</option>
            <option v-for="p in projects" :key="p.id" :value="p.id">{{ p.name }}</option>
            <option value="none">No project</option>
          </select>
        </label>
        <label
          class="flex min-h-10 items-center gap-2 rounded-[10px] border border-[var(--border-strong)] px-3 text-[var(--text-muted)] focus-within:ring-2 focus-within:ring-[var(--accent)]"
        >
          <svg
            aria-hidden="true"
            width="15"
            height="15"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <circle cx="11" cy="11" r="7" />
            <line x1="20" y1="20" x2="16" y2="16" />
          </svg>
          <span class="sr-only">Search monitors</span>
          <input
            v-model="search"
            type="search"
            placeholder="Search"
            class="w-40 bg-transparent text-sm text-[var(--text)] outline-none placeholder:text-[var(--text-muted)]"
          />
        </label>
      </div>
    </div>

    <!-- Phones: one stacked row per monitor -->
    <ul
      class="divide-y divide-[var(--border)] rounded-[14px] border border-[var(--border)] md:hidden"
    >
      <template v-for="g in groups" :key="g.key">
        <li v-if="grouped" class="bg-[var(--surface)]">
          <button
            type="button"
            class="flex w-full items-center gap-2 px-4 py-2.5 text-left text-[13px]"
            :aria-expanded="!collapsed.has(g.key)"
            @click="toggleGroup(g.key)"
          >
            <span aria-hidden="true" class="w-3 text-[var(--text-muted)]">{{
              collapsed.has(g.key) ? '▸' : '▾'
            }}</span>
            <span class="truncate font-semibold text-[var(--text-strong)]">{{ g.name }}</span>
            <span class="flex-shrink-0 text-[var(--text-muted)]">{{ g.rows.length }}</span>
            <span v-if="g.attention > 0" class="ml-auto flex-shrink-0 text-[var(--status-down)]">
              {{ g.attention }} need attention
            </span>
          </button>
        </li>
        <template v-if="!collapsed.has(g.key)">
          <li v-for="m in g.rows" :key="m.key">
            <RouterLink
              :to="{ name: detailRouteName[m.type], params: { id: m.id } }"
              class="block px-4 py-3 transition-colors hover:bg-[var(--surface)]"
            >
              <div class="flex items-center justify-between gap-3">
                <span class="truncate font-medium text-[var(--text-strong)]">{{ m.name }}</span>
                <StatusLabel :tone="m.tone" :label="m.statusLabel" />
              </div>
              <div class="mt-1 flex items-center justify-between gap-3 text-xs">
                <span class="truncate font-mono text-[var(--text-dim)]">{{ m.target }}</span>
                <span class="flex-shrink-0 font-mono text-[var(--text)]">{{ m.metric }}</span>
              </div>
              <CheckStrip v-if="m.history" size="sm" :checks="m.history" class="mt-2" />
            </RouterLink>
          </li>
        </template>
      </template>
    </ul>

    <!-- Desktop table -->
    <div class="hidden overflow-x-auto rounded-[14px] border border-[var(--border)] md:block">
      <table class="w-full min-w-[960px] border-collapse text-sm">
        <thead>
          <tr class="bg-[var(--surface)] text-left text-[12.5px] text-[var(--text-dim)]">
            <th class="w-[130px] px-4 py-2.5 font-medium">Status</th>
            <th class="px-3 py-2.5 font-medium">Monitor</th>
            <th class="w-[90px] px-3 py-2.5 font-medium">Type</th>
            <th class="w-[220px] px-3 py-2.5 font-medium">Last 48 hours</th>
            <th class="w-[200px] px-3 py-2.5 font-medium">Now</th>
            <th class="w-[100px] px-4 py-2.5 text-right font-medium">Checked</th>
          </tr>
        </thead>
        <tbody v-for="g in groups" :key="g.key">
          <tr v-if="grouped" class="border-t border-[var(--border)] bg-[var(--surface)]">
            <td colspan="6" class="p-0">
              <button
                type="button"
                class="flex w-full items-center gap-2 px-4 py-2.5 text-left text-[13px] focus-visible:ring-2 focus-visible:ring-[var(--accent)] focus-visible:outline-none focus-visible:ring-inset"
                :aria-expanded="!collapsed.has(g.key)"
                @click="toggleGroup(g.key)"
              >
                <span aria-hidden="true" class="w-3 text-[var(--text-muted)]">{{
                  collapsed.has(g.key) ? '▸' : '▾'
                }}</span>
                <span class="font-semibold text-[var(--text-strong)]">{{ g.name }}</span>
                <span class="text-[var(--text-muted)]">{{ monitorCount(g.rows.length) }}</span>
                <span v-if="g.attention > 0" class="text-[var(--status-down)]">
                  · {{ g.attention }} need attention
                </span>
              </button>
            </td>
          </tr>
          <template v-if="!collapsed.has(g.key)">
            <tr
              v-for="m in g.rows"
              :key="m.key"
              class="cursor-pointer border-t border-[var(--border)] transition-colors hover:bg-[var(--surface)]"
              @click="open(m)"
            >
              <td class="px-4 py-2.5">
                <StatusLabel :tone="m.tone" :label="m.statusLabel" />
              </td>
              <td class="max-w-0 px-3 py-2.5">
                <RouterLink
                  :to="{ name: detailRouteName[m.type], params: { id: m.id } }"
                  class="block truncate font-medium text-[var(--text-strong)] focus-visible:underline focus-visible:outline-none"
                  @click.stop
                >
                  {{ m.name }}
                </RouterLink>
                <div class="truncate font-mono text-xs text-[var(--text-dim)]">{{ m.target }}</div>
                <div v-if="m.note" class="truncate text-xs text-[var(--text-muted)]">
                  {{ m.note }}
                </div>
              </td>
              <td class="px-3 py-2.5 text-[13px] text-[var(--text-dim)]">
                {{ typeLabel[m.type] }}
              </td>
              <td class="px-3 py-2.5">
                <CheckStrip v-if="m.history" size="sm" :checks="m.history" />
                <span v-else class="text-xs text-[var(--text-muted)]">—</span>
              </td>
              <td class="px-3 py-2.5">
                <div
                  class="font-mono text-[13.5px]"
                  :class="m.tone === 'down' || m.tone === 'warn' ? 'font-medium' : ''"
                  :style="{
                    color:
                      m.tone === 'down'
                        ? 'var(--status-down)'
                        : m.tone === 'warn'
                          ? 'var(--status-degraded)'
                          : 'var(--text)',
                  }"
                >
                  {{ m.metric }}
                </div>
                <div v-if="m.metricSub" class="truncate text-xs text-[var(--text-dim)]">
                  {{ m.metricSub }}
                </div>
              </td>
              <td class="px-4 py-2.5 text-right font-mono text-xs text-[var(--text-dim)]">
                {{ m.lastChecked }}
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>

    <p v-if="visibleRows.length === 0" class="py-6 text-center text-sm text-[var(--text-dim)]">
      No monitors match this filter.
    </p>
  </div>
</template>
