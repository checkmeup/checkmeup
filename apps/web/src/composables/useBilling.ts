import { useQuery } from '@tanstack/vue-query'
import { billingApi } from '@/api/billing'

// Shared cache key — BillingView, UptimeMonitorCreateView, and
// UptimeMonitorEditView all need billing info; using the same query key
// here means they share one cached fetch instead of three independent ones.
export function useBilling() {
  return useQuery({ queryKey: ['billing'], queryFn: billingApi.getInfo })
}
