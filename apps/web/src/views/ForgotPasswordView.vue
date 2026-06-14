<script setup lang="ts">
import { ref } from 'vue'
import AuthLayout from '@/layouts/AuthLayout.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import { api } from '@/api/client'

const email = ref('')
const loading = ref(false)
const submitted = ref(false)

async function submit() {
  loading.value = true
  try {
    await api.post('/api/v1/auth/forgot-password', { email: email.value })
  } finally {
    loading.value = false
    submitted.value = true
  }
}
</script>

<template>
  <AuthLayout>
    <div class="space-y-1">
      <h1 class="text-lg font-semibold" style="color: var(--text-strong)">Reset password</h1>
      <p class="text-sm" style="color: var(--text-muted)">
        Enter your email and we'll send you a reset link.
      </p>
    </div>

    <template v-if="!submitted">
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

        <Button type="submit" class="w-full" :disabled="loading">
          {{ loading ? 'Sending…' : 'Send reset link' }}
        </Button>
      </form>
    </template>

    <template v-else>
      <p class="text-sm" style="color: var(--text-dim)">
        If an account exists for <strong>{{ email }}</strong>, you'll receive a reset link shortly.
      </p>
    </template>

    <p class="text-center text-sm" style="color: var(--text-muted)">
      <RouterLink to="/sign-in" class="font-medium" style="color: var(--color-green-500)">
        Back to sign in
      </RouterLink>
    </p>
  </AuthLayout>
</template>
