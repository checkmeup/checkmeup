<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { useTheme } from '@/lib/theme'
import { appVersion } from '@/lib/version'
import ThemeToggle from '@/components/ThemeToggle.vue'

const { logo } = useTheme()
const mobileMenuOpen = ref(false)
// A post's route is a sibling of /blog, not nested under it, so RouterLink's
// active-class alone leaves Blog unhighlighted on /blog/<slug>.
const route = useRoute()
const onBlogPost = computed(() => route.path.startsWith('/blog/'))

// The footer's link columns. Labels stay short enough not to wrap: the
// column's heading carries the context ("Monitoring", not "Cron job monitoring").
const footerColumns = [
  {
    title: 'Product',
    links: [
      { to: '/pricing', label: 'Pricing' },
      { to: '/docs', label: 'Docs' },
      { to: '/faq', label: 'FAQ' },
      { to: '/changelog', label: 'Changelog' },
      { to: '/sign-in', label: 'Sign in' },
      { to: '/sign-up', label: 'Sign up' },
    ],
  },
  {
    title: 'Monitoring',
    links: [
      { to: '/cron-monitoring', label: 'Cron jobs' },
      { to: '/uptime-monitoring', label: 'Uptime' },
      { to: '/ssl-monitoring', label: 'SSL certificates' },
      { to: '/domain-expiry-monitoring', label: 'Domain expiry' },
      { to: '/port-monitoring', label: 'TCP ports' },
      { to: '/dns-monitoring', label: 'DNS records' },
    ],
  },
  {
    title: 'Resources',
    links: [
      { to: '/blog', label: 'Blog' },
      { to: '/tools/ssl-checker', label: 'SSL checker' },
      { to: '/tools/domain-expiry-checker', label: 'Domain expiry checker' },
      { to: '/tools/dns-lookup', label: 'DNS lookup' },
      { to: '/alternatives/healthchecks-io', label: 'vs Healthchecks.io' },
      { to: '/alternatives/uptimerobot', label: 'vs UptimeRobot' },
      { to: '/alternatives/cronitor', label: 'vs Cronitor' },
    ],
  },
  {
    title: 'Company',
    links: [
      { to: '/about', label: 'About' },
      { to: '/terms', label: 'Terms' },
      { to: '/privacy', label: 'Privacy' },
      { to: '/refund', label: 'Refund policy' },
    ],
  },
]
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
            :class="{ '!text-[var(--text-strong)] font-medium': onBlogPost }"
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
          :class="{ '!text-[var(--text-strong)] font-medium': onBlogPost }"
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
        <div class="mb-9 flex flex-col items-start gap-10 lg:flex-row lg:justify-between">
          <div class="max-w-[260px]">
            <img :src="logo" alt="Checkmeup" class="h-6 mb-2" />
            <p class="text-[13px] leading-relaxed text-[var(--text-dim)]">
              Cron, uptime, SSL, domain, port, and DNS monitoring for freelancers and solo devs.
            </p>
          </div>
          <div class="grid w-full grid-cols-2 gap-x-8 gap-y-8 sm:grid-cols-4 lg:w-auto lg:gap-x-14">
            <nav
              v-for="column in footerColumns"
              :key="column.title"
              :aria-label="column.title"
              class="flex flex-col gap-2.5 text-[13px] text-[var(--text-dim)]"
            >
              <h2 class="mb-1 text-xs font-semibold text-[var(--text-strong)]">{{ column.title }}</h2>
              <RouterLink
                v-for="link in column.links"
                :key="link.to"
                :to="link.to"
                class="whitespace-nowrap transition-colors hover:text-[var(--text-strong)]"
                >{{ link.label }}</RouterLink
              >
            </nav>
          </div>
        </div>
        <div class="flex items-center justify-between gap-4 pt-5 border-t border-[var(--border)]">
          <span class="text-xs text-[var(--text-muted)]"
            >© 2026 checkmeup.net ·
            <RouterLink to="/changelog" class="transition-colors hover:text-[var(--text-strong)]">{{
              appVersion
            }}</RouterLink></span
          >
          <div class="flex items-center gap-4 text-xs text-[var(--text-muted)]">
            <!-- A plain link, not a RouterLink: the Go server renders status pages. -->
            <a href="/status/checkmeup-net" class="transition-colors hover:text-[var(--text-strong)]">Status</a>
            <a href="https://github.com/checkmeup/checkmeup" class="transition-colors hover:text-[var(--text-strong)]">GitHub</a>
          </div>
        </div>
      </div>
    </footer>
  </div>
</template>
