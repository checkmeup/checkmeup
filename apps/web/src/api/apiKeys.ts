import { api } from './client'

export interface ApiKey {
  id: string
  label: string
  keyPrefix: string
  createdAt: string
  lastUsedAt: string | null
}

export interface CreatedApiKey extends ApiKey {
  key: string
}

export const apiKeysApi = {
  list: () => api.get<ApiKey[]>('/api/v1/api-keys/'),
  create: (label: string) => api.post<CreatedApiKey>('/api/v1/api-keys/', { label }),
  revoke: (id: string) => api.delete(`/api/v1/api-keys/${id}`),
}
