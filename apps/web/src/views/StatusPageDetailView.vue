<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import { statusPagesApi } from '@/api/statusPages'
import { useStatusPage } from '@/composables/useStatusPages'
import { useCronMonitors } from '@/composables/useCronMonitors'
import { useUptimeMonitors } from '@/composables/useUptimeMonitors'
import { useSSLMonitors } from '@/composables/useSSLMonitors'
import { useDomainMonitors } from '@/composables/useDomainMonitors'
import { usePortMonitors } from '@/composables/usePortMonitors'
import { useDNSMonitors } from '@/composables/useDNSMonitors'

const router = useRouter()
const route = useRoute()
const id = route.params.id as string

const {
  data: detail,
  isPending: pageLoading,
  error: pageError,
  refetch: refetchPage,
} = useStatusPage(id)
const { data: cronData, isPending: cronLoading } = useCronMonitors()
const { data: uptimeData, isPending: uptimeLoading } = useUptimeMonitors()
const { data: sslData, isPending: sslLoading } = useSSLMonitors()
const { data: domainData, isPending: domainLoading } = useDomainMonitors()
const { data: portData, isPending: portLoading } = usePortMonitors()
const { data: dnsData, isPending: dnsLoading } = useDNSMonitors()

const loading = computed(
  () =>
    pageLoading.value ||
    cronLoading.value ||
    uptimeLoading.value ||
    sslLoading.value ||
    domainLoading.value ||
    portLoading.value ||
    dnsLoading.value,
)
const error = computed(() => pageError.value?.message ?? '')
const actionError = ref('')
const confirmDelete = ref(false)
const savingMonitors = ref(false)

// All monitors from all types
const cronMonitors = computed(() => cronData.value ?? [])
const uptimeMonitors = computed(() => uptimeData.value ?? [])
const sslMonitors = computed(() => sslData.value ?? [])
const domainMonitors = computed(() => domainData.value ?? [])
const portMonitors = computed(() => portData.value ?? [])
const dnsMonitors = computed(() => dnsData.value ?? [])

// Local editable list of selected monitors (copy from detail)
interface MonitorEntry {
  monitorType: 'cron' | 'uptime' | 'ssl' | 'domain' | 'port' | 'dns'
  monitorId: string
  displayName: string
  displayOrder: number
  // UI helper
  key: string
}

const monitorEntries = ref<MonitorEntry[]>([])

watch(
  detail,
  (page) => {
    if (!page) return
    monitorEntries.value = page.monitors.map((m, i) => ({
      ...m,
      key: `${m.monitorType}:${m.monitorId}`,
      displayOrder: i,
    }))
  },
  { immediate: true },
)

// All available monitors across all types
const allMonitors = computed(() => {
  const result: {
    key: string
    type: 'cron' | 'uptime' | 'ssl' | 'domain' | 'port' | 'dns'
    id: string
    name: string
  }[] = []
  cronMonitors.value.forEach((m) =>
    result.push({ key: `cron:${m.id}`, type: 'cron', id: m.id, name: m.name }),
  )
  uptimeMonitors.value.forEach((m) =>
    result.push({ key: `uptime:${m.id}`, type: 'uptime', id: m.id, name: m.name }),
  )
  sslMonitors.value.forEach((m) =>
    result.push({ key: `ssl:${m.id}`, type: 'ssl', id: m.id, name: m.name }),
  )
  domainMonitors.value.forEach((m) =>
    result.push({ key: `domain:${m.id}`, type: 'domain', id: m.id, name: m.name }),
  )
  portMonitors.value.forEach((m) =>
    result.push({ key: `port:${m.id}`, type: 'port', id: m.id, name: m.name }),
  )
  dnsMonitors.value.forEach((m) =>
    result.push({ key: `dns:${m.id}`, type: 'dns', id: m.id, name: m.name }),
  )
  return result
})

const selectedKeys = computed(() => new Set(monitorEntries.value.map((e) => e.key)))

