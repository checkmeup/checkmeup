<script setup lang="ts">
import { ref, onMounted } from 'vue'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import { settingsApi } from '@/api/settings'

const chatId = ref('')
const savedChatId = ref<string | null>(null)
const saving = ref(false)
const testing = ref(false)
const saveError = ref('')
const testError = ref('')
const testSuccess = ref(false)

onMounted(async () => {
  try {
    const s = await settingsApi.get()
    savedChatId.value = s.telegramChatId
    chatId.value = s.telegramChatId ?? ''
  } catch {
    // not critical
  }
})

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
</script>

<template>
  <AppLayout>
    <div class="p-8 max-w-2xl mx-auto">
      <h1 class="text-2xl font-semibold mb-6" style="color: var(--text-strong)">Settings</h1>

      <!-- Telegram -->
      <div class="rounded-xl border p-6" style="background-color: var(--surface); border-color: var(--border)">
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
            >@checkmeupnet_bot</a>
            in Telegram and send <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)">/start</code>
            — the bot will reply with your Chat ID
          </li>
          <li>Paste the Chat ID below and click <strong>Send test message</strong> to verify</li>
          <li>Click <strong>Save</strong></li>
        </ol>
        <p class="text-xs mb-5 p-3 rounded-lg" style="background-color: var(--surface-raised); color: var(--text-muted)">
          <strong style="color: var(--text-dim)">Note:</strong> The bot only works once it's deployed with a webhook registered.
          In local dev, get your Chat ID from
          <a href="https://t.me/userinfobot" target="_blank" rel="noopener" class="underline" style="color: var(--color-green-500)">@userinfobot</a>
          instead.
        </p>

        <div class="space-y-4">
          <div>
            <Label for="chat-id">Chat ID</Label>
            <Input
              id="chat-id"
              v-model="chatId"
              placeholder="-1001234567890"
              class="mt-1"
            />
          </div>

          <div class="flex items-center gap-3">
            <Button
              variant="secondary"
              :disabled="!chatId || testing"
              @click="test"
            >
              {{ testing ? 'Sending…' : 'Send test message' }}
            </Button>
            <Button
              :disabled="!chatId || saving"
              @click="save"
            >
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
    </div>
  </AppLayout>
</template>
