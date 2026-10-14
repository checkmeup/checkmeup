<script setup lang="ts">
import { ref } from 'vue'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import Label from '@/components/ui/Label.vue'
import NotificationChannelsCard from '@/components/NotificationChannelsCard.vue'
import ApiKeysCard from '@/components/ApiKeysCard.vue'
import { suggestionsApi } from '@/api/suggestions'
import { useAuthStore } from '@/stores/auth'
import { useTheme } from '@/lib/theme'

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
</script>

<template>
  <AppLayout>
    <div class="mx-auto max-w-[720px] px-4 pt-8 pb-20 md:px-6">
      <h1 class="mb-8 text-[30px] leading-[1.1] font-semibold tracking-[-0.035em] text-[var(--text-strong)] md:text-[34px]">Settings</h1>

      <!-- Appearance -->
      <div
        class="rounded-[14px] border border-[var(--border)] p-6 mb-6"
      >
        <h2 class="mb-1 text-lg font-semibold tracking-[-0.015em] text-[var(--text-strong)]">Appearance</h2>
        <p class="text-sm mb-5" style="color: var(--text-muted)">
          Choose how Checkmeup looks on this device.
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

      <!-- Notification channels -->
      <NotificationChannelsCard class="mt-6" />

      <!-- API keys -->
      <ApiKeysCard class="mt-6" />

      <!-- Legal -->
      <div
        class="rounded-[14px] border border-[var(--border)] p-6 mt-6"
      >
        <h2 class="mb-1 text-lg font-semibold tracking-[-0.015em] text-[var(--text-strong)]">Terms and Privacy</h2>
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
        class="rounded-[14px] border border-[var(--border)] p-6 mt-6"
      >
        <h2 class="mb-1 text-lg font-semibold tracking-[-0.015em] text-[var(--text-strong)]">Suggest a feature</h2>
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
              placeholder="What should Checkmeup do better?"
              class="mt-1 flex w-full rounded-lg border px-3 py-2 text-sm transition-colors focus:outline-none focus:ring-2 disabled:cursor-not-allowed disabled:opacity-50"
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
