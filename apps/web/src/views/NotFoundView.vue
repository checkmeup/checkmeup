<script setup lang="ts">
import { useRoute } from 'vue-router'
import { useHead } from '@unhead/vue'
import LandingLayout from '@/layouts/LandingLayout.vue'
import NotFoundHero from '@/components/NotFoundHero.vue'
import { useSeo } from '@/composables/useSeo'

const route = useRoute()

useSeo({
  title: 'Page not found — Checkmeup',
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
    <NotFoundHero
      badge="Monitor status: not found"
      heading="This page went down and stayed down."
      description="The page you're looking for doesn't exist, moved, or never came back up. Unlike your client's site, we won't be paging you about it."
      :primary-cta="{ label: 'Back to dashboard', to: '/dashboard' }"
      :secondary-cta="{ label: 'Go to homepage', to: '/' }"
    />
  </LandingLayout>
</template>
