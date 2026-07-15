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
// actually receives, same reasoning as ADR-037. onMounted covers the client
// browser's own render pass (a fresh mount, not a hydration of the
// prerendered HTML — ADR-037 again), and the watcher covers navigating
// client-side from one post straight to another without a full reload.
onServerPrefetch(loadPost)
onMounted(loadPost)
watch(() => route.params.slug, loadPost)

useSeo({
  title: () => (post.value ? `${post.value.title} — Checkmeup blog` : 'Post not found — Checkmeup'),
  description: () => post.value?.excerpt ?? 'This blog post could not be found.',
  path: () => route.fullPath,
})

// This route always serves HTTP 200 (client-rendered SPA), so an unknown
// slug would otherwise let a thin "not found" page get indexed.
useHead({
  meta: [{ name: 'robots', content: () => (post.value ? 'index, follow' : 'noindex') }],
})

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
    mainEntityOfPage: { '@type': 'WebPage', '@id': `https://checkmeup.net/blog/${post.value.slug}` },
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
      <p class="text-sm" style="color: var(--text-muted)">Loading…</p>
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
      <header class="max-w-3xl mx-auto px-4 sm:px-6 pt-16 pb-10 sm:pt-24 sm:pb-12">
        <RouterLink
          to="/blog"
          class="inline-flex items-center gap-1.5 text-sm mb-8 transition-colors"
          style="color: var(--text-muted)"
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
          All posts
        </RouterLink>

        <div class="flex items-center gap-3 mb-4 text-xs" style="color: var(--text-muted)">
          <span>{{ post.date }}</span>
          <span>·</span>
          <span>{{ post.readTime }}</span>
        </div>

        <h1
          class="text-3xl sm:text-4xl font-bold tracking-tight leading-tight mb-6"
          style="color: var(--text-strong)"
        >
          {{ post.title }}
        </h1>

        <p class="text-lg leading-relaxed" style="color: var(--text-dim)">{{ post.excerpt }}</p>

        <div class="mt-8 border-t" style="border-color: var(--border)"></div>
      </header>

      <!-- Content -->
      <article class="max-w-3xl mx-auto px-4 sm:px-6 pb-24 space-y-5">
        <template v-for="(block, i) in post.content" :key="i">
          <p
            v-if="block.type === 'p'"
            class="text-base leading-relaxed"
            style="color: var(--text-dim)"
          >
            {{ block.text }}
          </p>

          <h2
            v-else-if="block.type === 'h2'"
            class="text-xl font-bold pt-4"
            style="color: var(--text-strong)"
          >
            {{ block.text }}
          </h2>

          <h3
            v-else-if="block.type === 'h3'"
            class="text-base font-semibold pt-2"
            style="color: var(--text-strong)"
          >
            {{ block.text }}
          </h3>

          <pre
            v-else-if="block.type === 'code'"
            class="rounded-xl border p-5 text-xs overflow-x-auto font-mono leading-relaxed"
            style="
              background-color: var(--surface);
              border-color: var(--border);
              color: var(--color-green-300);
            "
          ><code>{{ block.text }}</code></pre>

          <ul v-else-if="block.type === 'ul'" class="space-y-2 pl-1">
            <li
              v-for="item in block.items"
              :key="item"
              class="flex items-start gap-3 text-sm leading-relaxed"
              style="color: var(--text-dim)"
            >
              <span
                class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-2"
                style="background-color: var(--color-green-500)"
              ></span>
              {{ item }}
            </li>
          </ul>

          <figure v-else-if="block.type === 'image'" class="py-2">
            <img
              :src="block.src"
              :alt="block.alt"
              class="rounded-xl border w-full"
              style="border-color: var(--border)"
              loading="lazy"
            />
            <figcaption
              v-if="block.caption"
              class="text-xs text-center mt-2"
              style="color: var(--text-muted)"
            >
              {{ block.caption }}
            </figcaption>
          </figure>

          <div
            v-else-if="block.type === 'table'"
            class="rounded-xl border"
            style="border-color: var(--border)"
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
                  class="px-2 py-3 text-xs font-semibold border-b break-words"
                  :class="hi === 0 ? 'text-left' : 'text-center'"
                  style="
                    border-color: var(--border);
                    background-color: var(--surface);
                    color: var(--text-strong);
                  "
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
                            ? 'var(--color-green-500)'
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
            class="border-l-2 pl-5 py-1 text-sm leading-relaxed italic"
            style="border-color: var(--color-green-500); color: var(--text-dim)"
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
            class="border-t my-4"
            style="border-color: var(--border)"
          ></div>
        </template>
      </article>

      <!-- License note -->
      <div class="max-w-3xl mx-auto px-4 sm:px-6 pb-8">
        <p class="text-xs leading-relaxed" style="color: var(--text-muted)">
          This work is licensed under a
          <a
            href="https://creativecommons.org/licenses/by-nc-sa/4.0/"
            target="_blank"
            rel="noopener noreferrer"
            class="underline hover:no-underline transition-colors"
            style="color: var(--color-green-500)"
            >Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License</a
          >.
        </p>
      </div>

      <!-- Footer nav -->
      <div
        class="max-w-3xl mx-auto px-4 sm:px-6 pb-20 border-t pt-8"
        style="border-color: var(--border)"
      >
        <div class="flex items-center justify-between">
          <RouterLink
            to="/blog"
            class="inline-flex items-center gap-1.5 text-sm transition-colors"
            style="color: var(--text-muted)"
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
            All posts
          </RouterLink>
          <RouterLink
            to="/sign-up"
            class="text-sm font-medium px-4 py-2 rounded-md transition-colors"
            style="background-color: var(--color-green-500); color: var(--on-accent)"
          >
            Try Checkmeup free →
          </RouterLink>
        </div>
      </div>
    </template>
  </LandingLayout>
</template>
