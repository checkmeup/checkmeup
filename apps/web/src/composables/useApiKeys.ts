import { useQuery } from '@tanstack/vue-query'
import { apiKeysApi } from '@/api/apiKeys'

export function useApiKeys() {
  return useQuery({
    queryKey: ['api-keys'],
    queryFn: apiKeysApi.list,
  })
}
