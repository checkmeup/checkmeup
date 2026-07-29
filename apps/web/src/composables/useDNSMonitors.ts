import { useQuery } from '@tanstack/vue-query'
import { monitorsApi } from '@/api/monitors'

export function useDNSMonitors() {
  return useQuery({ queryKey: ['dns-monitors'], queryFn: monitorsApi.listDns })
}

export function useDNSMonitor(id: string) {
  return useQuery({ queryKey: ['dns-monitor', id], queryFn: () => monitorsApi.getDns(id) })
}
