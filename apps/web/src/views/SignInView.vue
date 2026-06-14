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

async function submit() {
  error.value = ''
  loading.value = true
  try {
    const user = await api.post<User>('/api/v1/auth/sign-in', {
      email: email.value,
      password: password.value,
    })
    auth.setUser(user)
    router.push({ name: 'dashboard' })
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      error.value = 'Incorrect email or password.'
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
      <h1 class="text-lg font-semibold" style="color: var(--text-strong)">Sign in</h1>
      <p class="text-sm" style="color: var(--text-muted)">Enter your credentials to continue.</p>
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
        <div class="flex items-center justify-between">
          <Label for="password">Password</Label>
          <RouterLink
            to="/forgot-password"
            class="text-xs"
            style="color: var(--text-muted)"
          >
            Forgot password?
          </RouterLink>
        </div>
        <Input
          id="password"
          v-model="password"
          type="password"
          placeholder="••••••••"
          autocomplete="current-password"
          required
        />
      </div>

      <p v-if="error" class="text-sm" style="color: var(--status-down)">{{ error }}</p>

      <Button type="submit" class="w-full" :disabled="loading">
        {{ loading ? 'Signing in…' : 'Sign in' }}
      </Button>
    </form>

    <p class="text-center text-sm" style="color: var(--text-muted)">
      No account?
      <RouterLink to="/sign-up" class="font-medium" style="color: var(--color-green-500)">
        Sign up
      </RouterLink>
    </p>
  </AuthLayout>
</template>
