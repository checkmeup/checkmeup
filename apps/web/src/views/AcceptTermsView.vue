<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import AuthLayout from '@/layouts/AuthLayout.vue'
import Button from '@/components/ui/Button.vue'
import { api } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import type { User } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

const agreed = ref(false)
const submitting = ref(false)
const error = ref('')

const isReaccept = auth.user?.termsVersion != null

async function accept() {
  if (!agreed.value) return
  submitting.value = true
  error.value = ''
  try {
    const user = await api.post<User>('/api/v1/auth/accept-terms', {})
    auth.setUser(user)
    router.push({ name: 'dashboard' })
  } catch {
    error.value = 'Something went wrong. Please try again.'
  } finally {
    submitting.value = false
  }
}

async function signOut() {
  try {
    await api.post('/api/v1/auth/sign-out', {})
  } finally {
    auth.clear()
    router.push({ name: 'home' })
  }
}
</script>

<template>
  <AuthLayout>
    <div class="space-y-1">
      <h1 class="text-lg font-semibold" style="color: var(--text-strong)">
        {{ isReaccept ? "We've updated our Terms and Privacy Policy" : 'One more thing' }}
      </h1>
      <p class="text-sm" style="color: var(--text-muted)">
        Please review and accept before continuing.
      </p>
    </div>

    <label class="flex items-start gap-2 text-sm" style="color: var(--text-dim)">
      <input v-model="agreed" type="checkbox" class="mt-0.5" />
      <span>
        I agree to the
        <RouterLink to="/terms" target="_blank" class="underline" style="color: var(--color-green-500)"
          >Terms of Service</RouterLink
        >
        and
        <RouterLink to="/privacy" target="_blank" class="underline" style="color: var(--color-green-500)"
          >Privacy Policy</RouterLink
        >.
      </span>
    </label>

    <p v-if="error" class="text-sm" style="color: var(--status-down)">{{ error }}</p>

    <Button class="w-full" :disabled="!agreed || submitting" @click="accept">
      {{ submitting ? 'Saving…' : 'Accept and continue' }}
    </Button>

    <button
      class="w-full text-center text-sm transition-colors hover:cursor-pointer"
      style="color: var(--text-muted)"
      @click="signOut"
    >
      Sign out instead
    </button>
  </AuthLayout>
</template>
