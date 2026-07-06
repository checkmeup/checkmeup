<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import { useAuthStore } from '@/stores/auth'
import { useCronMonitors } from '@/composables/useCronMonitors'
import { useUptimeMonitors } from '@/composables/useUptimeMonitors'
import { useSSLMonitors } from '@/composables/useSSLMonitors'
import { useDomainMonitors } from '@/composables/useDomainMonitors'
import { usePortMonitors } from '@/composables/usePortMonitors'
import { useBilling } from '@/composables/useBilling'
import type {
  CronMonitor,
  UptimeMonitor,
  SSLMonitor,
  DomainMonitor,
  PortMonitor,
} from '@/api/monitors'
import {
  type MonitorType,
  type Row,
  typeLabel,
  detailRouteName,
  statusColors,
  statusLabels,
  relativeTime,
  fmtPct,
  fmtExpiry,
  buildRow,
} from './dashboardHelpers'

const router = useRouter()
const auth = useAuthStore()

// Per-card queries fail independently — one slow/broken monitor type
// doesn't block the others from rendering.
const { data: cronData } = useCronMonitors()
const { data: uptimeData } = useUptimeMonitors()
const { data: sslData } = useSSLMonitors()
const { data: domainData } = useDomainMonitors()
const { data: portData } = usePortMonitors()
const { data: billingInfo } = useBilling()

const cronRows = computed<Row[]>(() =>
  (cronData.value ?? []).map((m: CronMonitor) =>
    buildRow(
      'cron',
      m.id,
      m.name,
      m.schedule,
      m.status,
      'Next ping',
      relativeTime(m.nextPingAt),
      relativeTime(m.lastPingAt),
    ),
  ),
)

const uptimeRows = computed<Row[]>(() =>
  (uptimeData.value ?? []).map((m: UptimeMonitor) =>
    buildRow(
      'uptime',
      m.id,
      m.name,
      m.url,
      m.status,
      'Uptime 24h',
      fmtPct(m.uptime24h),
      relativeTime(m.lastCheckedAt),
    ),
  ),
)

const sslRows = computed<Row[]>(() =>
  (sslData.value ?? []).map((m: SSLMonitor) =>
    buildRow(
      'ssl',
      m.id,
      m.name,
      m.hostname,
      m.status,
      'Expires in',
      fmtExpiry(m.daysUntilExpiry),
      relativeTime(m.lastCheckedAt),
    ),
  ),
)

const domainRows = computed<Row[]>(() =>
  (domainData.value ?? []).map((m: DomainMonitor) =>
    buildRow(
      'domain',
      m.id,
      m.name,
      m.domain,
      m.status,
      'Expires in',
      fmtExpiry(m.daysUntilExpiry),
      relativeTime(m.lastCheckedAt),
    ),
  ),
)

const portRows = computed<Row[]>(() =>
  (portData.value ?? []).map((m: PortMonitor) =>
    buildRow(
      'port',
      m.id,
      m.name,
      `${m.host}:${m.port}`,
      m.status,
      'Uptime 24h',
      fmtPct(m.uptime24h),
      relativeTime(m.lastCheckedAt),
    ),
  ),
)

const allRows = computed<Row[]>(() => [
  ...cronRows.value,
  ...uptimeRows.value,
  ...sslRows.value,
  ...domainRows.value,
  ...portRows.value,
])

const hasAnyMonitors = computed(
  () =>
    (cronData.value?.length ?? 0) > 0 ||
    (uptimeData.value?.length ?? 0) > 0 ||
    (sslData.value?.length ?? 0) > 0 ||
    (domainData.value?.length ?? 0) > 0 ||
    (portData.value?.length ?? 0) > 0,
)

// Filter chips
const filter = ref<'all' | MonitorType>('all')
const chips = computed(() => {
  const counts: Record<string, number> = { all: allRows.value.length }
  ;(['cron', 'uptime', 'ssl', 'domain', 'port'] as MonitorType[]).forEach((t) => {
    counts[t] = allRows.value.filter((r) => r.type === t).length
  })
  return [
    { id: 'all', label: 'All' },
    ...(['cron', 'uptime', 'ssl', 'domain', 'port'] as MonitorType[]).map((t) => ({
      id: t,
      label: typeLabel[t],
    })),
  ].map((c) => ({ ...c, count: counts[c.id] }))
})
const filteredRows = computed(() =>
  filter.value === 'all' ? allRows.value : allRows.value.filter((r) => r.type === filter.value),
)

