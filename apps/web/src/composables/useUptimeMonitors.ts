import { useQuery } from '@tanstack/vue-query'
import { monitorsApi } from '@/api/monitors'

export function useUptimeMonitors() {
  return useQuery({ queryKey: ['uptime-monitors'], queryFn: monitorsApi.listUptime })
}

export function useUptimeMonitor(id: string) {
  return useQuery({ queryKey: ['uptime-monitor', id], queryFn: () => monitorsApi.getUptime(id) })
}
