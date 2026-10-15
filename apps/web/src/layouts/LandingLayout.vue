<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import { useTheme } from '@/lib/theme'
import { appVersion } from '@/lib/version'
import ThemeToggle from '@/components/ThemeToggle.vue'

const { logo } = useTheme()
const mobileMenuOpen = ref(false)
</script>

<template>
  <div class="flex min-h-screen flex-col bg-[var(--bg)] text-[var(--text)]">
    <!-- Top navigation -->
    <header
      class="sticky top-0 z-50 border-b border-[var(--border)] bg-[var(--nav-bg)] backdrop-blur-md"
    >
      <div class="max-w-[1200px] mx-auto px-4 sm:px-6 flex items-center gap-8 h-16">
        <RouterLink to="/" class="flex-shrink-0">
          <img :src="logo" alt="Checkmeup" class="h-6" />
        </RouterLink>

        <nav class="hidden md:flex items-center gap-1 text-sm text-[var(--text-dim)]">
          <RouterLink
            to="/docs"
            class="px-3 py-2 rounded-lg hover:text-[var(--text-strong)] transition-colors"
            active-class="!text-[var(--text-strong)] font-medium"
            >Docs</RouterLink
          >
          <RouterLink
            to="/faq"
            class="px-3 py-2 rounded-lg hover:text-[var(--text-strong)] transition-colors"
            active-class="!text-[var(--text-strong)] font-medium"
            >FAQ</RouterLink
          >
          <RouterLink
            to="/pricing"
            class="px-3 py-2 rounded-lg hover:text-[var(--text-strong)] transition-colors"
            active-class="!text-[var(--text-strong)] font-medium"
            >Pricing</RouterLink
          >
          <RouterLink
            to="/blog"
            class="px-3 py-2 rounded-lg hover:text-[var(--text-strong)] transition-colors"
            active-class="!text-[var(--text-strong)] font-medium"
            >Blog</RouterLink
          >
          <RouterLink
            to="/about"
            class="px-3 py-2 rounded-lg hover:text-[var(--text-strong)] transition-colors"
            active-class="!text-[var(--text-strong)] font-medium"
            >About</RouterLink
          >
        </nav>

        <div class="hidden md:flex items-center gap-2.5 ml-auto">
          <ThemeToggle />
          <RouterLink
            to="/sign-in"
            class="text-sm px-3.5 py-2 rounded-lg hover:text-[var(--text-strong)] transition-colors text-[var(--text-dim)]"
          >
            Sign in
          </RouterLink>
          <RouterLink
            to="/sign-up"
            class="text-sm font-semibold px-4 py-2 rounded-[10px] transition-colors hover:bg-[var(--accent-hover)] bg-[var(--accent)] text-[var(--on-accent)]"
          >
            Sign up free
          </RouterLink>
        </div>

        <button
          class="md:hidden p-2 rounded-[10px] ml-auto text-[var(--text-muted)]"
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
        class="md:hidden border-t px-4 py-4 space-y-3 border-[var(--border)] bg-[var(--bg)]"
      >
        <RouterLink
          to="/docs"
          class="block text-sm py-1 text-[var(--text-dim)]"
          active-class="!text-[var(--text-strong)] font-medium"
          >Docs</RouterLink
        >
        <RouterLink
          to="/faq"
          class="block text-sm py-1 text-[var(--text-dim)]"
          active-class="!text-[var(--text-strong)] font-medium"
          >FAQ</RouterLink
        >
        <RouterLink
          to="/pricing"
          class="block text-sm py-1 text-[var(--text-dim)]"
          active-class="!text-[var(--text-strong)] font-medium"
          >Pricing</RouterLink
        >
        <RouterLink
          to="/blog"
          class="block text-sm py-1 text-[var(--text-dim)]"
          active-class="!text-[var(--text-strong)] font-medium"
          >Blog</RouterLink
        >
        <RouterLink
          to="/about"
          class="block text-sm py-1 text-[var(--text-dim)]"
          active-class="!text-[var(--text-strong)] font-medium"
          >About</RouterLink
        >
        <div class="pt-2 flex flex-col gap-2">
          <RouterLink
            to="/sign-in"
            class="block text-sm text-center px-4 py-2 rounded-[10px] border hover:text-[var(--text-strong)] transition-colors text-[var(--text-dim)] border-[var(--border)]"
          >
            Sign in
          </RouterLink>
          <RouterLink
            to="/sign-up"
            class="block text-sm font-semibold text-center px-4 py-2.5 rounded-[10px] transition-colors hover:bg-[var(--accent-hover)] bg-[var(--accent)] text-[var(--on-accent)]"
          >
            Sign up free
          </RouterLink>
        </div>
      </div>
    </header>

    <!-- Page content -->
    <main class="flex-1">
      <slot />
    </main>

    <!-- Footer -->
    <footer class="border-t border-[var(--border)]">
      <div class="max-w-[1200px] mx-auto px-4 sm:px-6 py-12">
        <div class="flex flex-col sm:flex-row gap-10 sm:gap-12 items-start mb-9">
          <div class="flex-shrink-0 sm:min-w-[190px]">
            <img :src="logo" alt="Checkmeup" class="h-6 mb-2" />
            <p class="text-[13px] leading-relaxed text-[var(--text-dim)]">
              Cron, uptime, SSL, domain, port, and DNS monitoring for freelancers and solo devs.
            </p>
          </div>
          <div class="grid grid-cols-2 sm:grid-cols-3 gap-x-10 sm:gap-x-14 gap-y-6 sm:ml-auto">
            <nav class="flex flex-col gap-2.5 text-[13px] text-[var(--text-dim)]">
              <RouterLink to="/docs" class="hover:text-[var(--text-strong)] transition-colors"
                >Docs</RouterLink
              >
              <RouterLink to="/faq" class="hover:text-[var(--text-strong)] transition-colors"
                >FAQ</RouterLink
              >
              <RouterLink to="/pricing" class="hover:text-[var(--text-strong)] transition-colors"
                >Pricing</RouterLink
              >
              <RouterLink to="/blog" class="hover:text-[var(--text-strong)] transition-colors"
                >Blog</RouterLink
              >
              <RouterLink to="/changelog" class="hover:text-[var(--text-strong)] transition-colors"
                >Changelog</RouterLink
              >
            </nav>
            <nav class="flex flex-col gap-2.5 text-[13px] text-[var(--text-dim)]">
              <RouterLink to="/about" class="hover:text-[var(--text-strong)] transition-colors"
                >About</RouterLink
              >
              <RouterLink to="/sign-in" class="hover:text-[var(--text-strong)] transition-colors"
                >Sign in</RouterLink
              >
              <RouterLink to="/sign-up" class="hover:text-[var(--text-strong)] transition-colors"
                >Sign up</RouterLink
              >
            </nav>
            <nav class="flex flex-col gap-2.5 text-[13px] text-[var(--text-dim)]">
              <RouterLink to="/terms" class="hover:text-[var(--text-strong)] transition-colors"
                >Terms</RouterLink
              >
              <RouterLink to="/privacy" class="hover:text-[var(--text-strong)] transition-colors"
                >Privacy</RouterLink
              >
              <RouterLink to="/refund" class="hover:text-[var(--text-strong)] transition-colors"
                >Refund Policy</RouterLink
              >
            </nav>
          </div>
        </div>
        <div class="pt-5 border-t border-[var(--border)]">
          <span class="text-xs text-[var(--text-muted)]"
            >© 2026 checkmeup.net ·
            <RouterLink to="/changelog" class="transition-colors hover:text-[var(--text-strong)]">{{
              appVersion
            }}</RouterLink></span
          >
        </div>
      </div>
    </footer>
  </div>
</template>
