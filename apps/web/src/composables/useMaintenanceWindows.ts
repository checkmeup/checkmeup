import { useQuery } from '@tanstack/vue-query'
import { maintenanceApi } from '@/api/maintenance'

export function useMaintenanceWindows() {
  return useQuery({ queryKey: ['maintenance-windows'], queryFn: maintenanceApi.list })
}

export function useMaintenanceWindow(id: string) {
  return useQuery({ queryKey: ['maintenance-window', id], queryFn: () => maintenanceApi.get(id) })
}
