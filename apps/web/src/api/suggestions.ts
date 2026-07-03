import { api } from './client'

export const suggestionsApi = {
  submit: (text: string) => api.post('/api/v1/suggestions', { text }),
}
