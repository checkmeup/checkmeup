<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import AuthLayout from '@/layouts/AuthLayout.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import { api, ApiError } from '@/api/client'

const router = useRouter()
const route = useRoute()

const token = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

onMounted(() => {
  token.value = String(route.query.token ?? '')
  if (!token.value) {
    error.value = 'Invalid or missing reset link.'
  }
})

async function submit() {
  if (password.value.length < 8) {
    error.value = 'Password must be at least 8 characters.'
    return
  }
  error.value = ''
  loading.value = true
  try {
    await api.post('/api/v1/auth/reset-password', {
      token: token.value,
      password: password.value,
    })
    router.push({ name: 'sign-in' })
  } catch (err) {
    if (err instanceof ApiError && err.status === 400) {
      error.value = 'This reset link is invalid or has expired.'
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
      <h1 class="text-lg font-semibold" style="color: var(--text-strong)">Set new password</h1>
      <p class="text-sm" style="color: var(--text-muted)">Choose a new password for your account.</p>
    </div>

    <form class="space-y-4" @submit.prevent="submit">
      <div class="space-y-1.5">
        <Label for="password">New password</Label>
        <Input
          id="password"
          v-model="password"
          type="password"
          placeholder="8+ characters"
          autocomplete="new-password"
          required
          :disabled="!token"
        />
      </div>

      <p v-if="error" class="text-sm" style="color: var(--status-down)">{{ error }}</p>

      <Button type="submit" class="w-full" :disabled="loading || !token">
        {{ loading ? 'Saving…' : 'Set new password' }}
      </Button>
    </form>
  </AuthLayout>
</template>
