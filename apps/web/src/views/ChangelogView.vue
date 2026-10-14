<script setup lang="ts">
import { RouterLink } from 'vue-router'
import LandingLayout from '@/layouts/LandingLayout.vue'
import PageHero from '@/components/PageHero.vue'
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
    <PageHero title="Changelog">
      Every release, newest first. Guides and essays live on the
      <RouterLink to="/blog" class="text-[var(--accent)] underline hover:no-underline"
        >blog</RouterLink
      >.
    </PageHero>

    <section class="mx-auto max-w-[1200px] px-4 pb-24 sm:px-6">
      <article
        v-for="release in sortedReleases"
        :id="release.slug"
        :key="release.slug"
        class="grid scroll-mt-20 gap-x-12 gap-y-2 border-t border-[var(--border)] py-8 md:grid-cols-[260px_minmax(0,1fr)]"
      >
        <div class="text-[13px] text-[var(--text-muted)]">{{ release.date }}</div>
        <div class="max-w-[680px]">
          <h2 class="mb-2 text-xl font-semibold tracking-[-0.02em] text-[var(--text-strong)]">
            {{ release.title }}
          </h2>
          <p class="mb-4 text-[15px] leading-relaxed text-[var(--text-dim)]">
            {{ release.excerpt }}
          </p>
          <RouterLink
            :to="`/blog/${release.slug}`"
            class="text-sm font-medium text-[var(--accent)] hover:underline"
          >
            Read the full release notes →
          </RouterLink>
        </div>
      </article>
    </section>
  </LandingLayout>
</template>
