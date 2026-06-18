import { api } from './client'

export const suggestionsApi = {
  submit: (text: string) => api.post<void>('/api/v1/suggestions', { text }),
}
