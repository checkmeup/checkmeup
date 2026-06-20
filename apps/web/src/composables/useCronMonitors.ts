import { useQuery } from '@tanstack/vue-query'
import { monitorsApi } from '@/api/monitors'

export function useCronMonitors() {
  return useQuery({ queryKey: ['cron-monitors'], queryFn: monitorsApi.listCron })
}

export function useCronMonitor(id: string) {
  return useQuery({ queryKey: ['cron-monitor', id], queryFn: () => monitorsApi.getCron(id) })
}
