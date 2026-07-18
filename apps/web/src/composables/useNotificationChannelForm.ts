import { computed, ref, watch } from 'vue'
import { ApiError } from '@/api/client'
import {
  notificationChannelsApi,
  type NotificationChannel,
  type NotificationChannelType,
} from '@/api/notificationChannels'
import {
  buildChannelConfig,
  configKey,
  typeIconPath,
  typeLabel,
  validateChannelSaveInput,
} from '@/lib/notificationChannelTypes'

// typeLabel/configKey/buildChannelConfig/validateChannelSaveInput/etc. live in
// lib/notificationChannelTypes.ts, not here — they're static/pure and shared with
// the channel list (NotificationChannelsCard.vue), which shouldn't have to
// instantiate this form's reactive state just to render a label or icon.

// Holds the add/edit form's state and actions for NotificationChannelsCard —
// split out from the component so the list-management half (toggle/remove)
// isn't tangled up with this form's much larger surface area.
export function useNotificationChannelForm(opts: {
  billingInfo: { value: { plan: string } | undefined }
  refetch: () => Promise<unknown>
}) {
  const { billingInfo, refetch } = opts

  const showForm = ref(false)
  const editingId = ref<string | null>(null)
  const type = ref<NotificationChannelType>('telegram')
  const name = ref('')
  const value = ref('')
  const enabled = ref(true)
  // Only populated when editing an existing webhook channel — the secret
  // doesn't exist until the channel is first saved (US-1401).
  const secret = ref('')
  // TCPA-style opt-in checkbox (US-1901, ADR-029) — required before any sms
  // channel can be saved or tested. Pre-checked when editing an existing
  // channel that already has consent on file (consentAt below).
  const smsConsent = ref(false)
  const consentAt = ref('')
  const consentedPhone = ref('')
  // True once the phone number has been edited away from the one consent_at
  // applies to — a changed number is a new recipient, so it needs fresh
  // consent (ADR-029), same rule the backend enforces in
  // resolveUpdatedChannelConfig.
  const smsConsentOnFile = computed(
    () => !!consentAt.value && value.value.trim() === consentedPhone.value,
  )
  // SMS is a paid-plan-only channel (server-enforced in Create/TestNotificationChannel);
  // hide it from the picker on Hobby rather than letting the user pick it and
  // hit a 402 on save. Still shown (but disabled, since the type select is
  // disabled while editing) if an existing channel is already sms — e.g. an
  // org that created one on a paid plan and later downgraded.
  const hobbyPlanNoSms = computed(() => billingInfo.value?.plan === 'hobby' && type.value !== 'sms')

  const channelTypeOptions = computed(() =>
    (Object.keys(typeLabel) as NotificationChannelType[])
      .filter((t) => t !== 'sms' || !hobbyPlanNoSms.value)
      // t only ever ranges over the fixed NotificationChannelType keys of
      // typeLabel/typeIconPath above, never external input.
      // eslint-disable-next-line security/detect-object-injection
      .map((t) => ({ value: t, label: typeLabel[t], iconPath: typeIconPath[t] })),
  )

  // Editing the number away from the one consent is on file for un-checks the
  // box — a changed number needs a fresh, conscious opt-in, not a carried-over
  // checkmark from the previous number.
  watch(value, () => {
    if (type.value === 'sms' && !smsConsentOnFile.value) {
      smsConsent.value = false
    }
  })

  const saving = ref(false)
  const testing = ref(false)
  const regeneratingSecret = ref(false)
  const regenerateError = ref('')
  const formError = ref('')
  const testSuccess = ref(false)
  const testError = ref('')
  // Set when the API rejects an sms save/test with plan_limit_reached (Hobby
  // plan — interim guard ahead of ADR-032's credit quotas), same pattern as
  // the monitor/status-page create views' plan-limit handling.
  const limitReached = ref(false)

  function startAdd() {
    editingId.value = null
    type.value = 'telegram'
    name.value = ''
    value.value = ''
    secret.value = ''
    smsConsent.value = false
    consentAt.value = ''
    consentedPhone.value = ''
    enabled.value = true
    formError.value = ''
    testSuccess.value = false
    testError.value = ''
    regenerateError.value = ''
    limitReached.value = false
    showForm.value = true
  }

  function startEdit(c: NotificationChannel) {
    editingId.value = c.id
    type.value = c.type
    name.value = c.name
    secret.value = c.type === 'webhook' ? (c.config.secret ?? '') : ''
    // Set before `value` so the watcher above (which reacts to `value`
    // changing) sees the matching consentedPhone already in place and doesn't
    // mistake this initial population for the user editing the number.
    consentAt.value = c.type === 'sms' ? (c.config.consent_at ?? '') : ''
    consentedPhone.value = c.type === 'sms' ? (c.config.phone_number ?? '') : ''
    smsConsent.value = !!consentAt.value
    value.value = c.config[configKey[c.type]] ?? ''
    enabled.value = c.enabled
    formError.value = ''
    testSuccess.value = false
    testError.value = ''
    regenerateError.value = ''
    limitReached.value = false
    showForm.value = true
  }

  async function test() {
    testing.value = true
    testError.value = ''
    testSuccess.value = false
    limitReached.value = false
    try {
      await notificationChannelsApi.test({
        type: type.value,
        config: buildChannelConfig(type.value, value.value, smsConsent.value),
      })
      testSuccess.value = true
    } catch (e: unknown) {
      if (e instanceof ApiError && e.code === 'plan_limit_reached') {
        limitReached.value = true
        testError.value = e.message
      } else {
        testError.value = e instanceof Error ? e.message : 'Failed to send test message'
      }
    } finally {
      testing.value = false
    }
  }

  async function regenerateSecret() {
    if (!editingId.value) return
    regeneratingSecret.value = true
    regenerateError.value = ''
    try {
      const updated = await notificationChannelsApi.regenerateWebhookSecret(editingId.value)
      secret.value = updated.config.secret ?? ''
      await refetch()
    } catch (e: unknown) {
      regenerateError.value = e instanceof Error ? e.message : 'Failed to regenerate secret'
    } finally {
      regeneratingSecret.value = false
    }
  }

  async function persistChannel() {
    const config = buildChannelConfig(type.value, value.value, smsConsent.value)
    if (editingId.value) {
      await notificationChannelsApi.update(editingId.value, {
        type: type.value,
        name: name.value.trim(),
        config,
        enabled: enabled.value,
      })
    } else {
      await notificationChannelsApi.create({
        type: type.value,
        name: name.value.trim(),
        config,
      })
    }
  }

  async function save() {
    formError.value = ''
    limitReached.value = false
    const validationError = validateChannelSaveInput(
      type.value,
      name.value,
      value.value,
      smsConsent.value,
    )
    if (validationError) {
      formError.value = validationError
      return
    }
    saving.value = true
    try {
      await persistChannel()
      await refetch()
      cancelForm()
    } catch (e: unknown) {
      if (e instanceof ApiError && e.code === 'plan_limit_reached') {
        limitReached.value = true
        formError.value = e.message
      } else {
        formError.value = e instanceof Error ? e.message : 'Failed to save channel'
      }
    } finally {
      saving.value = false
    }
  }

  function cancelForm() {
    showForm.value = false
    editingId.value = null
  }

  return {
    showForm,
    editingId,
    type,
    name,
    value,
    enabled,
    secret,
    smsConsent,
    consentAt,
    smsConsentOnFile,
    hobbyPlanNoSms,
    channelTypeOptions,
    saving,
    testing,
    regeneratingSecret,
    regenerateError,
    formError,
    testSuccess,
    testError,
    limitReached,
    startAdd,
    startEdit,
    cancelForm,
    regenerateSecret,
    test,
    save,
  }
}
