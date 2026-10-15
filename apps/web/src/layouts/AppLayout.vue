<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/api/client'
import { useTheme } from '@/lib/theme'
import { appVersion } from '@/lib/version'
import ThemeToggle from '@/components/ThemeToggle.vue'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const { logo } = useTheme()

const menuOpen = ref(false)
const accountOpen = ref(false)

watch(
  () => route.path,
  () => {
    menuOpen.value = false
    accountOpen.value = false
  },
)

// A section stays highlighted on its detail, create and edit pages too —
// e.g. "Monitors" on /monitors/uptime/42.
const sections = [
  { label: 'Overview', to: '/dashboard', prefix: '/dashboard' },
  { label: 'Monitors', to: '/monitors', prefix: '/monitors' },
  { label: 'Projects', to: '/projects', prefix: '/projects' },
  { label: 'Status pages', to: '/status-pages', prefix: '/status-pages' },
  { label: 'Incidents', to: '/incidents', prefix: '/incidents' },
  { label: 'Maintenance', to: '/maintenance', prefix: '/maintenance' },
]

function isActive(prefix: string) {
  return route.path === prefix || route.path.startsWith(prefix + '/')
}

const initials = computed(() => (auth.user?.email ?? '?').slice(0, 2).toUpperCase())

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
  <div class="min-h-screen bg-[var(--bg)]">
    <header
      class="sticky top-0 z-30 border-b border-[var(--border)] bg-[var(--nav-bg)] backdrop-blur-md"
    >
      <div
        class="mx-auto flex min-h-16 max-w-[1200px] items-center justify-between gap-4 px-4 md:px-6"
      >
        <div class="flex items-center gap-8">
          <RouterLink to="/dashboard" class="flex-shrink-0" aria-label="Checkmeup — overview">
            <img :src="logo" alt="" class="block h-6 w-auto" />
          </RouterLink>
          <nav aria-label="Main" class="hidden gap-1 md:flex">
            <RouterLink
              v-for="s in sections"
              :key="s.to"
              :to="s.to"
              class="rounded-lg px-3 py-2.5 text-sm transition-colors focus-visible:ring-2 focus-visible:ring-[var(--accent)] focus-visible:outline-none"
              :class="
                isActive(s.prefix)
                  ? 'bg-[var(--surface-raised)] font-semibold text-[var(--text-strong)]'
                  : 'text-[var(--text-dim)] hover:text-[var(--text)]'
              "
              :aria-current="isActive(s.prefix) ? 'page' : undefined"
            >
              {{ s.label }}
            </RouterLink>
          </nav>
        </div>

        <div class="flex items-center gap-2">
          <ThemeToggle />

          <!-- Account menu (desktop) -->
          <div class="relative hidden md:block">
            <button
              type="button"
              class="flex h-9 w-9 items-center justify-center rounded-full bg-[var(--accent-wash)] text-[13px] font-semibold text-[var(--accent)] focus-visible:ring-2 focus-visible:ring-[var(--accent)] focus-visible:outline-none"
              :aria-expanded="accountOpen"
              aria-haspopup="menu"
              aria-label="Account menu"
              @click="accountOpen = !accountOpen"
            >
              {{ initials }}
            </button>
            <template v-if="accountOpen">
              <div class="fixed inset-0 z-40" @click="accountOpen = false"></div>
              <div
                role="menu"
                class="absolute top-[calc(100%+8px)] right-0 z-50 w-60 overflow-hidden rounded-xl border border-[var(--border)] bg-[var(--bg)] py-1.5 shadow-[0_16px_40px_rgb(0_0_0/18%)]"
              >
                <div class="border-b border-[var(--border)] px-4 pt-1.5 pb-3">
                  <div class="truncate text-sm text-[var(--text-strong)]">
                    {{ auth.user?.email }}
                  </div>
                  <div class="mt-0.5 font-mono text-xs text-[var(--text-muted)]">
                    {{ appVersion }}
                  </div>
                </div>
                <RouterLink
                  to="/settings"
                  role="menuitem"
                  class="block px-4 py-2.5 text-sm text-[var(--text)] hover:bg-[var(--surface-raised)]"
                  >Settings</RouterLink
                >
                <RouterLink
                  to="/billing"
                  role="menuitem"
                  class="block px-4 py-2.5 text-sm text-[var(--text)] hover:bg-[var(--surface-raised)]"
                  >Billing</RouterLink
                >
                <button
                  type="button"
                  role="menuitem"
                  class="block w-full px-4 py-2.5 text-left text-sm text-[var(--text)] hover:bg-[var(--surface-raised)]"
                  @click="signOut"
                >
                  Sign out
                </button>
              </div>
            </template>
          </div>

          <!-- Menu button (phones) -->
          <button
            type="button"
            class="flex h-10 w-10 items-center justify-center rounded-lg text-[var(--text-dim)] md:hidden"
            :aria-expanded="menuOpen"
            :aria-label="menuOpen ? 'Close menu' : 'Open menu'"
            @click="menuOpen = !menuOpen"
          >
            <svg
              aria-hidden="true"
              width="20"
              height="20"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
            >
              <template v-if="menuOpen">
                <line x1="6" y1="6" x2="18" y2="18" />
                <line x1="18" y1="6" x2="6" y2="18" />
              </template>
              <template v-else>
                <line x1="3" y1="6" x2="21" y2="6" />
                <line x1="3" y1="12" x2="21" y2="12" />
                <line x1="3" y1="18" x2="21" y2="18" />
              </template>
            </svg>
          </button>
        </div>
      </div>

      <!-- Phone menu panel -->
      <nav
        v-if="menuOpen"
        aria-label="Main"
        class="flex flex-col border-t border-[var(--border)] px-4 py-3 md:hidden"
      >
        <RouterLink
          v-for="s in sections"
          :key="s.to"
          :to="s.to"
          class="rounded-lg px-3 py-3 text-[15px]"
          :class="
            isActive(s.prefix)
              ? 'bg-[var(--surface-raised)] font-semibold text-[var(--text-strong)]'
              : 'text-[var(--text-dim)]'
          "
        >
          {{ s.label }}
        </RouterLink>
        <div class="my-2 h-px bg-[var(--border)]"></div>
        <RouterLink to="/settings" class="rounded-lg px-3 py-3 text-[15px] text-[var(--text-dim)]"
          >Settings</RouterLink
        >
        <RouterLink to="/billing" class="rounded-lg px-3 py-3 text-[15px] text-[var(--text-dim)]"
          >Billing</RouterLink
        >
        <button
          type="button"
          class="rounded-lg px-3 py-3 text-left text-[15px] text-[var(--text-dim)]"
          @click="signOut"
        >
          Sign out
        </button>
        <div class="px-3 pt-2 text-xs text-[var(--text-muted)]">
          {{ auth.user?.email }} · <span class="font-mono">{{ appVersion }}</span>
        </div>
      </nav>
    </header>

    <main>
      <slot />
    </main>
  </div>
</template>
