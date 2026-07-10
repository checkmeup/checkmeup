<script setup lang="ts">
import { RouterLink } from 'vue-router'

interface Cta {
  label: string
  to: string
}

defineProps<{
  badge: string
  heading: string
  description: string
  primaryCta: Cta
  secondaryCta: Cta
}>()
</script>

<template>
  <!-- Design: Claude Design "CheckMeUp 404" — variant 1a (standard, on-brand),
       shared between the site-wide 404 route and BlogPostView's unknown-slug state. -->
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
        {{ badge }}
      </div>

      <div
        class="font-mono text-8xl sm:text-9xl font-semibold tracking-tight leading-none mb-2"
        style="color: var(--text-strong)"
      >
        404
      </div>

      <h1 class="text-2xl sm:text-3xl font-extrabold tracking-tight mb-4" style="color: var(--text-strong)">
        {{ heading }}
      </h1>

      <p class="text-[15px] leading-relaxed mb-8" style="color: var(--text-dim)">
        {{ description }}
      </p>

      <div class="flex items-center justify-center gap-2.5 flex-wrap">
        <RouterLink
          :to="primaryCta.to"
          class="inline-flex items-center px-5 py-2.5 rounded-lg text-sm font-semibold transition-colors"
          style="background-color: var(--accent); color: var(--on-accent)"
        >
          {{ primaryCta.label }}
        </RouterLink>
        <RouterLink
          :to="secondaryCta.to"
          class="inline-flex items-center px-5 py-2.5 rounded-lg text-sm transition-colors"
          style="border: 1px solid var(--border); color: var(--text-dim)"
        >
          {{ secondaryCta.label }}
        </RouterLink>
      </div>
    </div>
  </section>
</template>
