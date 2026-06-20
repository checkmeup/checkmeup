import { useQuery } from '@tanstack/vue-query'
import { statusPagesApi } from '@/api/statusPages'

export function useStatusPages() {
  return useQuery({ queryKey: ['status-pages'], queryFn: statusPagesApi.list })
}

export function useStatusPage(id: string) {
  return useQuery({ queryKey: ['status-page', id], queryFn: () => statusPagesApi.get(id) })
}
