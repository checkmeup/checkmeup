import { computed } from 'vue'
import { useCronMonitors } from '@/composables/useCronMonitors'
import { useUptimeMonitors } from '@/composables/useUptimeMonitors'
import { useSSLMonitors } from '@/composables/useSSLMonitors'
import { useDomainMonitors } from '@/composables/useDomainMonitors'
import { usePortMonitors } from '@/composables/usePortMonitors'
import { useDNSMonitors } from '@/composables/useDNSMonitors'
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

  const rows = computed<MonitorRow[]>(() => {
    const now = Date.now()
    return sortRows([
      ...(cron.data.value ?? []).map((m) => cronRow(m, now)),
      ...(uptime.data.value ?? []).map((m) => uptimeRow(m, now)),
      ...(ssl.data.value ?? []).map((m) => sslRow(m, now)),
      ...(domain.data.value ?? []).map((m) => domainRow(m, now)),
      ...(port.data.value ?? []).map((m) => portRow(m, now)),
      ...(dns.data.value ?? []).map((m) => dnsRow(m, now)),
    ])
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
