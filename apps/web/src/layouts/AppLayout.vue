<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/api/client'
import logoIcon from '@/assets/logo-icon.svg'

const router = useRouter()
const auth = useAuthStore()

async function signOut() {
  try {
    await api.post('/api/v1/auth/sign-out', {})
  } finally {
    auth.clear()
    router.push({ name: 'sign-in' })
  }
}
</script>

<template>
  <div class="flex h-screen" style="background-color: var(--bg)">
    <!-- Sidebar -->
    <aside class="w-56 flex-shrink-0 flex flex-col border-r"
           style="background-color: var(--surface); border-color: var(--border)">
      <!-- Logo -->
      <div class="flex items-center gap-2 px-4 py-5 border-b" style="border-color: var(--border)">
        <img :src="logoIcon" alt="" class="h-6 w-6" />
        <span class="text-sm font-semibold" style="color: var(--text-strong)">checkmeup</span>
      </div>

      <!-- Nav -->
      <nav class="flex-1 px-2 py-4 space-y-1">
        <RouterLink
          to="/dashboard"
          class="flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors"
          style="color: var(--text-dim)"
          active-class="!text-white"
          :style="{ '--active-bg': 'var(--surface-raised)' }"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/></svg>
          Dashboard
        </RouterLink>
      </nav>

      <!-- User -->
      <div class="px-4 py-3 border-t" style="border-color: var(--border)">
        <div class="text-xs truncate mb-2" style="color: var(--text-muted)">
          {{ auth.user?.email }}
        </div>
        <button
          class="text-xs w-full text-left transition-colors"
          style="color: var(--text-muted)"
          @click="signOut"
        >
          Sign out
        </button>
      </div>
    </aside>

    <!-- Main -->
    <main class="flex-1 overflow-y-auto">
      <slot />
    </main>
  </div>
</template>
