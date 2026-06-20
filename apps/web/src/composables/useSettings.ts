import { useQuery } from '@tanstack/vue-query'
import { settingsApi } from '@/api/settings'

export function useSettings() {
  return useQuery({ queryKey: ['settings'], queryFn: settingsApi.get })
}