// Hero stats
const totalMonitors = computed(() => allRows.value.length)
const healthyMonitors = computed(() => allRows.value.filter((r) => r.status === 'up').length)
const healthyPct = computed(() =>
  totalMonitors.value > 0 ? (healthyMonitors.value / totalMonitors.value) * 100 : 0,
)

const uptimeSamples = computed(() =>
  [...uptimeRows.value, ...portRows.value]
    .map((r) => r.metricValue)
    .filter((v) => v !== '—')
    .map((v) => parseFloat(v)),
)
const avgUptimeStr = computed(() => {
  if (uptimeSamples.value.length === 0) return '—'
  const avg = uptimeSamples.value.reduce((a, b) => a + b, 0) / uptimeSamples.value.length
  return `${avg.toFixed(2)}%`
})

interface AttentionItem {
  key: string
  title: string
  detail: string
  actionLabel: string
  severity: 0 | 1
  color: string
  wash: string
  routeName: string
  id: string
}

const attentionItems = computed<AttentionItem[]>(() => {
  const items: AttentionItem[] = []
  for (const r of allRows.value) {
    if (r.status === 'down') {
      items.push({
        key: r.key,
        title: `${r.name} is down`,
        detail: r.target,
        actionLabel: 'Investigate',
        severity: 0,
        color: 'var(--status-down)',
        wash: 'color-mix(in srgb, var(--status-down) 10%, transparent)',
        routeName: detailRouteName[r.type],
        id: r.id,
      })
    } else if (r.status === 'error' || r.status === 'expired') {
      items.push({
        key: r.key,
        title: r.status === 'expired' ? `${r.name} has expired` : `${r.name} check failed`,
        detail: r.target,
        actionLabel: r.status === 'expired' ? 'Renew' : 'View',
        severity: 0,
        color: 'var(--status-down)',
        wash: 'color-mix(in srgb, var(--status-down) 10%, transparent)',
        routeName: detailRouteName[r.type],
        id: r.id,
      })
    } else if (r.status === 'expiring_soon') {
      items.push({
        key: r.key,
        title: `${r.name} ${r.metricLabel.toLowerCase()} ${r.metricValue}`,
        detail: r.target,
        actionLabel: 'Renew',
        severity: 1,
        color: 'var(--status-degraded)',
        wash: 'color-mix(in srgb, var(--status-degraded) 10%, transparent)',
        routeName: detailRouteName[r.type],
        id: r.id,
      })
    }
  }
  return items.sort((a, b) => a.severity - b.severity)
})

const attentionColor = computed(() => {
  if (attentionItems.value.some((i) => i.severity === 0)) return 'var(--status-down)'
  if (attentionItems.value.length > 0) return 'var(--status-degraded)'
  return 'var(--text-strong)'
})

const attentionNote = computed(() => {
  const n = attentionItems.value.length
  if (n === 0) return 'All clear'
  const first = attentionItems.value[0].title
  return n > 1 ? `${first} + ${n - 1} more` : first
})

const smsUsed = computed(() => billingInfo.value?.smsCreditsUsed ?? 0)
const smsTotal = computed(() => billingInfo.value?.smsCreditsLimit ?? 0)
const smsPct = computed(() => (smsTotal.value > 0 ? (smsUsed.value / smsTotal.value) * 100 : 0))

// Upcoming expirations panel (SSL + domain, sorted soonest-first)
interface ExpiryRow {
  key: string
  type: MonitorType
  id: string
  name: string
  daysUntilExpiry: number
  color: string
}

