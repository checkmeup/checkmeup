<script setup lang="ts">
import { ref } from 'vue'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import Label from '@/components/ui/Label.vue'
import NotificationChannelsCard from '@/components/NotificationChannelsCard.vue'
import ApiKeysCard from '@/components/ApiKeysCard.vue'
import LabeledSection from '@/components/LabeledSection.vue'
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
    <div class="mx-auto max-w-[1000px] px-4 pt-8 pb-20 md:px-6">
      <h1 class="mb-8 text-[30px] leading-[1.1] font-semibold tracking-[-0.035em] text-[var(--text-strong)] md:text-[34px]">Settings</h1>

      <!-- Appearance -->
      <LabeledSection title="Appearance">
        <template #description>Choose how Checkmeup looks on this device.</template>

        <div class="inline-flex gap-0.5 rounded-[10px] bg-[var(--surface-raised)] p-[3px]">
          <button
            type="button"
            class="min-h-[34px] rounded-lg px-4 text-sm font-medium transition-colors hover:cursor-pointer"
            :class="
              theme === 'dark'
                ? 'bg-[var(--bg)] text-[var(--text-strong)] shadow-[0_1px_2px_rgb(0_0_0/12%)]'
                : 'text-[var(--text-dim)] hover:text-[var(--text)]'
            "
            @click="setTheme('dark')"
          >
            Dark
          </button>
          <button
            type="button"
            class="min-h-[34px] rounded-lg px-4 text-sm font-medium transition-colors hover:cursor-pointer"
            :class="
              theme === 'light'
                ? 'bg-[var(--bg)] text-[var(--text-strong)] shadow-[0_1px_2px_rgb(0_0_0/12%)]'
                : 'text-[var(--text-dim)] hover:text-[var(--text)]'
            "
            @click="setTheme('light')"
          >
            Light
          </button>
        </div>
      </LabeledSection>

      <!-- Notification channels -->
      <NotificationChannelsCard />

      <!-- API keys -->
      <ApiKeysCard />

      <!-- Legal -->
      <LabeledSection title="Terms and Privacy">
        <p v-if="auth.user?.termsAcceptedAt" class="text-sm text-[var(--text-dim)]">
          You accepted the
          <RouterLink to="/terms" class="text-[var(--accent)] underline"
            >Terms of Service</RouterLink
          >
          and
          <RouterLink to="/privacy" class="text-[var(--accent)] underline"
            >Privacy Policy</RouterLink
          >
          (version {{ auth.user?.termsVersion }}) on {{ formatDate(auth.user.termsAcceptedAt) }}.
        </p>
      </LabeledSection>

      <!-- Suggest a feature -->
      <LabeledSection title="Suggest a feature">
        <template #description>
          There's no support ticket queue — this goes straight to the founder.
        </template>

        <div class="space-y-4">
          <div>
            <Label for="suggestion">Your suggestion</Label>
            <textarea
              id="suggestion"
              v-model="suggestionText"
              rows="4"
              placeholder="What should Checkmeup do better?"
              class="mt-1 flex w-full rounded-lg border px-3 py-2 text-sm transition-colors disabled:cursor-not-allowed disabled:opacity-50"
            ></textarea>
          </div>

          <div class="flex items-center justify-between gap-3">
            <p class="text-xs text-[var(--text-muted)]">Sent as {{ auth.user?.email }}</p>
            <Button
              :disabled="!suggestionText.trim() || submittingSuggestion"
              @click="submitSuggestion"
            >
              {{ submittingSuggestion ? 'Sending…' : 'Send suggestion' }}
            </Button>
          </div>

          <p v-if="suggestionSent" class="text-sm text-[var(--status-up)]">
            Thanks — this reaches an engineer directly.
          </p>
          <p v-if="suggestionError" class="text-sm text-[var(--status-down)]">
            {{ suggestionError }}
          </p>
        </div>
      </LabeledSection>
    </div>
  </AppLayout>
</template>
