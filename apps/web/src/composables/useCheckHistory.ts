import { useQuery } from '@tanstack/vue-query'
import { monitorsApi } from '@/api/monitors'

export function useCheckHistory() {
  return useQuery({ queryKey: ['check-history'], queryFn: monitorsApi.checkHistory })
}