const expiryRows = computed<ExpiryRow[]>(() => {
  const rows: ExpiryRow[] = []
  ;(sslData.value ?? []).forEach((m: SSLMonitor) => {
    if (m.daysUntilExpiry !== null) {
      rows.push({
        key: `ssl:${m.id}`,
        type: 'ssl',
        id: m.id,
        name: m.name,
        daysUntilExpiry: m.daysUntilExpiry,
        color:
          m.daysUntilExpiry < 7
            ? 'var(--status-down)'
            : m.daysUntilExpiry < 30
              ? 'var(--status-degraded)'
              : 'var(--text-muted)',
      })
    }
  })
  ;(domainData.value ?? []).forEach((m: DomainMonitor) => {
    if (m.daysUntilExpiry !== null) {
      rows.push({
        key: `domain:${m.id}`,
        type: 'domain',
        id: m.id,
        name: m.name,
        daysUntilExpiry: m.daysUntilExpiry,
        color:
          m.daysUntilExpiry < 7
            ? 'var(--status-down)'
            : m.daysUntilExpiry < 30
              ? 'var(--status-degraded)'
              : 'var(--text-muted)',
      })
    }
  })
  return rows.sort((a, b) => a.daysUntilExpiry - b.daysUntilExpiry).slice(0, 8)
})

const hasSslOrDomainMonitors = computed(
  () => (sslData.value?.length ?? 0) > 0 || (domainData.value?.length ?? 0) > 0,
)

// Add-monitor menu
const addMenuOpen = ref(false)
const monitorTypeLinks: { label: string; desc: string; routeName: string }[] = [
  {
    label: 'Cron monitor',
    desc: 'Alert on missed scheduled jobs',
    routeName: 'cron-monitor-create',
  },
  {
    label: 'Uptime monitor',
    desc: 'Ping a URL and detect downtime',
    routeName: 'uptime-monitor-create',
  },
  { label: 'SSL monitor', desc: 'Track certificate expiry', routeName: 'ssl-monitor-create' },
  {
    label: 'Domain monitor',
    desc: 'Track registration expiry',
    routeName: 'domain-monitor-create',
  },
  { label: 'Port monitor', desc: 'Raw TCP connect checks', routeName: 'port-monitor-create' },
]
function goToCreate(routeName: string) {
  addMenuOpen.value = false
  router.push({ name: routeName })
}
</script>

