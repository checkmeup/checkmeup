<script setup lang="ts">
import { RouterLink } from 'vue-router'
import LandingLayout from '@/layouts/LandingLayout.vue'
import PageHero from '@/components/PageHero.vue'
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
    <PageHero title="Blog">
      Monitoring guides, building notes, and the occasional opinion. Release notes are in the
      <RouterLink to="/changelog" class="text-[var(--accent)] underline hover:no-underline"
        >changelog</RouterLink
      >.
    </PageHero>

    <section class="mx-auto max-w-[1200px] px-4 pb-24 sm:px-6">
      <RouterLink
        v-for="post in sortedPosts"
        :key="post.slug"
        :to="`/blog/${post.slug}`"
        class="group grid gap-x-12 gap-y-2 border-t border-[var(--border)] py-8 md:grid-cols-[260px_minmax(0,1fr)]"
      >
        <div class="text-[13px] text-[var(--text-muted)]">
          {{ post.date }} · {{ post.readTime }}
        </div>
        <div class="max-w-[680px]">
          <h2
            class="mb-2 text-xl font-semibold tracking-[-0.02em] text-[var(--text-strong)] decoration-[var(--accent)] underline-offset-4 group-hover:underline"
          >
            {{ post.title }}
          </h2>
          <p class="text-[15px] leading-relaxed text-[var(--text-dim)]">{{ post.excerpt }}</p>
        </div>
      </RouterLink>
    </section>
  </LandingLayout>
</template>
