import { api } from './client'

export interface OrgSettings {
  telegramChatId: string | null
}

export const settingsApi = {
  get: () => api.get<OrgSettings>('/api/v1/settings/'),
  saveTelegram: (chatId: string) => api.put<OrgSettings>('/api/v1/settings/telegram', { chatId }),
  testTelegram: (chatId: string) => api.post<void>('/api/v1/settings/telegram/test', { chatId }),
}
