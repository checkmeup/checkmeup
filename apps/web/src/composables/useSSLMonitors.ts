import { useQuery } from '@tanstack/vue-query'
import { monitorsApi } from '@/api/monitors'

export function useSSLMonitors() {
  return useQuery({ queryKey: ['ssl-monitors'], queryFn: monitorsApi.listSSL })
}

export function useSSLMonitor(id: string) {
  return useQuery({ queryKey: ['ssl-monitor', id], queryFn: () => monitorsApi.getSSL(id) })
}
