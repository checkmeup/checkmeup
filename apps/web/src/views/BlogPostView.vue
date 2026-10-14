<script setup lang="ts">
import { computed, onMounted, onServerPrefetch, ref, watch } from 'vue'
import { useRoute, RouterLink } from 'vue-router'
import { useHead } from '@unhead/vue'
import LandingLayout from '@/layouts/LandingLayout.vue'
import NotFoundHero from '@/components/NotFoundHero.vue'
import type { BlogPost } from '@/blog/posts'
import { getPost } from '@/blog/posts'
import { useSeo } from '@/composables/useSeo'

const route = useRoute()
const post = ref<BlogPost | undefined>(undefined)
// Distinguishes "haven't resolved the slug yet" from "resolved to nothing" —
// getPost() is async (each post's full content is a separate, lazily-loaded
// chunk, so the list page doesn't have to ship every post's prose), so
// !post is briefly true during a genuine in-flight load too.
const loading = ref(true)

async function loadPost() {
  loading.value = true
  post.value = await getPost(route.params.slug as string)
  loading.value = false
}

// onServerPrefetch: awaited by @vue/server-renderer's renderToString before
// it serializes HTML (see scripts/prerender.mts) — this is what gets a real
// post's content into the prerendered page a crawler or a non-JS client
// actually receives, same reasoning as ADR seo. onMounted covers the client
// browser's own render pass (a fresh mount, not a hydration of the
// prerendered HTML — ADR seo again), and the watcher covers navigating
// client-side from one post straight to another without a full reload.
onServerPrefetch(loadPost)
onMounted(loadPost)
watch(() => route.params.slug, loadPost)

useSeo({
  title: () => (post.value ? `${post.value.title} — Checkmeup blog` : 'Post not found — Checkmeup'),
  description: () => post.value?.excerpt ?? 'This blog post could not be found.',
  path: () => route.path,
})

const isRelease = computed(() => post.value?.kind === 'release')

// This route always serves HTTP 200 (client-rendered SPA), so an unknown
// slug would otherwise let a thin "not found" page get indexed. Release
// notes are noindex too, but keep `follow` so their links still pass weight
// to the guides and product pages (see BlogPostMeta.kind). The tag has to be
// in the prerendered HTML, not just added by JS after load — it is, because
// onServerPrefetch resolves the post before scripts/prerender.mts
// serializes <head>.
useHead({
  meta: [
    {
      name: 'robots',
      content: () => {
        if (!post.value) return 'noindex'
        return isRelease.value ? 'noindex, follow' : 'index, follow'
      },
    },
  ],
})

const backLink = computed(() =>
  isRelease.value
    ? { to: '/changelog', label: 'All releases' }
    : { to: '/blog', label: 'All posts' },
)

// Makes matched posts eligible for richer search listings (author, publish
// date). No entry at all for an unknown slug — same reasoning as the
// noindex tag above, nothing to describe.
const articleSchema = computed(() => {
  if (!post.value) return null
  const publishedAt = new Date(post.value.date).toISOString()
  return {
    '@context': 'https://schema.org',
    '@type': 'Article',
    headline: post.value.title,
    description: post.value.excerpt,
    datePublished: publishedAt,
    dateModified: publishedAt,
    author: { '@type': 'Person', name: 'Andrew Molyuk', url: 'https://checkmeup.net/about' },
    publisher: {
      '@type': 'Organization',
      name: 'Checkmeup',
      logo: { '@type': 'ImageObject', url: 'https://checkmeup.net/img/checkmeup-og.png' },
    },
    mainEntityOfPage: {
      '@type': 'WebPage',
      '@id': `https://checkmeup.net/blog/${post.value.slug}`,
    },
  }
})

useHead({
  script: computed(() =>
    articleSchema.value
      ? [{ type: 'application/ld+json', innerHTML: JSON.stringify(articleSchema.value) }]
      : [],
  ),
})
</script>

