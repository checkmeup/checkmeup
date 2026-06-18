import { api } from './client'

export interface OrgSettings {
  telegramChatId: string | null
  alertEmail: string | null
  emailAlertsEnabled: boolean
}

export const settingsApi = {
  get: () => api.get<OrgSettings>('/api/v1/settings/'),
  saveTelegram: (chatId: string) => api.put<OrgSettings>('/api/v1/settings/telegram', { chatId }),
  testTelegram: (chatId: string) => api.post<void>('/api/v1/settings/telegram/test', { chatId }),
  saveEmail: (email: string) => api.put<OrgSettings>('/api/v1/settings/email', { email }),
  setEmailAlertsEnabled: (enabled: boolean) =>
    api.put<OrgSettings>('/api/v1/settings/email/enabled', { enabled }),
  testEmail: (email: string) => api.post<void>('/api/v1/settings/email/test', { email }),
}
