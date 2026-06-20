<script setup lang="ts">
import { ref, watch } from 'vue'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import { settingsApi } from '@/api/settings'
import { suggestionsApi } from '@/api/suggestions'
import { useAuthStore } from '@/stores/auth'
import { useTheme } from '@/lib/theme'
import { useSettings } from '@/composables/useSettings'

const auth = useAuthStore()
const { theme, setTheme } = useTheme()

const suggestionText = ref('')
const submittingSuggestion = ref(false)
const suggestionError = ref('')
const suggestionSent = ref(false)

async function submitSuggestion() {
  submittingSuggestion.value = true
  suggestionError.value = ''
  try {
    await suggestionsApi.submit(suggestionText.value)
    suggestionSent.value = true
    suggestionText.value = ''
  } catch (e: unknown) {
    suggestionError.value = e instanceof Error ? e.message : 'Failed to send'
  } finally {
    submittingSuggestion.value = false
  }
}

function formatDate(iso: string | null): string {
  if (!iso) return ''
  return new Date(iso).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

const chatId = ref('')
const savedChatId = ref<string | null>(null)
const saving = ref(false)
const testing = ref(false)
const saveError = ref('')
const testError = ref('')
const testSuccess = ref(false)

async function save() {
  saving.value = true
  saveError.value = ''
  try {
    const s = await settingsApi.saveTelegram(chatId.value)
    savedChatId.value = s.telegramChatId
  } catch (e: unknown) {
    saveError.value = e instanceof Error ? e.message : 'Failed to save'
  } finally {
    saving.value = false
  }
}

async function test() {
  testing.value = true
  testError.value = ''
  testSuccess.value = false
  try {
    await settingsApi.testTelegram(chatId.value)
    testSuccess.value = true
  } catch (e: unknown) {
    testError.value = e instanceof Error ? e.message : 'Failed to send test message'
  } finally {
    testing.value = false
  }
}

const alertEmail = ref('')
const savedAlertEmail = ref<string | null>(null)
const emailAlertsEnabled = ref(false)
const savingEmail = ref(false)
const testingEmail = ref(false)
const togglingEmail = ref(false)
const saveEmailError = ref('')
const testEmailError = ref('')
const testEmailSuccess = ref(false)
const toggleEmailError = ref('')

async function saveEmail() {
  savingEmail.value = true
  saveEmailError.value = ''
  try {
    const s = await settingsApi.saveEmail(alertEmail.value)
    savedAlertEmail.value = s.alertEmail
  } catch (e: unknown) {
    saveEmailError.value = e instanceof Error ? e.message : 'Failed to save'
  } finally {
    savingEmail.value = false
  }
}

async function testEmail() {
  testingEmail.value = true
  testEmailError.value = ''
  testEmailSuccess.value = false
  try {
    await settingsApi.testEmail(alertEmail.value)
    testEmailSuccess.value = true
  } catch (e: unknown) {
    testEmailError.value = e instanceof Error ? e.message : 'Failed to send test email'
  } finally {
    testingEmail.value = false
  }
}

async function toggleEmailAlerts() {
  togglingEmail.value = true
  toggleEmailError.value = ''
  const desired = emailAlertsEnabled.value
  try {
    const s = await settingsApi.setEmailAlertsEnabled(desired)
    emailAlertsEnabled.value = s.emailAlertsEnabled
  } catch (e: unknown) {
    emailAlertsEnabled.value = !desired
    toggleEmailError.value = e instanceof Error ? e.message : 'Failed to update'
  } finally {
    togglingEmail.value = false
  }
}

const { data: settings } = useSettings()
watch(
  settings,
  (s) => {
    if (!s) return
    savedChatId.value = s.telegramChatId
    chatId.value = s.telegramChatId ?? ''
    savedAlertEmail.value = s.alertEmail
    alertEmail.value = s.alertEmail ?? ''
    emailAlertsEnabled.value = s.emailAlertsEnabled
  },
  { immediate: true },
)
// Load failures stay silent, same as the original onMounted's empty catch —
// not critical, the form just starts blank/unconfigured.
</script>

<template>
  <AppLayout>
    <div class="p-8 max-w-2xl mx-auto">
      <h1 class="text-2xl font-semibold mb-6" style="color: var(--text-strong)">Settings</h1>

      <!-- Appearance -->
      <div
        class="rounded-xl border p-6 mb-6"
        style="background-color: var(--surface); border-color: var(--border)"
      >
        <h2 class="font-medium mb-1" style="color: var(--text-strong)">Appearance</h2>
        <p class="text-sm mb-5" style="color: var(--text-muted)">
          Choose how checkmeup looks on this device.
        </p>

        <div class="inline-flex rounded-md border p-1" style="border-color: var(--border)">
          <button
            type="button"
            class="px-3 py-1.5 rounded text-sm transition-colors hover:cursor-pointer"
            :style="
              theme === 'dark'
                ? 'background-color: var(--surface-raised); color: var(--text-strong)'
                : 'color: var(--text-muted)'
            "
            @click="setTheme('dark')"
          >
            Dark
          </button>
          <button
            type="button"
            class="px-3 py-1.5 rounded text-sm transition-colors hover:cursor-pointer"
            :style="
              theme === 'light'
                ? 'background-color: var(--surface-raised); color: var(--text-strong)'
                : 'color: var(--text-muted)'
            "
            @click="setTheme('light')"
          >
            Light
          </button>
        </div>
      </div>

      <!-- Telegram -->
      <div
        class="rounded-xl border p-6"
        style="background-color: var(--surface); border-color: var(--border)"
      >
        <h2 class="font-medium mb-1" style="color: var(--text-strong)">Telegram alerts</h2>
        <p class="text-sm mb-5" style="color: var(--text-muted)">
          Connect a Telegram chat to receive down and recovery alerts.
        </p>

        <ol class="text-sm space-y-2 mb-5 list-decimal list-inside" style="color: var(--text-dim)">
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
          <li>Click <strong>Save</strong></li>
        </ol>

        <div class="space-y-4">
          <div>
            <Label for="chat-id">Chat ID</Label>
            <Input id="chat-id" v-model="chatId" placeholder="-1001234567890" class="mt-1" />
          </div>

          <div class="flex items-center gap-3">
            <Button variant="secondary" :disabled="!chatId || testing" @click="test">
              {{ testing ? 'Sending…' : 'Send test message' }}
            </Button>
            <Button :disabled="!chatId || saving" @click="save">
              {{ saving ? 'Saving…' : 'Save' }}
            </Button>
          </div>

          <p v-if="testSuccess" class="text-sm" style="color: var(--status-up)">
            Test message sent! Check your Telegram chat.
          </p>
          <p v-if="testError" class="text-sm" style="color: var(--status-down)">
            {{ testError }}
          </p>
          <p v-if="saveError" class="text-sm" style="color: var(--status-down)">
            {{ saveError }}
          </p>
          <p v-if="savedChatId && !saveError" class="text-xs" style="color: var(--text-muted)">
            Currently connected: {{ savedChatId }}
          </p>
        </div>
      </div>

      <!-- Email -->
      <div class="rounded-xl border p-6 mt-6" style="background-color: var(--surface); border-color: var(--border)">
        <h2 class="font-medium mb-1" style="color: var(--text-strong)">Email alerts</h2>
        <p class="text-sm mb-5" style="color: var(--text-muted)">
          Receive down and recovery alerts by email, independent of Telegram.
        </p>

        <div class="space-y-4">
          <div>
            <Label for="alert-email">Alert email address</Label>
            <Input
              id="alert-email"
              v-model="alertEmail"
              type="email"
              placeholder="alerts@yourteam.com"
              class="mt-1"
            />
          </div>

          <div class="flex items-center gap-3">
            <Button
              variant="secondary"
              :disabled="!alertEmail || testingEmail"
              @click="testEmail"
            >
              {{ testingEmail ? 'Sending…' : 'Send test email' }}
            </Button>
            <Button
              :disabled="!alertEmail || savingEmail"
              @click="saveEmail"
            >
              {{ savingEmail ? 'Saving…' : 'Save' }}
            </Button>
          </div>

          <p v-if="testEmailSuccess" class="text-sm" style="color: var(--status-up)">
            Test email sent! Check your inbox.
          </p>
          <p v-if="testEmailError" class="text-sm" style="color: var(--status-down)">
            {{ testEmailError }}
          </p>
          <p v-if="saveEmailError" class="text-sm" style="color: var(--status-down)">
            {{ saveEmailError }}
          </p>
          <p v-if="savedAlertEmail && !saveEmailError" class="text-xs" style="color: var(--text-muted)">
            Currently set: {{ savedAlertEmail }}
          </p>

          <div class="flex items-center gap-3 pt-2 border-t" style="border-color: var(--border)">
            <input
              id="email-alerts-enabled"
              v-model="emailAlertsEnabled"
              type="checkbox"
              class="rounded"
              :disabled="togglingEmail"
              @change="toggleEmailAlerts"
            />
            <Label for="email-alerts-enabled" class="cursor-pointer">Enable email alerts</Label>
          </div>
          <p v-if="toggleEmailError" class="text-sm" style="color: var(--status-down)">
            {{ toggleEmailError }}
          </p>
        </div>
      </div>

      <!-- Legal -->
      <div
        class="rounded-xl border p-6 mt-6"
        style="background-color: var(--surface); border-color: var(--border)"
      >
        <h2 class="font-medium mb-1" style="color: var(--text-strong)">Terms and Privacy</h2>
        <p v-if="auth.user?.termsAcceptedAt" class="text-sm" style="color: var(--text-muted)">
          You accepted the
          <RouterLink to="/terms" class="underline" style="color: var(--color-green-500)"
            >Terms of Service</RouterLink
          >
          and
          <RouterLink to="/privacy" class="underline" style="color: var(--color-green-500)"
            >Privacy Policy</RouterLink
          >
          (version {{ auth.user?.termsVersion }}) on {{ formatDate(auth.user.termsAcceptedAt) }}.
        </p>
      </div>

      <!-- Suggest a feature -->
      <div
        class="rounded-xl border p-6 mt-6"
        style="background-color: var(--surface); border-color: var(--border)"
      >
        <h2 class="font-medium mb-1" style="color: var(--text-strong)">Suggest a feature</h2>
        <p class="text-sm mb-5" style="color: var(--text-muted)">
          There's no support ticket queue — this goes straight to the founder.
        </p>

        <div class="space-y-4">
          <div>
            <Label for="suggestion">Your suggestion</Label>
            <textarea
              id="suggestion"
              v-model="suggestionText"
              rows="4"
              placeholder="What should checkmeup do better?"
              class="mt-1 flex w-full rounded-md border px-3 py-2 text-sm transition-colors focus:outline-none focus:ring-2 disabled:cursor-not-allowed disabled:opacity-50"
              style="
                border-color: var(--border);
                background-color: var(--surface);
                color: var(--text);
              "
            ></textarea>
          </div>

          <div class="flex items-center justify-between gap-3">
            <p class="text-xs" style="color: var(--text-muted)">Sent as {{ auth.user?.email }}</p>
            <Button
              :disabled="!suggestionText.trim() || submittingSuggestion"
              @click="submitSuggestion"
            >
              {{ submittingSuggestion ? 'Sending…' : 'Send suggestion' }}
            </Button>
          </div>

          <p v-if="suggestionSent" class="text-sm" style="color: var(--status-up)">
            Thanks — this reaches an engineer directly.
          </p>
          <p v-if="suggestionError" class="text-sm" style="color: var(--status-down)">
            {{ suggestionError }}
          </p>
        </div>
      </div>
    </div>
  </AppLayout>
</template>