<template>
  <LandingLayout>
    <!-- Loading: the requested post's content chunk hasn't resolved yet -->
    <div v-if="loading" class="max-w-3xl mx-auto px-4 sm:px-6 py-24 text-center">
      <p class="text-sm text-[var(--text-muted)]">Loading…</p>
    </div>

    <!-- 404 -->
    <NotFoundHero
      v-else-if="!post"
      badge="Post status: not found"
      heading="This post doesn't exist, or it moved without telling anyone."
      description="The post you're looking for isn't here — check the link, or head back to the blog for everything that's actually been published."
      :primary-cta="{ label: 'Back to blog', to: '/blog' }"
      :secondary-cta="{ label: 'Go to homepage', to: '/' }"
    />

    <template v-else>
      <!-- Header -->
      <header class="mx-auto max-w-[1200px] px-4 pt-16 sm:px-6 sm:pt-24">
        <RouterLink
          :to="backLink.to"
          class="mb-8 inline-flex items-center gap-1.5 text-sm text-[var(--text-muted)] transition-colors hover:text-[var(--text)]"
        >
          <svg
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <polyline points="15 18 9 12 15 6" />
          </svg>
          {{ backLink.label }}
        </RouterLink>

        <h1
          class="mb-5 max-w-[900px] text-[36px] leading-[1.05] font-semibold tracking-[-0.04em] text-[var(--text-strong)] sm:text-[52px]"
        >
          {{ post.title }}
        </h1>

        <p class="max-w-2xl text-lg leading-relaxed text-[var(--text-dim)]">{{ post.excerpt }}</p>

        <div class="mt-10 border-t border-[var(--border)]"></div>
      </header>

      <div
        class="mx-auto grid max-w-[1200px] gap-x-12 gap-y-6 px-4 pt-10 pb-24 sm:px-6 md:grid-cols-[260px_minmax(0,1fr)]"
      >
        <div
          class="flex flex-col gap-1 text-[13px] text-[var(--text-muted)] md:sticky md:top-24 md:self-start"
        >
          <span>{{ post.date }}</span>
          <span>{{ post.readTime }}</span>
        </div>

        <div class="max-w-[680px] min-w-0">
          <!-- Content -->
          <article class="space-y-5">
            <template v-for="(block, i) in post.content" :key="i">
              <p v-if="block.type === 'p'" class="text-base leading-[1.7] text-[var(--text-dim)]">
                {{ block.text }}
              </p>

              <h2
                v-else-if="block.type === 'h2'"
                class="pt-6 text-[22px] font-semibold tracking-[-0.02em] text-[var(--text-strong)]"
              >
                {{ block.text }}
              </h2>

              <h3
                v-else-if="block.type === 'h3'"
                class="text-base font-semibold pt-2 text-[var(--text-strong)]"
              >
                {{ block.text }}
              </h3>

              <pre
                v-else-if="block.type === 'code'"
                class="overflow-x-auto rounded-[14px] bg-[var(--surface)] p-5 font-mono text-[13px] leading-relaxed text-[var(--text)]"
              ><code>{{ block.text }}</code></pre>

              <ul v-else-if="block.type === 'ul'" class="space-y-2 pl-1">
                <li
                  v-for="item in block.items"
                  :key="item"
                  class="flex items-start gap-3 text-base leading-relaxed text-[var(--text-dim)]"
                >
                  <span
                    class="mt-2.5 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-[var(--accent)]"
                  ></span>
                  {{ item }}
                </li>
              </ul>

              <figure v-else-if="block.type === 'image'" class="py-2">
                <img
                  :src="block.src"
                  :alt="block.alt"
                  class="w-full rounded-[14px] border border-[var(--border)]"
                  loading="lazy"
                />
                <figcaption
                  v-if="block.caption"
                  class="text-xs text-center mt-2 text-[var(--text-muted)]"
                >
                  {{ block.caption }}
                </figcaption>
              </figure>

              <div
                v-else-if="block.type === 'table'"
                class="rounded-[14px] border border-[var(--border)]"
              >
                <div
                  class="grid"
                  :style="{
                    gridTemplateColumns: `minmax(0, 1.3fr) repeat(${block.headers.length - 1}, minmax(0, 1fr))`,
                  }"
                >
                  <div class="contents">
                    <div
                      v-for="(h, hi) in block.headers"
                      :key="h"
                      class="px-2 py-3 text-xs font-semibold border-b break-words border-[var(--border)] bg-[var(--surface)] text-[var(--text-strong)]"
                      :class="hi === 0 ? 'text-left' : 'text-center'"
                    >
                      {{ h }}
                    </div>
                  </div>
                  <template v-for="(row, ri) in block.rows" :key="ri">
                    <div
                      v-for="(cell, ci) in row"
                      :key="ci"
                      class="px-2 py-3 text-xs border-b last:border-b-0 break-words"
                      :class="ci === 0 ? 'text-left' : 'text-center font-medium'"
                      :style="{
                        borderColor: 'var(--border)',
                        backgroundColor: ri % 2 === 0 ? 'var(--bg)' : 'var(--surface)',
                        color:
                          ci === 0
                            ? 'var(--text-dim)'
                            : cell === '—'
                              ? 'var(--text-muted)'
                              : cell === '✓'
                                ? 'var(--accent)'
                                : 'var(--text-strong)',
                      }"
                    >
                      {{ cell }}
                    </div>
                  </template>
                </div>
              </div>

              <blockquote
                v-else-if="block.type === 'blockquote'"
                class="border-l-2 border-[var(--accent)] py-1 pl-5 text-base leading-relaxed text-[var(--text)] italic"
              >
                {{ block.text }}
              </blockquote>

              <p
                v-else-if="block.type === 'signature'"
                class="text-base pt-2 italic"
                style="color: var(--text-strong); font-family: Georgia, 'Times New Roman', serif"
              >
                {{ block.text }}
              </p>

              <div
                v-else-if="block.type === 'divider'"
                class="border-t my-4 border-[var(--border)]"
              ></div>
            </template>
          </article>

          <!-- License note -->
          <p class="mt-16 text-xs leading-relaxed text-[var(--text-muted)]">
            This work is licensed under a
            <a
              href="https://creativecommons.org/licenses/by-nc-sa/4.0/"
              target="_blank"
              rel="noopener noreferrer"
              class="text-[var(--accent)] underline transition-colors hover:no-underline"
              >Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License</a
            >.
          </p>

          <!-- Footer nav -->
          <div class="mt-8 flex items-center justify-between border-t border-[var(--border)] pt-8">
            <RouterLink
              :to="backLink.to"
              class="inline-flex items-center gap-1.5 text-sm text-[var(--text-muted)] transition-colors hover:text-[var(--text)]"
            >
              <svg
                width="14"
                height="14"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
              >
                <polyline points="15 18 9 12 15 6" />
              </svg>
              {{ backLink.label }}
            </RouterLink>
            <RouterLink
              to="/sign-up"
              class="rounded-[10px] bg-[var(--accent)] px-4 py-2 text-sm font-medium text-[var(--on-accent)] transition-colors hover:bg-[var(--accent-hover)]"
            >
              Try Checkmeup free →
            </RouterLink>
          </div>
        </div>
      </div>
    </template>
  </LandingLayout>
</template>