<template>
  <AppLayout>
    <div class="max-w-[1360px] mx-auto px-4 md:px-10 pt-8 pb-16">
      <!-- HEADER -->
      <div class="flex items-start justify-between gap-5 mb-7 flex-wrap">
        <div>
          <h1 class="text-2xl font-bold" style="color: var(--text-strong); letter-spacing: -0.02em">
            Dashboard
          </h1>
          <p class="mt-1 text-sm" style="color: var(--text-muted)">
            Welcome back{{ auth.user?.email ? `, ${auth.user.email}` : '' }}.
          </p>
        </div>
        <div class="relative flex-shrink-0">
          <Button class="gap-1.5" @click="addMenuOpen = !addMenuOpen">
            <svg
              width="13"
              height="13"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="3"
              stroke-linecap="round"
            >
              <line x1="12" y1="5" x2="12" y2="19" />
              <line x1="5" y1="12" x2="19" y2="12" />
            </svg>
            Add monitor
          </Button>
          <template v-if="addMenuOpen">
            <div class="fixed inset-0 z-40" @click="addMenuOpen = false"></div>
            <div
              class="absolute top-[calc(100%+8px)] right-0 w-[250px] rounded-xl border z-50 overflow-hidden"
              style="
                background-color: var(--surface);
                border-color: var(--border);
                box-shadow: 0 20px 48px rgb(0 0 0 / 40%);
              "
            >
              <div
                v-for="mt in monitorTypeLinks"
                :key="mt.routeName"
                class="px-3.5 py-2.5 border-b cursor-pointer transition-colors last:border-0 hover:bg-[var(--surface-raised)]"
                style="border-color: var(--border)"
                @click="goToCreate(mt.routeName)"
              >
                <div class="text-[13px] font-semibold" style="color: var(--text-strong)">
                  {{ mt.label }}
                </div>
                <div class="text-xs mt-0.5" style="color: var(--text-muted)">{{ mt.desc }}</div>
              </div>
            </div>
          </template>
        </div>
      </div>

      <!-- HERO STATS -->
      <div class="grid gap-3.5 sm:grid-cols-2 lg:grid-cols-4 mb-7">
        <div
          class="rounded-2xl border p-[22px]"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div class="text-xs font-medium mb-2.5" style="color: var(--text-dim)">
            Monitors healthy
          </div>
          <div class="flex items-baseline gap-1.5 mb-3">
            <span
              class="font-mono text-[28px] font-semibold"
              style="color: var(--text-strong); letter-spacing: -0.02em"
              >{{ healthyMonitors }}</span
            >
            <span class="text-[15px]" style="color: var(--text-muted)">/ {{ totalMonitors }}</span>
          </div>
          <div class="h-1.5 rounded overflow-hidden" style="background-color: var(--accent-wash)">
            <div
              class="h-full rounded"
              :style="{ backgroundColor: 'var(--accent)', width: `${healthyPct}%` }"
            ></div>
          </div>
          <div class="text-xs mt-2" style="color: var(--text-muted)">
            {{ healthyPct.toFixed(1) }}% healthy
          </div>
        </div>

        <div
          class="rounded-2xl border p-[22px]"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div class="text-xs font-medium mb-2.5" style="color: var(--text-dim)">
            Avg uptime (24h)
          </div>
          <div
            class="font-mono text-[28px] font-semibold"
            style="color: var(--text-strong); letter-spacing: -0.02em"
          >
            {{ avgUptimeStr }}
          </div>
          <div class="text-xs mt-3" style="color: var(--text-muted)">
            across {{ uptimeSamples.length }} uptime/port monitor{{
              uptimeSamples.length === 1 ? '' : 's'
            }}
          </div>
        </div>

        <div
          class="rounded-2xl border p-[22px]"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div class="text-xs font-medium mb-2.5" style="color: var(--text-dim)">
            Needs attention
          </div>
          <div
            class="font-mono text-[28px] font-semibold"
            :style="{ color: attentionColor, letterSpacing: '-0.02em' }"
          >
            {{ attentionItems.length }}
          </div>
          <div class="text-xs mt-3 truncate" style="color: var(--text-muted)">
            {{ attentionNote }}
          </div>
        </div>

        <div
          class="rounded-2xl border p-[22px]"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div class="text-xs font-medium mb-2.5" style="color: var(--text-dim)">
            SMS credits used
          </div>
          <template v-if="smsTotal > 0">
            <div class="flex items-baseline gap-1.5 mb-3">
              <span
                class="font-mono text-[28px] font-semibold"
                style="color: var(--text-strong); letter-spacing: -0.02em"
                >{{ smsUsed }}</span
              >
              <span class="text-[15px]" style="color: var(--text-muted)">/ {{ smsTotal }}</span>
            </div>
            <div
              class="h-1.5 rounded overflow-hidden"
              style="background-color: var(--surface-raised)"
            >
              <div
                class="h-full rounded"
                :style="{ backgroundColor: 'var(--accent)', width: `${smsPct}%` }"
              ></div>
            </div>
            <div class="text-xs mt-2" style="color: var(--text-muted)">this billing cycle</div>
          </template>
          <div v-else class="font-mono text-[15px] font-normal" style="color: var(--text-muted)">
            Not available on your plan
          </div>
        </div>
      </div>

      <!-- NEEDS ATTENTION -->
      <div v-if="attentionItems.length > 0" class="mb-7">
        <div
          class="text-[13px] font-semibold mb-3 flex items-center gap-2"
          style="color: var(--text-strong)"
        >
          <svg
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="var(--status-degraded)"
            stroke-width="2.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path
              d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"
            />
            <line x1="12" y1="9" x2="12" y2="13" />
            <line x1="12" y1="17" x2="12.01" y2="17" />
          </svg>
          Needs attention
        </div>
        <div class="grid gap-3" style="grid-template-columns: repeat(auto-fit, minmax(280px, 1fr))">
          <div
            v-for="item in attentionItems"
            :key="item.key"
            class="rounded-xl border p-4 flex flex-col gap-2 cursor-pointer"
            :style="{
              borderColor: `color-mix(in srgb, ${item.color} 30%, transparent)`,
              backgroundColor: item.wash,
            }"
            @click="router.push({ name: item.routeName, params: { id: item.id } })"
          >
            <div class="text-[13.5px] font-semibold" style="color: var(--text-strong)">
              {{ item.title }}
            </div>
            <div class="text-xs leading-relaxed font-mono" style="color: var(--text-dim)">
              {{ item.detail }}
            </div>
            <span class="text-xs font-semibold mt-0.5" :style="{ color: item.color }"
              >{{ item.actionLabel }} →</span
            >
          </div>
        </div>
      </div>

      <!-- MAIN GRID -->
      <div class="flex gap-5 flex-wrap items-start">
        <!-- MONITORS TABLE -->
        <div
          class="flex-1 min-w-0 rounded-2xl border overflow-hidden"
          style="background-color: var(--surface); border-color: var(--border); flex-basis: 560px"
        >
          <div
            class="flex items-center justify-between gap-3 px-5 py-4 border-b flex-wrap"
            style="border-color: var(--border)"
          >
            <span class="text-sm font-semibold" style="color: var(--text-strong)">Monitors</span>
            <div class="flex gap-1.5 flex-wrap">
              <button
                v-for="chip in chips"
                :key="chip.id"
                class="px-3 py-1 rounded-full text-xs font-medium border transition-colors"
                :style="{
                  borderColor: filter === chip.id ? 'var(--accent)' : 'var(--border)',
                  backgroundColor: filter === chip.id ? 'var(--accent-wash)' : 'transparent',
                  color: filter === chip.id ? 'var(--accent)' : 'var(--text-dim)',
                }"
                @click="filter = chip.id as 'all' | MonitorType"
              >
                {{ chip.label }} <span class="opacity-60">{{ chip.count }}</span>
              </button>
            </div>
          </div>

          <div v-if="!hasAnyMonitors" class="p-10 text-center">
            <p class="text-sm mb-4" style="color: var(--text-muted)">
              No monitors yet. Add a cron, uptime, SSL, domain, or port monitor to get started.
            </p>
            <Button @click="router.push({ name: 'uptime-monitor-create' })"
              >Add your first monitor</Button
            >
          </div>

          <template v-else>
            <!-- Mobile cards -->
            <div class="md:hidden">
              <div
                v-for="m in filteredRows"
                :key="m.key"
                class="p-4 border-b cursor-pointer transition-colors hover:bg-[var(--surface-raised)]"
                style="border-color: var(--border)"
                @click="router.push({ name: detailRouteName[m.type], params: { id: m.id } })"
              >
                <div class="flex items-center justify-between mb-1.5">
                  <span class="font-medium text-sm truncate" style="color: var(--text-strong)">{{
                    m.name
                  }}</span>
                  <span
                    class="inline-flex items-center gap-1.5 text-xs font-medium flex-shrink-0"
                    :style="{ color: statusColors[m.status] }"
                  >
                    <span
                      class="w-1.5 h-1.5 rounded-full"
                      :style="{ backgroundColor: statusColors[m.status] }"
                    ></span>
                    {{ statusLabels[m.status] ?? m.status }}
                  </span>
                </div>
                <div class="flex items-center justify-between text-xs">
                  <span class="font-mono truncate" style="color: var(--text-muted)">{{
                    m.target
                  }}</span>
                  <span class="font-mono flex-shrink-0" style="color: var(--text-dim)">{{
                    m.metricValue
                  }}</span>
                </div>
              </div>
            </div>

            <!-- Desktop table -->
            <div class="hidden md:block overflow-x-auto">
              <table class="w-full text-sm" style="min-width: 640px">
                <thead>
                  <tr style="background-color: var(--surface-raised)">
                    <th class="w-6"></th>
                    <th
                      class="text-left px-3 py-2.5 text-[10.5px] font-semibold uppercase tracking-wider"
                      style="color: var(--text-muted)"
                    >
                      Monitor
                    </th>
                    <th
                      class="text-left px-3 py-2.5 text-[10.5px] font-semibold uppercase tracking-wider"
                      style="color: var(--text-muted)"
                    >
                      Metric
                    </th>
                    <th
                      class="text-right px-5 py-2.5 text-[10.5px] font-semibold uppercase tracking-wider"
                      style="color: var(--text-muted)"
                    >
                      Last checked
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="m in filteredRows"
                    :key="m.key"
                    class="cursor-pointer transition-colors border-b hover:bg-[var(--surface-raised)]"
                    style="border-color: var(--border)"
                    @click="router.push({ name: detailRouteName[m.type], params: { id: m.id } })"
                  >
                    <td class="pl-5">
                      <span
                        class="inline-block w-[9px] h-[9px] rounded-full"
                        :style="{ backgroundColor: statusColors[m.status] }"
                      ></span>
                    </td>
                    <td class="px-3 py-3 min-w-0">
                      <div class="flex items-center gap-1.5">
                        <span class="font-semibold truncate" style="color: var(--text-strong)">{{
                          m.name
                        }}</span>
                        <span
                          class="flex-shrink-0 px-1.5 py-0.5 rounded-full text-[10px] font-medium"
                          style="background-color: var(--surface-raised); color: var(--text-dim)"
                        >
                          {{ typeLabel[m.type] }}
                        </span>
                      </div>
                      <div
                        class="font-mono text-[11.5px] truncate mt-0.5"
                        style="color: var(--text-muted)"
                      >
                        {{ m.target }}
                      </div>
                    </td>
                    <td class="px-3 py-3">
                      <div class="text-[11.5px] mb-0.5" style="color: var(--text-muted)">
                        {{ m.metricLabel }}
                      </div>
                      <div
                        class="font-mono text-[13.5px] font-semibold"
                        :style="{
                          color: ['down', 'error', 'expired', 'expiring_soon'].includes(m.status)
                            ? statusColors[m.status]
                            : 'var(--text-strong)',
                        }"
                      >
                        {{ m.metricValue }}
                      </div>
                    </td>
                    <td
                      class="px-5 py-3 font-mono text-[11.5px] text-right"
                      style="color: var(--text-muted)"
                    >
                      {{ m.lastChecked }}
                    </td>
                  </tr>
                </tbody>
              </table>
              <div
                v-if="filteredRows.length === 0"
                class="p-8 text-center text-sm"
                style="color: var(--text-muted)"
              >
                No monitors match this filter.
              </div>
            </div>
          </template>
        </div>

        <!-- UPCOMING EXPIRATIONS -->
        <div
          class="rounded-2xl border p-5"
          style="
            flex: 1 1 320px;
            min-width: 300px;
            background-color: var(--surface);
            border-color: var(--border);
          "
        >
          <div class="text-sm font-semibold mb-4" style="color: var(--text-strong)">
            Upcoming expirations
          </div>

          <p v-if="!hasSslOrDomainMonitors" class="text-xs" style="color: var(--text-muted)">
            Add an SSL or domain monitor to track certificate and registration expirations here.
          </p>
          <p v-else-if="expiryRows.length === 0" class="text-xs" style="color: var(--text-muted)">
            No expiry data yet — checks run shortly after a monitor is created.
          </p>
          <div v-else class="flex flex-col gap-3">
            <div
              v-for="row in expiryRows"
              :key="row.key"
              class="flex items-center justify-between gap-3 cursor-pointer"
              @click="router.push({ name: detailRouteName[row.type], params: { id: row.id } })"
            >
              <div class="min-w-0">
                <div class="text-[13px] font-medium truncate" style="color: var(--text-strong)">
                  {{ row.name }}
                </div>
                <div class="text-[11px] mt-0.5" style="color: var(--text-muted)">
                  {{ typeLabel[row.type] }}
                </div>
              </div>
              <div
                class="font-mono text-xs font-semibold flex-shrink-0"
                :style="{ color: row.color }"
              >
                {{ row.daysUntilExpiry < 0 ? 'Expired' : `${row.daysUntilExpiry}d` }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>
