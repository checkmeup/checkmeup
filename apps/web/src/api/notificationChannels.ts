import { api } from './client'

export type NotificationChannelType = 'telegram' | 'email' | 'webhook' | 'slack' | 'sms'

export interface NotificationChannel {
  id: string
  type: NotificationChannelType
  name: string
  // sms config: { phone_number: '+972...', consent_at: '2026-...' } —
  // consent_at is set server-side once, never re-sent by the client (ADR-029)
  config: Record<string, string>
  enabled: boolean
  createdAt: string
  lastDeliveryStatus?: 'success' | 'failed'
  lastDeliveryDetail?: string
  lastDeliveryAt?: string
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
  delete: (id: string) => api.delete(`/api/v1/notification-channels/${id}/`),
  test: (input: TestNotificationChannelInput) =>
    api.post('/api/v1/notification-channels/test', input),
  regenerateWebhookSecret: (id: string) =>
    api.post<NotificationChannel>(`/api/v1/notification-channels/${id}/regenerate-secret`, null),
}
