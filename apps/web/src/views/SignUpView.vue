<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import AuthLayout from '@/layouts/AuthLayout.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import { api, ApiError } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import type { User } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

function validate(): string | null {
  if (password.value.length < 8) return 'Password must be at least 8 characters.'
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
    })
    auth.setUser(user)
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
      <h1 class="text-lg font-semibold" style="color: var(--text-strong)">Create account</h1>
      <p class="text-sm" style="color: var(--text-muted)">Start monitoring your services for free.</p>
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

      <p v-if="error" class="text-sm" style="color: var(--status-down)">{{ error }}</p>

      <Button type="submit" class="w-full" :disabled="loading">
        {{ loading ? 'Creating account…' : 'Create account' }}
      </Button>
    </form>

    <p class="text-center text-sm" style="color: var(--text-muted)">
      Already have an account?
      <RouterLink to="/sign-in" class="font-medium" style="color: var(--color-green-500)">
        Sign in
      </RouterLink>
    </p>
  </AuthLayout>
</template>
