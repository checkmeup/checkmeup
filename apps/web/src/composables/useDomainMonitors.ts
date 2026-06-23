import { useQuery } from '@tanstack/vue-query'
import { monitorsApi } from '@/api/monitors'

export function useDomainMonitors() {
  return useQuery({ queryKey: ['domain-monitors'], queryFn: monitorsApi.listDomain })
}

export function useDomainMonitor(id: string) {
  return useQuery({ queryKey: ['domain-monitor', id], queryFn: () => monitorsApi.getDomain(id) })
}
