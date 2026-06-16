<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/api/client'
import logoDark from '@/assets/logo-dark.svg'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const sidebarOpen = ref(false)
const appVersion = import.meta.env.VITE_APP_VERSION ?? 'dev'

watch(
  () => route.path,
  () => {
    sidebarOpen.value = false
  },
)

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
    <!-- Mobile backdrop -->
    <div
      v-if="sidebarOpen"
      class="fixed inset-0 z-20 bg-black/50 md:hidden"
      @click="sidebarOpen = false"
    />

    <!-- Sidebar — fixed overlay on mobile, static column on desktop -->
    <aside
      class="fixed inset-y-0 left-0 z-30 w-56 flex-shrink-0 flex flex-col border-r transition-transform duration-200 md:static md:translate-x-0"
      :class="sidebarOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0'"
      style="background-color: var(--surface); border-color: var(--border)"
    >
      <!-- Logo -->
      <div class="flex items-center gap-2 px-4 py-5 border-b" style="border-color: var(--border)">
        <img :src="logoDark" alt="" class="h-6" />
        <div class="text-xs mt-1" style="color: var(--text-muted)">
          {{ appVersion }}
        </div>
      </div>

      <!-- Nav -->
      <nav class="flex-1 px-2 py-4 space-y-1">
        <RouterLink
          to="/dashboard"
          class="flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors"
          style="color: var(--text-dim)"
          active-class="font-medium"
        >
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <rect x="3" y="3" width="7" height="7" />
            <rect x="14" y="3" width="7" height="7" />
            <rect x="3" y="14" width="7" height="7" />
            <rect x="14" y="14" width="7" height="7" />
          </svg>
          Dashboard
        </RouterLink>

        <div
          class="pt-2 pb-1 px-3 text-xs font-medium uppercase tracking-wider"
          style="color: var(--text-muted)"
        >
          Monitors
        </div>

        <RouterLink
          to="/monitors/cron"
          class="flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors"
          style="color: var(--text-dim)"
          active-class="font-medium"
        >
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <circle cx="12" cy="12" r="10" />
            <polyline points="12 6 12 12 16 14" />
          </svg>
          Cron
        </RouterLink>

        <RouterLink
          to="/monitors/uptime"
          class="flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors"
          style="color: var(--text-dim)"
          active-class="font-medium"
        >
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
          </svg>
          Uptime
        </RouterLink>

        <RouterLink
          to="/monitors/ssl"
          class="flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors"
          style="color: var(--text-dim)"
          active-class="font-medium"
        >
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
            <path d="M7 11V7a5 5 0 0 1 10 0v4" />
          </svg>
          SSL
        </RouterLink>

        <div
          class="pt-2 pb-1 px-3 text-xs font-medium uppercase tracking-wider"
          style="color: var(--text-muted)"
        >
          Public
        </div>

        <RouterLink
          to="/status-pages"
          class="flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors"
          style="color: var(--text-dim)"
          active-class="font-medium"
        >
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <circle cx="12" cy="12" r="10" />
            <line x1="2" y1="12" x2="22" y2="12" />
            <path
              d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"
            />
          </svg>
          Status pages
        </RouterLink>

        <RouterLink
          to="/maintenance"
          class="flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors"
          style="color: var(--text-dim)"
          active-class="font-medium"
        >
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <circle cx="12" cy="12" r="10" />
            <path d="M12 6v6l4 2" />
          </svg>
          Maintenance
        </RouterLink>

        <div
          class="pt-2 pb-1 px-3 text-xs font-medium uppercase tracking-wider"
          style="color: var(--text-muted)"
        >
          Account
        </div>

        <RouterLink
          to="/billing"
          class="flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors"
          style="color: var(--text-dim)"
          active-class="font-medium"
        >
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <rect x="1" y="4" width="22" height="16" rx="2" ry="2" />
            <line x1="1" y1="10" x2="23" y2="10" />
          </svg>
          Billing
        </RouterLink>

        <RouterLink
          to="/settings"
          class="flex items-center gap-2 px-3 py-2 rounded-md text-sm transition-colors"
          style="color: var(--text-dim)"
          active-class="font-medium"
        >
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <circle cx="12" cy="12" r="3" />
            <path
              d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"
            />
          </svg>
          Settings
        </RouterLink>
      </nav>

      <!-- User -->
      <div class="px-4 py-3 border-t" style="border-color: var(--border)">
        <div class="text-xs truncate mb-2" style="color: var(--text-muted)">
          {{ auth.user?.email }}
        </div>
        <button
          class="text-xs w-full text-left transition-colors hover:cursor-pointer"
          style="color: var(--text-muted)"
          @click="signOut"
        >
          Sign out
        </button>
      </div>
    </aside>

    <!-- Main column -->
    <div class="flex flex-col flex-1 min-w-0">
      <!-- Mobile header (hidden on desktop) -->
      <header
        class="flex md:hidden items-center gap-3 px-4 h-14 border-b flex-shrink-0"
        style="background-color: var(--surface); border-color: var(--border)"
      >
        <button
          class="p-1 -ml-1 rounded"
          style="color: var(--text-muted)"
          aria-label="Open menu"
          @click="sidebarOpen = true"
        >
          <svg
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <line x1="3" y1="6" x2="21" y2="6" />
            <line x1="3" y1="12" x2="21" y2="12" />
            <line x1="3" y1="18" x2="21" y2="18" />
          </svg>
        </button>
        <img :src="logoDark" alt="" class="h-5" />
        <div class="text-xs mt-1" style="color: var(--text-muted)">
          {{ appVersion }}
        </div>
      </header>

      <!-- Page content -->
      <main class="flex-1 overflow-y-auto">
        <slot />
      </main>
    </div>
  </div>
</template>
