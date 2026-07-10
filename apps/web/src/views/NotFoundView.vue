<script setup lang="ts">
import { useRoute, RouterLink } from 'vue-router'
import { useHead } from '@unhead/vue'
import LandingLayout from '@/layouts/LandingLayout.vue'
import { useSeo } from '@/composables/useSeo'

const route = useRoute()

useSeo({
  title: 'Page not found — checkmeup',
  description: "The page you're looking for doesn't exist or may have moved.",
  path: () => route.fullPath,
})

// Any unmatched path lands here (see router's catch-all route) — always
// noindex, unlike BlogPostView's conditional noindex for a bad slug.
useHead({
  meta: [{ name: 'robots', content: 'noindex' }],
})
</script>

<template>
  <LandingLayout>
    <!-- Design: Claude Design "CheckMeUp 404" — variant 1a (standard, on-brand) -->
    <section class="relative py-24 sm:py-32 flex items-center justify-center overflow-hidden">
      <!-- Grid texture -->
      <div
        class="absolute inset-0 pointer-events-none"
        style="
          background-image:
            linear-gradient(var(--border) 1px, transparent 1px),
            linear-gradient(90deg, var(--border) 1px, transparent 1px);
          background-size: 48px 48px;
        "
      ></div>
      <!-- Accent glow, top-right -->
      <div
        class="absolute pointer-events-none"
        style="
          top: -80px;
          right: -100px;
          width: 600px;
          height: 500px;
          background: radial-gradient(ellipse at 70% 30%, var(--accent-wash) 0%, transparent 65%);
        "
      ></div>

      <div class="relative z-10 text-center max-w-[520px] px-4 sm:px-6">
        <div
          class="inline-flex items-center gap-1.5 text-xs font-medium px-3 py-1.5 rounded-full mb-7"
          style="border: 1px solid var(--border); background-color: var(--card); color: var(--text-dim)"
        >
          <span
            class="w-1.5 h-1.5 rounded-full flex-shrink-0 animate-pulse"
            style="background-color: var(--status-down)"
          ></span>
          Monitor status: not found
        </div>

        <div
          class="font-mono text-8xl sm:text-9xl font-semibold tracking-tight leading-none mb-2"
          style="color: var(--text-strong)"
        >
          404
        </div>

        <h1
          class="text-2xl sm:text-3xl font-extrabold tracking-tight mb-4"
          style="color: var(--text-strong)"
        >
          This page went down and stayed down.
        </h1>

        <p class="text-[15px] leading-relaxed mb-8" style="color: var(--text-dim)">
          The page you're looking for doesn't exist, moved, or never came back up. Unlike your
          client's site, we won't be paging you about it.
        </p>

        <div class="flex items-center justify-center gap-2.5 flex-wrap">
          <RouterLink
            to="/dashboard"
            class="inline-flex items-center px-5 py-2.5 rounded-lg text-sm font-semibold transition-colors"
            style="background-color: var(--accent); color: var(--on-accent)"
          >
            Back to dashboard
          </RouterLink>
          <RouterLink
            to="/"
            class="inline-flex items-center px-5 py-2.5 rounded-lg text-sm transition-colors"
            style="border: 1px solid var(--border); color: var(--text-dim)"
          >
            Go to homepage
          </RouterLink>
        </div>
      </div>
    </section>
  </LandingLayout>
</template>
