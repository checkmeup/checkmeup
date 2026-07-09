<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import DashboardStatCards from '@/components/DashboardStatCards.vue'
import AttentionPanel from '@/components/AttentionPanel.vue'
import MonitorsTable from '@/components/MonitorsTable.vue'
import UpcomingExpirationsCard from '@/components/UpcomingExpirationsCard.vue'
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

      <DashboardStatCards
        :healthy-monitors="healthyMonitors"
        :total-monitors="totalMonitors"
        :healthy-pct="healthyPct"
        :avg-uptime-str="avgUptimeStr"
        :uptime-samples-count="uptimeSamples.length"
        :attention-color="attentionColor"
        :attention-count="attentionItems.length"
        :attention-note="attentionNote"
        :sms-used="smsUsed"
        :sms-total="smsTotal"
        :sms-pct="smsPct"
      />

      <AttentionPanel :items="attentionItems" />

      <!-- MAIN GRID -->
      <div class="flex gap-5 flex-wrap items-start">
        <MonitorsTable
          :chips="chips"
          :filter="filter"
          :has-any-monitors="hasAnyMonitors"
          :filtered-rows="filteredRows"
          @update:filter="filter = $event"
        />

        <UpcomingExpirationsCard
          :has-ssl-or-domain-monitors="hasSslOrDomainMonitors"
          :expiry-rows="expiryRows"
        />
      </div>
    </div>
  </AppLayout>
</template>
