<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import AuthLayout from '@/layouts/AuthLayout.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import { api, ApiError } from '@/api/client'
import { trackSignUp } from '@/lib/analytics'
import { useAuthStore } from '@/stores/auth'
import type { User } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

const email = ref('')
const password = ref('')
const acceptedTerms = ref(false)
const error = ref('')
const loading = ref(false)

function validate(): string | null {
  if (password.value.length < 8) return 'Password must be at least 8 characters.'
  if (!acceptedTerms.value) return 'You must accept the Terms of Service and Privacy Policy.'
  return null
}

async function submit() {
  error.value = validate() ?? ''
  if (error.value) return

  loading.value = true
  try {
    const user = await api.post<User>('/api/v1/auth/sign-up', {
      email: email.value,
      password: password.value,
      acceptedTerms: acceptedTerms.value,
    })
    auth.setUser(user)
    trackSignUp()
    router.push({ name: 'dashboard' })
  } catch (err) {
    if (err instanceof ApiError && err.status === 409) {
      error.value = 'An account with this email already exists.'
    } else {
      error.value = 'Something went wrong. Please try again.'
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <AuthLayout>
    <div class="space-y-1">
      <h1 class="text-[26px] font-semibold tracking-[-0.03em] text-[var(--text-strong)]">Create account</h1>
      <p class="text-sm text-[var(--text-dim)]">Start monitoring your services for free.</p>
    </div>

    <form class="space-y-4" @submit.prevent="submit">
      <div class="space-y-1.5">
        <Label for="email">Email</Label>
        <Input
          id="email"
          v-model="email"
          type="email"
          placeholder="you@example.com"
          autocomplete="email"
          required
        />
      </div>

      <div class="space-y-1.5">
        <Label for="password">Password</Label>
        <Input
          id="password"
          v-model="password"
          type="password"
          placeholder="8+ characters"
          autocomplete="new-password"
          required
        />
      </div>

      <label class="flex items-start gap-2 text-sm text-[var(--text-dim)]">
        <input v-model="acceptedTerms" type="checkbox" class="mt-0.5" />
        <span>
          I agree to the
          <RouterLink to="/terms" target="_blank" class="underline text-[var(--accent)]"
            >Terms of Service</RouterLink
          >
          and
          <RouterLink to="/privacy" target="_blank" class="underline text-[var(--accent)]"
            >Privacy Policy</RouterLink
          >.
        </span>
      </label>

      <p v-if="error" class="text-sm text-[var(--status-down)]">{{ error }}</p>

      <Button type="submit" class="w-full" :disabled="loading || !acceptedTerms">
        {{ loading ? 'Creating account…' : 'Create account' }}
      </Button>
    </form>

    <p class="text-center text-sm text-[var(--text-muted)]">
      Already have an account?
      <RouterLink to="/sign-in" class="font-medium text-[var(--accent)]">
        Sign in
      </RouterLink>
    </p>

    <p class="text-center text-xs text-[var(--text-muted)]">
      Questions?
      <RouterLink class="text-[var(--accent)]" to="/faq#getting-started">
        See the FAQ
      </RouterLink>
    </p>
  </AuthLayout>
</template>
