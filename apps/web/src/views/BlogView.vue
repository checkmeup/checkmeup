<script setup lang="ts">
import { RouterLink } from 'vue-router'
import LandingLayout from '@/layouts/LandingLayout.vue'
import { articlesMeta } from '@/blog/posts'
import { useSeo } from '@/composables/useSeo'

useSeo({
  title: 'Blog — Checkmeup',
  description: 'Monitoring guides, building notes, and the occasional opinion from Checkmeup.',
  path: '/blog',
})

// Release notes are listed on /changelog instead (see BlogPostMeta.kind).
const sortedPosts = [...articlesMeta].reverse()
</script>

<template>
  <LandingLayout>

    <section class="max-w-3xl mx-auto px-4 sm:px-6 pt-16 pb-10 sm:pt-24 sm:pb-14">
      <h1 class="text-[44px] leading-[1.02] font-semibold tracking-[-0.045em] sm:text-[60px] mb-3 text-[var(--text-strong)]">Blog</h1>
      <p class="text-lg text-[var(--text-dim)]">Monitoring guides, building notes, and the occasional opinion. Release notes are in the <RouterLink to="/changelog" class="underline hover:no-underline text-[var(--accent)]">changelog</RouterLink>.</p>
    </section>

    <section class="max-w-3xl mx-auto px-4 sm:px-6 pb-24">
      <div class="space-y-0 divide-y border-[var(--border)]">
        <RouterLink
          v-for="post in sortedPosts"
          :key="post.slug"
          :to="`/blog/${post.slug}`"
          class="block py-8 group border-[var(--border)]"
        >
          <div class="flex items-center gap-3 mb-3 text-xs text-[var(--text-muted)]">
            <span>{{ post.date }}</span>
            <span>·</span>
            <span>{{ post.readTime }}</span>
          </div>
          <h2
            class="text-xl font-semibold mb-2 transition-colors group-hover:underline"
            style="color: var(--text-strong); text-decoration-color: var(--color-green-500)"
          >
            {{ post.title }}
          </h2>
          <p class="text-sm leading-relaxed mb-4 text-[var(--text-dim)]">{{ post.excerpt }}</p>
          <span class="text-sm font-medium transition-colors text-[var(--accent)]">
            Read post →
          </span>
        </RouterLink>
      </div>
    </section>

  </LandingLayout>
</template>
