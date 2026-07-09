<script setup lang="ts">
import { RouterLink } from 'vue-router'
import LandingLayout from '@/layouts/LandingLayout.vue'
import { posts } from '@/blog/posts'
import { useSeo } from '@/composables/useSeo'

useSeo({
  title: 'Blog — checkmeup',
  description: 'Building notes, release changelogs, and the occasional opinion from checkmeup.',
  path: '/blog',
})

const sortedPosts = [...posts].reverse()
</script>

<template>
  <LandingLayout>

    <section class="max-w-3xl mx-auto px-4 sm:px-6 pt-16 pb-10 sm:pt-24 sm:pb-14">
      <h1 class="text-4xl sm:text-5xl font-bold tracking-tight mb-3" style="color: var(--text-strong)">Blog</h1>
      <p class="text-lg" style="color: var(--text-dim)">Building notes, release changelogs, and the occasional opinion.</p>
    </section>

    <section class="max-w-3xl mx-auto px-4 sm:px-6 pb-24">
      <div class="space-y-0 divide-y" style="border-color: var(--border)">
        <RouterLink
          v-for="post in sortedPosts"
          :key="post.slug"
          :to="`/blog/${post.slug}`"
          class="block py-8 group"
          style="border-color: var(--border)"
        >
          <div class="flex items-center gap-3 mb-3 text-xs" style="color: var(--text-muted)">
            <span>{{ post.date }}</span>
            <span>·</span>
            <span>{{ post.readTime }}</span>
          </div>
          <h2
            class="text-xl font-bold mb-2 transition-colors group-hover:underline"
            style="color: var(--text-strong); text-decoration-color: var(--color-green-500)"
          >
            {{ post.title }}
          </h2>
          <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">{{ post.excerpt }}</p>
          <span class="text-sm font-medium transition-colors" style="color: var(--color-green-500)">
            Read post →
          </span>
        </RouterLink>
      </div>
    </section>

  </LandingLayout>
</template>
