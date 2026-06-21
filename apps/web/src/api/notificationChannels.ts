import { api } from './client'

export type NotificationChannelType = 'telegram' | 'email'

export interface NotificationChannel {
  id: string
  type: NotificationChannelType
  name: string
  config: Record<string, string>
  enabled: boolean
  createdAt: string
}

export interface NotificationChannelInput {
  type: NotificationChannelType
  name: string
  config: Record<string, string>
  enabled?: boolean
}

export interface TestNotificationChannelInput {
  type: NotificationChannelType
  config: Record<string, string>
}

export const notificationChannelsApi = {
  list: () => api.get<NotificationChannel[]>('/api/v1/notification-channels/'),
  create: (input: NotificationChannelInput) =>
    api.post<NotificationChannel>('/api/v1/notification-channels/', input),
  update: (id: string, input: NotificationChannelInput) =>
    api.patch<NotificationChannel>(`/api/v1/notification-channels/${id}/`, input),
  delete: (id: string) => api.delete<void>(`/api/v1/notification-channels/${id}/`),
  test: (input: TestNotificationChannelInput) =>
    api.post<void>('/api/v1/notification-channels/test', input),
}
