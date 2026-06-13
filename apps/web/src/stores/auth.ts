import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/api/client'

export interface User {
  id: string
  email: string
  orgId: string
}

type AuthStatus = 'idle' | 'loading' | 'authenticated' | 'unauthenticated'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const status = ref<AuthStatus>('idle')

  const isAuthenticated = computed(() => status.value === 'authenticated')

  async function init() {
    if (status.value !== 'idle') return
    status.value = 'loading'
    try {
      user.value = await api.get<User>('/api/v1/me')
      status.value = 'authenticated'
    } catch {
      user.value = null
      status.value = 'unauthenticated'
    }
  }

  function setUser(u: User) {
    user.value = u
    status.value = 'authenticated'
  }

  function clear() {
    user.value = null
    status.value = 'unauthenticated'
  }

  return { user, status, isAuthenticated, init, setUser, clear }
})