function toggleMonitor(m: {
  key: string
  type: 'cron' | 'uptime' | 'ssl' | 'domain' | 'port' | 'dns'
  id: string
  name: string
}) {
  if (selectedKeys.value.has(m.key)) {
    monitorEntries.value = monitorEntries.value.filter((e) => e.key !== m.key)
  } else {
    monitorEntries.value = [
      ...monitorEntries.value,
      {
        monitorType: m.type,
        monitorId: m.id,
        displayName: m.name,
        displayOrder: monitorEntries.value.length,
        key: m.key,
      },
    ]
  }
}

function moveUp(i: number) {
  if (i === 0) return
  const arr = [...monitorEntries.value]
  ;[arr[i - 1], arr[i]] = [arr[i], arr[i - 1]]
  monitorEntries.value = arr.map((e, j) => ({ ...e, displayOrder: j }))
}

function moveDown(i: number) {
  if (i === monitorEntries.value.length - 1) return
  const arr = [...monitorEntries.value]
  ;[arr[i], arr[i + 1]] = [arr[i + 1], arr[i]]
  monitorEntries.value = arr.map((e, j) => ({ ...e, displayOrder: j }))
}

async function saveMonitors() {
  savingMonitors.value = true
  actionError.value = ''
  try {
    const monitors = monitorEntries.value.map((e, i) => ({
      monitorType: e.monitorType,
      monitorId: e.monitorId,
      displayName: e.displayName,
      displayOrder: i,
    }))
    await statusPagesApi.setMonitors(id, { monitors })
    await refetchPage()
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : 'Failed to save monitors'
  } finally {
    savingMonitors.value = false
  }
}

async function deletePage() {
  actionError.value = ''
  try {
    await statusPagesApi.delete(id)
    router.push({ name: 'status-pages' })
  } catch (e: unknown) {
    actionError.value = e instanceof Error ? e.message : 'Delete failed'
    confirmDelete.value = false
  }
}

const typeLabel: Record<string, string> = {
  cron: 'Cron',
  uptime: 'Uptime',
  ssl: 'SSL',
  domain: 'Domain',
  port: 'Port',
  dns: 'DNS',
}

// ─── badges (PRD status-badges) ──────────────────────────────────────────────────────────

interface Badge {
  key: string
  label: string
  badgeUrl: string
  markdown: string
  html: string
}

function toBadge(label: string, badgeUrl: string, publicUrl: string, key: string): Badge {
  return {
    key,
    label,
    badgeUrl,
    markdown: `![${label}](${badgeUrl})`,
    html: `<a href="${publicUrl}"><img src="${badgeUrl}" alt="${label}"></a>`,
  }
}

const pageBadge = computed<Badge | null>(() => {
  if (!detail.value) return null
  return toBadge(
    detail.value.title,
    `${detail.value.publicUrl}/badge.svg`,
    detail.value.publicUrl,
    'page',
  )
})

const monitorBadges = computed<Badge[]>(() => {
  if (!detail.value) return []
  return detail.value.monitors.map((m) =>
    toBadge(
      m.displayName,
      `${detail.value!.publicUrl}/badge/${m.monitorId}.svg`,
      detail.value!.publicUrl,
      `monitor:${m.monitorId}`,
    ),
  )
})

const copiedKey = ref<string | null>(null)
let copiedTimer: ReturnType<typeof setTimeout> | undefined

function copySnippet(key: string, format: 'markdown' | 'html', snippet: string) {
  navigator.clipboard.writeText(snippet)
  copiedKey.value = `${key}:${format}`
  clearTimeout(copiedTimer)
  copiedTimer = setTimeout(() => {
    copiedKey.value = null
  }, 1500)
}
</script>

