import { useQuery } from '@tanstack/vue-query'
import { monitorsApi } from '@/api/monitors'

export function usePortMonitors() {
  return useQuery({ queryKey: ['port-monitors'], queryFn: monitorsApi.listPort })
}

export function usePortMonitor(id: string) {
  return useQuery({ queryKey: ['port-monitor', id], queryFn: () => monitorsApi.getPort(id) })
}
