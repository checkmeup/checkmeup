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
      <h1 class="text-4xl sm:text-5xl font-semibold tracking-[-0.035em] mb-3" style="color: var(--text-strong)">Changelog</h1>
      <p class="text-lg" style="color: var(--text-dim)">Every release, newest first. Guides and essays live on the <RouterLink to="/blog" class="underline hover:no-underline" style="color: var(--accent)">blog</RouterLink>.</p>
    </section>

    <section class="max-w-3xl mx-auto px-4 sm:px-6 pb-24">
      <div class="space-y-0 divide-y" style="border-color: var(--border)">
        <article
          v-for="release in sortedReleases"
          :id="release.slug"
          :key="release.slug"
          class="py-8 scroll-mt-20"
          style="border-color: var(--border)"
        >
          <div class="mb-3 text-xs" style="color: var(--text-muted)">{{ release.date }}</div>
          <h2 class="text-xl font-semibold mb-2" style="color: var(--text-strong)">{{ release.title }}</h2>
          <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">{{ release.excerpt }}</p>
          <RouterLink
            :to="`/blog/${release.slug}`"
            class="text-sm font-medium hover:underline"
            style="color: var(--accent)"
          >
            Read the full release notes →
          </RouterLink>
        </article>
      </div>
    </section>

  </LandingLayout>
</template>
