<script setup lang="ts">
import { RouterLink } from 'vue-router'
import LandingLayout from '@/layouts/LandingLayout.vue'
import { releasesMeta } from '@/blog/posts'
import { useSeo } from '@/composables/useSeo'

// The one indexable home for release history: each vX.Y post stays at its
// own /blog/:slug URL for readers but is noindex there (see
// BlogPostMeta.kind), so this page is what search engines see instead.
useSeo({
  title: 'Changelog — Checkmeup',
  description:
    'Every Checkmeup release, newest first — new monitor types, alerting changes, status page updates, and fixes.',
  path: '/changelog',
})

const sortedReleases = [...releasesMeta].reverse()
</script>

<template>
  <LandingLayout>

    <section class="max-w-3xl mx-auto px-4 sm:px-6 pt-16 pb-10 sm:pt-24 sm:pb-14">
      <h1 class="text-[44px] leading-[1.02] font-semibold tracking-[-0.045em] sm:text-[60px] mb-3 text-[var(--text-strong)]">Changelog</h1>
      <p class="text-lg text-[var(--text-dim)]">Every release, newest first. Guides and essays live on the <RouterLink to="/blog" class="underline hover:no-underline text-[var(--accent)]">blog</RouterLink>.</p>
    </section>

    <section class="max-w-3xl mx-auto px-4 sm:px-6 pb-24">
      <div class="space-y-0 divide-y border-[var(--border)]">
        <article
          v-for="release in sortedReleases"
          :id="release.slug"
          :key="release.slug"
          class="py-8 scroll-mt-20 border-[var(--border)]"
        >
          <div class="mb-3 text-xs text-[var(--text-muted)]">{{ release.date }}</div>
          <h2 class="mb-2 text-xl font-semibold tracking-[-0.02em] text-[var(--text-strong)]">{{ release.title }}</h2>
          <p class="text-sm leading-relaxed mb-4 text-[var(--text-dim)]">{{ release.excerpt }}</p>
          <RouterLink
            :to="`/blog/${release.slug}`"
            class="text-sm font-medium hover:underline text-[var(--accent)]"
          >
            Read the full release notes →
          </RouterLink>
        </article>
      </div>
    </section>

  </LandingLayout>
</template>