<template>
  <AppLayout>
    <div class="mx-auto flex max-w-[1200px] flex-col gap-10 px-4 pt-8 pb-20 md:px-6">
      <div class="flex flex-col gap-5">
        <RouterLink
          :to="{ name: 'status-pages' }"
          class="self-start py-1 text-sm text-[var(--text-dim)] hover:text-[var(--text)]"
        >
          ← Status pages
        </RouterLink>
        <div class="flex flex-wrap items-start justify-between gap-6">
          <div class="flex max-w-[720px] min-w-0 flex-col gap-2">
            <h1
              class="text-[34px] leading-[1.08] font-semibold tracking-[-0.035em] break-words text-[var(--text-strong)] md:text-[44px]"
            >
              {{ detail?.title ?? 'Status page' }}
            </h1>
            <template v-if="detail">
              <p class="font-mono text-sm text-[var(--text-dim)]">/status/{{ detail.slug }}</p>
              <p v-if="detail.description" class="text-[15px] text-[var(--text-dim)]">
                {{ detail.description }}
              </p>
            </template>
          </div>
          <div v-if="detail" class="flex flex-wrap items-center gap-2">
            <a
              :href="detail.publicUrl"
              target="_blank"
              rel="noopener"
              class="inline-flex h-10 items-center rounded-lg border border-[var(--border-strong)] px-4 text-sm font-medium text-[var(--text)] transition-colors hover:bg-[var(--surface-raised)]"
            >
              View public page →
            </a>
            <Button
              variant="secondary"
              @click="router.push({ name: 'status-page-edit', params: { id } })"
            >
              Edit
            </Button>
            <Button
              v-if="!confirmDelete"
              variant="secondary"
              class="text-[var(--status-down)]"
              @click="confirmDelete = true"
            >
              Delete
            </Button>
            <template v-else>
              <Button
                class="bg-[var(--status-down)] text-[var(--on-status-down)]"
                @click="deletePage"
              >
                Confirm delete
              </Button>
              <Button variant="secondary" @click="confirmDelete = false">Cancel</Button>
            </template>
          </div>
        </div>
      </div>

      <div v-if="loading" class="text-sm text-[var(--text-muted)]">Loading…</div>
      <div v-else-if="error" class="text-sm text-[var(--status-down)]">{{ error }}</div>

      <template v-else-if="detail">
        <p v-if="actionError" class="text-sm text-[var(--status-down)]">{{ actionError }}</p>

        <!-- Monitor management -->
        <section class="flex flex-col gap-4">
          <h2 class="text-xl font-semibold tracking-[-0.02em] text-[var(--text-strong)]">
            Monitors
          </h2>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
            <!-- Available monitors -->
            <div class="rounded-[14px] border border-[var(--border)]">
              <div
                class="rounded-t-[14px] border-b border-[var(--border)] bg-[var(--surface)] px-4 py-2.5 text-[13px] font-medium text-[var(--text-dim)]"
              >
                Available monitors
              </div>
              <div
                v-if="allMonitors.length === 0"
                class="px-4 py-3 text-xs text-[var(--text-muted)]"
              >
                No monitors yet.
              </div>
              <ul v-else>
                <li
                  v-for="m in allMonitors"
                  :key="m.key"
                  class="flex items-center gap-3 px-4 py-2.5 cursor-pointer transition-colors border-b last:border-0 border-[var(--border)]"
                  @click="toggleMonitor(m)"
                >
                  <div
                    class="w-4 h-4 rounded border flex-shrink-0 flex items-center justify-center text-xs font-bold"
                    :style="{
                      backgroundColor: selectedKeys.has(m.key) ? 'var(--accent)' : 'transparent',
                      borderColor: selectedKeys.has(m.key)
                        ? 'var(--accent)'
                        : 'var(--border-strong)',
                      color: 'var(--on-accent)',
                    }"
                  >
                    {{ selectedKeys.has(m.key) ? '✓' : '' }}
                  </div>
                  <span class="text-sm flex-1 truncate text-[var(--text)]">{{ m.name }}</span>
                  <span class="text-xs flex-shrink-0 text-[var(--text-muted)]">{{
                    typeLabel[m.type]
                  }}</span>
                </li>
              </ul>
            </div>

            <!-- Selected monitors + order -->
            <div class="rounded-[14px] border border-[var(--border)]">
              <div
                class="rounded-t-[14px] border-b border-[var(--border)] bg-[var(--surface)] px-4 py-2.5 text-[13px] font-medium text-[var(--text-dim)]"
              >
                On this page ({{ monitorEntries.length }})
              </div>
              <div
                v-if="monitorEntries.length === 0"
                class="px-4 py-3 text-xs text-[var(--text-muted)]"
              >
                Select monitors from the left.
              </div>
              <ul v-else>
                <li
                  v-for="(m, i) in monitorEntries"
                  :key="m.key"
                  class="flex items-center gap-2 px-4 py-2.5 border-b last:border-0 border-[var(--border)]"
                >
                  <div class="flex flex-col gap-0.5">
                    <button
                      class="text-xs leading-none px-1 text-[var(--text-muted)]"
                      :disabled="i === 0"
                      @click="moveUp(i)"
                    >
                      ▲
                    </button>
                    <button
                      class="text-xs leading-none px-1 text-[var(--text-muted)]"
                      :disabled="i === monitorEntries.length - 1"
                      @click="moveDown(i)"
                    >
                      ▼
                    </button>
                  </div>
                  <input
                    v-model="m.displayName"
                    class="flex-1 text-sm bg-transparent border-0 focus:outline-none min-w-0 text-[var(--text)]"
                    :placeholder="m.displayName"
                  />
                  <span class="text-xs flex-shrink-0 text-[var(--text-muted)]">
                    {{ typeLabel[m.monitorType] }}
                  </span>
                  <button
                    class="text-xs ml-1 flex-shrink-0 text-[var(--text-muted)]"
                    @click="
                      toggleMonitor({
                        key: m.key,
                        type: m.monitorType,
                        id: m.monitorId,
                        name: m.displayName,
                      })
                    "
                  >
                    ✕
                  </button>
                </li>
              </ul>
              <div class="px-4 py-3 border-t border-[var(--border)]">
                <Button size="sm" :disabled="savingMonitors" @click="saveMonitors">
                  {{ savingMonitors ? 'Saving…' : 'Save changes' }}
                </Button>
              </div>
            </div>
          </div>
        </section>

        <!-- Badges (PRD status-badges) -->
        <section class="flex flex-col gap-4">
          <div class="flex flex-col gap-1">
            <h2 class="text-xl font-semibold tracking-[-0.02em] text-[var(--text-strong)]">
              Badges
            </h2>
            <p class="text-sm text-[var(--text-dim)]">
              Live status images to embed in a README or a client's site.
            </p>
          </div>
          <div class="rounded-[14px] border border-[var(--border)]">
            <ul v-if="pageBadge">
              <li
                v-for="badge in [pageBadge, ...monitorBadges]"
                :key="badge.key"
                class="px-4 py-3 border-b last:border-0 border-[var(--border)]"
              >
                <div class="flex items-center gap-3">
                  <img :src="badge.badgeUrl" :alt="badge.label" class="flex-shrink-0" />
                  <span class="text-sm flex-1 truncate text-[var(--text)]">
                    {{ badge.key === 'page' ? 'Overall status' : badge.label }}
                  </span>
                  <Button
                    variant="secondary"
                    size="sm"
                    @click="copySnippet(badge.key, 'markdown', badge.markdown)"
                  >
                    {{ copiedKey === `${badge.key}:markdown` ? 'Copied!' : 'Copy Markdown' }}
                  </Button>
                  <Button
                    variant="secondary"
                    size="sm"
                    @click="copySnippet(badge.key, 'html', badge.html)"
                  >
                    {{ copiedKey === `${badge.key}:html` ? 'Copied!' : 'Copy HTML' }}
                  </Button>
                </div>
              </li>
            </ul>
          </div>
        </section>
      </template>
    </div>
  </AppLayout>
</template>
