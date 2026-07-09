<script setup lang="ts">
import type { UnwrapNestedRefs } from 'vue'
import { RouterLink } from 'vue-router'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import UpgradePrompt from '@/components/UpgradePrompt.vue'
import {
  valueLabel,
  valuePlaceholder,
  useNotificationChannelForm,
} from '@/composables/useNotificationChannelForm'

defineProps<{
  form: UnwrapNestedRefs<ReturnType<typeof useNotificationChannelForm>>
}>()
</script>

<template>
  <div
    class="rounded-lg border p-4 space-y-4"
    style="border-color: var(--border); background-color: var(--surface-raised)"
  >
    <div>
      <Label>Type</Label>
      <div
        class="mt-1 grid gap-2"
        style="grid-template-columns: repeat(auto-fill, minmax(104px, 1fr))"
      >
        <button
          v-for="t in form.channelTypeOptions"
          :key="t.value"
          type="button"
          data-testid="channel-type-option"
          :aria-pressed="form.type === t.value"
          class="flex flex-col items-center gap-1.5 rounded-lg border px-1.5 py-3 text-xs font-medium transition-colors disabled:opacity-50 disabled:pointer-events-none"
          :style="{
            borderColor: form.type === t.value ? 'var(--accent)' : 'var(--border)',
            backgroundColor: form.type === t.value ? 'var(--accent-wash)' : 'var(--surface)',
            color: form.type === t.value ? 'var(--accent)' : 'var(--text-dim)',
          }"
          :disabled="!!form.editingId"
          @click="form.type = t.value"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path :d="t.iconPath" />
          </svg>
          {{ t.label }}
        </button>
      </div>
      <p v-if="form.hobbyPlanNoSms" class="mt-1 text-xs" style="color: var(--text-muted)">
        SMS alerts require a paid plan —
        <RouterLink to="/billing" class="underline" style="color: var(--color-green-500)"
          >view plans</RouterLink
        >.
      </p>
    </div>
    <ol
      v-if="form.type === 'telegram'"
      class="text-sm space-y-2 list-decimal list-inside"
      style="color: var(--text-dim)"
    >
      <li>
        Open
        <a
          href="https://t.me/checkmeupnet_bot"
          target="_blank"
          rel="noopener"
          class="underline"
          style="color: var(--color-green-500)"
          >@checkmeupnet_bot</a
        >
        in Telegram and send
        <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
          >/start</code
        >
        — the bot will reply with your Chat ID
      </li>
      <li>Paste the Chat ID below and click <strong>Send test message</strong> to verify</li>
    </ol>
    <p v-else-if="form.type === 'webhook'" class="text-sm" style="color: var(--text-dim)">
      checkmeup will POST a JSON payload to this URL on every down/recovery event, signed with
      <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
        >X-Checkmeup-Signature</code
      >. Must be
      <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
        >https://</code
      >.
    </p>
    <ol
      v-else-if="form.type === 'slack'"
      class="text-sm space-y-2 list-decimal list-inside"
      style="color: var(--text-dim)"
    >
      <li>
        In Slack, go to <strong>Apps → Incoming Webhooks</strong> and create a new webhook for
        your target channel
      </li>
      <li>
        Copy the <strong>Webhook URL</strong> (starts with
        <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
          >https://hooks.slack.com/services/</code
        >) and paste it below
      </li>
      <li>Click <strong>Send test message</strong> to verify the connection before saving</li>
    </ol>
    <p v-else-if="form.type === 'sms'" class="text-sm" style="color: var(--text-dim)">
      checkmeup will text down/recovery alerts to this number. Real per-message cost applies once
      Twilio account setup is complete.
    </p>
    <div>
      <Label for="channel-name">Name</Label>
      <Input id="channel-name" v-model="form.name" placeholder="e.g. Ops Telegram" class="mt-1" />
    </div>
    <div>
      <Label for="channel-value">{{ valueLabel[form.type] }}</Label>
      <Input
        id="channel-value"
        v-model="form.value"
        :type="form.type === 'email' ? 'email' : form.type === 'webhook' ? 'url' : 'text'"
        :placeholder="valuePlaceholder[form.type]"
        class="mt-1"
      />
    </div>
    <div v-if="form.type === 'sms'">
      <p v-if="form.smsConsentOnFile" class="text-xs" style="color: var(--text-muted)">
        Consent given on {{ new Date(form.consentAt).toLocaleString() }}.
      </p>
      <label v-else class="flex items-start gap-2 text-sm" style="color: var(--text-dim)">
        <input v-model="form.smsConsent" type="checkbox" class="mt-0.5" />
        <span>I agree to receive automated SMS alerts from checkmeup at this number.</span>
      </label>
    </div>
    <div v-if="form.type === 'webhook' && form.editingId" class="space-y-2">
      <Label for="channel-secret">Signing secret</Label>
      <div class="flex items-center gap-3">
        <Input
          id="channel-secret"
          :model-value="form.secret"
          disabled
          class="mt-1 font-mono text-xs"
        />
        <Button
          variant="secondary"
          size="sm"
          :disabled="form.regeneratingSecret"
          @click="form.regenerateSecret"
        >
          {{ form.regeneratingSecret ? 'Regenerating…' : 'Regenerate' }}
        </Button>
      </div>
      <p class="text-xs" style="color: var(--text-muted)">
        Verify a request by computing HMAC-SHA256 of the raw request body using this secret as the
        key, hex-encoding it, and comparing it to the
        <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
          >X-Checkmeup-Signature</code
        >
        header. Regenerating invalidates the signature for future sends only — already-delivered
        requests aren't affected.
      </p>
      <p v-if="form.regenerateError" class="text-xs" style="color: var(--status-down)">
        {{ form.regenerateError }}
      </p>
    </div>
    <p v-else-if="form.type === 'webhook'" class="text-xs" style="color: var(--text-muted)">
      A signing secret is generated automatically once you save this channel.
    </p>
    <div class="flex items-center gap-3">
      <Button
        variant="secondary"
        :disabled="!form.value.trim() || form.testing || (form.type === 'sms' && !form.smsConsent)"
        @click="form.test"
      >
        {{
          form.testing
            ? 'Sending…'
            : form.type === 'webhook'
              ? 'Send test webhook'
              : form.type === 'sms'
                ? 'Send test SMS'
                : 'Send test message'
        }}
      </Button>
      <Button
        :disabled="form.saving || (form.type === 'sms' && !form.smsConsent)"
        @click="form.save"
      >
        {{ form.saving ? 'Saving…' : form.editingId ? 'Save changes' : 'Add channel' }}
      </Button>
      <Button variant="secondary" type="button" @click="form.cancelForm">Cancel</Button>
    </div>
    <p v-if="form.testSuccess" class="text-sm" style="color: var(--status-up)">
      {{
        form.type === 'webhook'
          ? 'Test webhook sent!'
          : form.type === 'sms'
            ? 'Test SMS sent!'
            : 'Test message sent!'
      }}
    </p>
    <UpgradePrompt v-if="form.limitReached" :message="form.testError || form.formError" />
    <p v-else-if="form.testError" class="text-sm" style="color: var(--status-down)">{{ form.testError }}</p>
    <p v-else-if="form.formError" class="text-sm" style="color: var(--status-down)">{{ form.formError }}</p>
  </div>
</template>
