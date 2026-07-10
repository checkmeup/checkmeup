import { describe, expect, it } from 'vitest'
import { buildChannelConfig, validateChannelSaveInput } from './notificationChannelTypes'

describe('validateChannelSaveInput', () => {
  it('requires a name', () => {
    expect(validateChannelSaveInput('telegram', '', '-1001', false)).toBe('Name is required')
  })

  it('requires a value', () => {
    expect(validateChannelSaveInput('telegram', 'Ops', '', false)).toBe('Chat ID is required')
  })

  it('requires webhook URLs to start with https://', () => {
    expect(validateChannelSaveInput('webhook', 'Ops', 'http://example.com', false)).toBe(
      'Webhook URL must start with https://',
    )
    expect(validateChannelSaveInput('webhook', 'Ops', 'https://example.com', false)).toBe('')
  })

  it('requires slack URLs to point at hooks.slack.com', () => {
    expect(validateChannelSaveInput('slack', 'Ops', 'https://example.com', false)).toBe(
      'Must be a Slack Incoming Webhook URL (https://hooks.slack.com/...)',
    )
    expect(
      validateChannelSaveInput('slack', 'Ops', 'https://hooks.slack.com/services/x', false),
    ).toBe('')
  })

  it('requires sms numbers to be E.164 and consent to be given', () => {
    expect(validateChannelSaveInput('sms', 'Ops', 'not-a-number', false)).toBe(
      'Phone number must be in E.164 format (e.g. +14155551234)',
    )
    expect(validateChannelSaveInput('sms', 'Ops', '+14155551234', false)).toBe(
      'You must agree to receive SMS alerts at this number before saving',
    )
    expect(validateChannelSaveInput('sms', 'Ops', '+14155551234', true)).toBe('')
  })

  it('passes for a valid telegram/email channel', () => {
    expect(validateChannelSaveInput('telegram', 'Ops', '-1001234567890', false)).toBe('')
    expect(validateChannelSaveInput('email', 'Ops', 'alerts@example.com', false)).toBe('')
  })
})

describe('buildChannelConfig', () => {
  it('keys the value under the type-specific config key', () => {
    expect(buildChannelConfig('telegram', ' -1001 ', false)).toEqual({ chatId: '-1001' })
    expect(buildChannelConfig('email', ' a@b.com ', false)).toEqual({ email: 'a@b.com' })
  })

  it('adds a consent field for sms only', () => {
    expect(buildChannelConfig('sms', '+14155551234', true)).toEqual({
      phone_number: '+14155551234',
      consent: 'true',
    })
    expect(buildChannelConfig('sms', '+14155551234', false)).toEqual({
      phone_number: '+14155551234',
      consent: 'false',
    })
    expect(buildChannelConfig('webhook', 'https://example.com', false)).toEqual({
      url: 'https://example.com',
    })
  })
})
