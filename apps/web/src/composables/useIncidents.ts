import { useQuery } from '@tanstack/vue-query'
import { incidentsApi } from '@/api/incidents'

export function useIncidents() {
  return useQuery({ queryKey: ['incidents'], queryFn: incidentsApi.list })
}

export function useIncident(id: string) {
  return useQuery({ queryKey: ['incident', id], queryFn: () => incidentsApi.get(id) })
}
