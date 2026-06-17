import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface User {
  id: string
  email: string
  orgId: string
  termsVersion: string | null
  termsAcceptedAt: string | null
  needsTermsAcceptance: boolean
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
      // Use plain fetch to avoid the 401 interceptor triggering a redirect
      // during the initial auth check — a 401 here just means "not logged in".
      const res = await fetch('/api/v1/me', { credentials: 'include' })
      if (!res.ok) throw new Error('unauthenticated')
      user.value = (await res.json()) as User
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
