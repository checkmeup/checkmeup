<script setup lang="ts">
import { ref, computed } from 'vue'
import { RouterLink } from 'vue-router'
import { useTheme } from '@/lib/theme'
import ThemeToggle from '@/components/ThemeToggle.vue'
import logoDark from '@/assets/logo-dark.svg'
import logoLight from '@/assets/logo-light.svg'
import logoGrey from '@/assets/logo-grey.svg'

const { theme } = useTheme()
const logo = computed(() => (theme.value === 'light' ? logoLight : logoDark))
const mobileMenuOpen = ref(false)
</script>

<template>
  <div style="background-color: var(--bg); color: var(--text); min-height: 100vh">
    <!-- Top navigation -->
    <header
      class="sticky top-0 z-50 border-b"
      style="
        background-color: var(--nav-bg);
        border-color: var(--border);
        backdrop-filter: blur(12px);
      "
    >
      <div class="max-w-[1100px] mx-auto px-4 sm:px-7 flex items-center gap-6 h-14">
        <RouterLink to="/" class="flex-shrink-0">
          <img :src="logo" alt="checkmeup" class="h-[22px]" />
        </RouterLink>

        <nav class="hidden md:flex items-center gap-0.5 text-sm" style="color: var(--text-muted)">
          <RouterLink to="/docs" class="px-2.5 py-1.5 rounded-md hover:text-[var(--text-strong)] transition-colors">Docs</RouterLink>
          <RouterLink to="/faq" class="px-2.5 py-1.5 rounded-md hover:text-[var(--text-strong)] transition-colors">FAQ</RouterLink>
          <RouterLink to="/pricing" class="px-2.5 py-1.5 rounded-md hover:text-[var(--text-strong)] transition-colors">Pricing</RouterLink>
          <RouterLink to="/blog" class="px-2.5 py-1.5 rounded-md hover:text-[var(--text-strong)] transition-colors">Blog</RouterLink>
          <RouterLink to="/about" class="px-2.5 py-1.5 rounded-md hover:text-[var(--text-strong)] transition-colors">About</RouterLink>
        </nav>

        <div class="hidden md:flex items-center gap-2.5 ml-auto">
          <ThemeToggle />
          <RouterLink
            to="/sign-in"
            class="text-sm px-3.5 py-1.5 rounded-md hover:text-[var(--text-strong)] transition-colors"
            style="color: var(--text-muted)"
          >
            Sign in
          </RouterLink>
          <RouterLink
            to="/sign-up"
            class="text-sm font-semibold px-4 py-2 rounded-lg transition-opacity hover:opacity-90"
            style="background-color: var(--accent); color: var(--on-accent)"
          >
            Sign up free
          </RouterLink>
        </div>

        <button
          class="md:hidden p-2 rounded-md ml-auto"
          style="color: var(--text-muted)"
          aria-label="Toggle menu"
          @click="mobileMenuOpen = !mobileMenuOpen"
        >
          <svg
            v-if="!mobileMenuOpen"
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
          <svg
            v-else
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </svg>
        </button>
      </div>

      <div
        v-if="mobileMenuOpen"
        class="md:hidden border-t px-4 py-4 space-y-3"
        style="border-color: var(--border); background-color: var(--surface)"
      >
        <RouterLink to="/docs" class="block text-sm py-1" style="color: var(--text-dim)">Docs</RouterLink>
        <RouterLink to="/faq" class="block text-sm py-1" style="color: var(--text-dim)">FAQ</RouterLink>
        <RouterLink to="/pricing" class="block text-sm py-1" style="color: var(--text-dim)"
          >Pricing</RouterLink
        >
        <RouterLink to="/blog" class="block text-sm py-1" style="color: var(--text-dim)">Blog</RouterLink>
        <RouterLink to="/about" class="block text-sm py-1" style="color: var(--text-dim)">About</RouterLink>
        <div class="pt-2 flex flex-col gap-2">
          <RouterLink
            to="/sign-in"
            class="block text-sm text-center px-4 py-2 rounded-md border hover:text-[var(--text-strong)] transition-colors"
            style="color: var(--text-dim); border-color: var(--border)"
          >
            Sign in
          </RouterLink>
          <RouterLink
            to="/sign-up"
            class="block text-sm font-semibold text-center px-4 py-2 rounded-lg transition-opacity hover:opacity-90"
            style="background-color: var(--accent); color: var(--on-accent)"
          >
            Sign up free
          </RouterLink>
        </div>
      </div>
    </header>

    <!-- Page content -->
    <slot />

    <!-- Footer -->
    <footer class="border-t" style="border-color: var(--border)">
      <div class="max-w-[1100px] mx-auto px-4 sm:px-6 py-10 sm:py-11">
        <div class="flex flex-col sm:flex-row gap-10 sm:gap-12 items-start mb-9">
          <div class="flex-shrink-0 sm:min-w-[190px]">
            <img :src="logoGrey" alt="checkmeup" class="h-6 mb-2" />
            <p class="text-xs leading-relaxed" style="color: var(--text-muted)">
              Cron, uptime, SSL, domain, and port monitoring for freelancers and solo devs.
            </p>
          </div>
          <div class="grid grid-cols-2 sm:grid-cols-3 gap-x-10 sm:gap-x-14 gap-y-6 sm:ml-auto">
            <nav class="flex flex-col gap-2.5 text-xs" style="color: var(--text-muted)">
              <RouterLink to="/docs" class="hover:text-[var(--text-strong)] transition-colors">Docs</RouterLink>
              <RouterLink to="/faq" class="hover:text-[var(--text-strong)] transition-colors">FAQ</RouterLink>
              <RouterLink to="/pricing" class="hover:text-[var(--text-strong)] transition-colors">Pricing</RouterLink>
              <RouterLink to="/blog" class="hover:text-[var(--text-strong)] transition-colors">Blog</RouterLink>
            </nav>
            <nav class="flex flex-col gap-2.5 text-xs" style="color: var(--text-muted)">
              <RouterLink to="/about" class="hover:text-[var(--text-strong)] transition-colors">About</RouterLink>
              <RouterLink to="/sign-in" class="hover:text-[var(--text-strong)] transition-colors">Sign in</RouterLink>
              <RouterLink to="/sign-up" class="hover:text-[var(--text-strong)] transition-colors">Sign up</RouterLink>
            </nav>
            <nav class="flex flex-col gap-2.5 text-xs" style="color: var(--text-muted)">
              <RouterLink to="/terms" class="hover:text-[var(--text-strong)] transition-colors">Terms</RouterLink>
              <RouterLink to="/privacy" class="hover:text-[var(--text-strong)] transition-colors">Privacy</RouterLink>
              <RouterLink to="/refund" class="hover:text-[var(--text-strong)] transition-colors">Refund Policy</RouterLink>
            </nav>
          </div>
        </div>
        <div class="pt-5 border-t" style="border-color: var(--border)">
          <span class="text-xs" style="color: var(--text-muted)">© 2026 checkmeup.net</span>
        </div>
      </div>
    </footer>
  </div>
</template>
