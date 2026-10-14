import { computed } from 'vue'
import { useCronMonitors } from '@/composables/useCronMonitors'
import { useUptimeMonitors } from '@/composables/useUptimeMonitors'
import { useSSLMonitors } from '@/composables/useSSLMonitors'
import { useDomainMonitors } from '@/composables/useDomainMonitors'
import { usePortMonitors } from '@/composables/usePortMonitors'
import { useDNSMonitors } from '@/composables/useDNSMonitors'
import { useCheckHistory } from '@/composables/useCheckHistory'
import {
  type MonitorRow,
  cronRow,
  dnsRow,
  domainRow,
  portRow,
  sortRows,
  sslRow,
  uptimeRow,
} from '@/lib/monitorRows'

// Every monitor of every type as one sorted list of rows. Each type's query
// still loads and fails on its own — one broken type doesn't hide the rest.
export function useMonitorRows() {
  const cron = useCronMonitors()
  const uptime = useUptimeMonitors()
  const ssl = useSSLMonitors()
  const domain = useDomainMonitors()
  const port = usePortMonitors()
  const dns = useDNSMonitors()
  const queries = [cron, uptime, ssl, domain, port, dns]
  const history = useCheckHistory()

  // type:id → that monitor's hourly ticks, in the shape CheckStrip takes.
  const historyByKey = computed(() => {
    const map = new Map<string, { ok: boolean | null }[]>()
    for (const m of history.data.value?.monitors ?? []) {
      map.set(
        `${m.type}:${m.id}`,
        m.ticks.map((t) => ({ ok: t === null ? null : t === 'up' })),
      )
    }
    return map
  })

  const rows = computed<MonitorRow[]>(() => {
    const now = Date.now()
    const sorted = sortRows([
      ...(cron.data.value ?? []).map((m) => cronRow(m, now)),
      ...(uptime.data.value ?? []).map((m) => uptimeRow(m, now)),
      ...(ssl.data.value ?? []).map((m) => sslRow(m, now)),
      ...(domain.data.value ?? []).map((m) => domainRow(m, now)),
      ...(port.data.value ?? []).map((m) => portRow(m, now)),
      ...(dns.data.value ?? []).map((m) => dnsRow(m, now)),
    ])
    return sorted.map((r) => {
      const ticks = historyByKey.value.get(r.key)
      return ticks ? { ...r, history: ticks } : r
    })
  })

  // Still loading only while no type has answered yet — once any list is in,
  // show what's there rather than a spinner.
  const isPending = computed(
    () => queries.every((q) => q.data.value == null) && queries.some((q) => q.isPending?.value),
  )

  const error = computed(() => queries.map((q) => q.error?.value?.message).find(Boolean) ?? '')

  function refetch() {
    queries.forEach((q) => q.refetch?.())
  }

  return { rows, isPending, error, refetch }
}
