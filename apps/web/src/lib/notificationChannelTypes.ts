import type { NotificationChannelType } from '@/api/notificationChannels'

// E.164 phone number pattern (US-1901): a leading +, no leading zero, up to
// 15 digits total — mirrors the backend's e164Pattern in
// apps/api/internal/handler/notification_channels.go.
const E164_PATTERN = /^\+[1-9]\d{1,14}$/

// Static per-type metadata, used by both the channel list (NotificationChannelsCard.vue)
// and the add/edit form (useNotificationChannelForm.ts) — kept here rather than inside
// the form composable so the list doesn't need to instantiate any form state just to
// render a label/icon.
export const typeLabel: Record<NotificationChannelType, string> = {
  telegram: 'Telegram',
  email: 'Email',
  webhook: 'Webhook',
  slack: 'Slack',
  sms: 'SMS',
}
export const typeIconPath: Record<NotificationChannelType, string> = {
  telegram: 'M22 2L11 13 M22 2l-7 20-4-9-9-4 20-7z',
  email: 'M4 4h16v16H4z M22 6l-10 7L2 6',
  webhook:
    'M10 13a5 5 0 0 0 7.07 0l1.93-1.93a5 5 0 0 0-7.07-7.07L10.5 5.5 M14 11a5 5 0 0 0-7.07 0L5 12.93a5 5 0 0 0 7.07 7.07L13.5 18.5',
  slack: 'M4 9h16 M4 15h16 M9 4v16 M15 4v16',
  sms: 'M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z',
}
export const configKey: Record<NotificationChannelType, string> = {
  telegram: 'chatId',
  email: 'email',
  webhook: 'url',
  slack: 'url',
  sms: 'phone_number',
}
export const valueLabel: Record<NotificationChannelType, string> = {
  telegram: 'Chat ID',
  email: 'Email address',
  webhook: 'Webhook URL',
  slack: 'Incoming Webhook URL',
  sms: 'Phone number',
}
export const valuePlaceholder: Record<NotificationChannelType, string> = {
  telegram: '-1001234567890',
  email: 'alerts@yourteam.com',
  webhook: 'https://example.com/hooks/checkmeup',
  slack: 'https://hooks.slack.com/services/...',
  sms: '+14155551234',
}

// Returns the first validation failure message, or '' if the form is valid
// to submit. A pure function (no Vue reactivity) so it's testable and
// reusable on its own, independent of the form composable's ref state.
export function validateChannelSaveInput(
  type: NotificationChannelType,
  name: string,
  value: string,
  smsConsent: boolean,
): string {
  if (!name.trim()) {
    return 'Name is required'
  }
  if (!value.trim()) {
    // t only ever ranges over the fixed NotificationChannelType keys of valueLabel,
    // never external input.
    // eslint-disable-next-line security/detect-object-injection
    return `${valueLabel[type]} is required`
  }
  if (type === 'webhook' && !value.trim().startsWith('https://')) {
    return 'Webhook URL must start with https://'
  }
  if (type === 'slack' && !value.trim().startsWith('https://hooks.slack.com/')) {
    return 'Must be a Slack Incoming Webhook URL (https://hooks.slack.com/...)'
  }
  if (type === 'sms') {
    if (!E164_PATTERN.test(value.trim())) {
      return 'Phone number must be in E.164 format (e.g. +14155551234)'
    }
    if (!smsConsent) {
      return 'You must agree to receive SMS alerts at this number before saving'
    }
  }
  return ''
}

// Builds the channel config payload sent to the API. A pure function for
// the same reason as validateChannelSaveInput above.
export function buildChannelConfig(
  type: NotificationChannelType,
  value: string,
  smsConsent: boolean,
): Record<string, string> {
  // type only ever ranges over the fixed NotificationChannelType keys of configKey,
  // never external input.
  // eslint-disable-next-line security/detect-object-injection
  const config: Record<string, string> = { [configKey[type]]: value.trim() }
  if (type === 'sms') {
    config.consent = smsConsent ? 'true' : 'false'
  }
  return config
}
