<script setup lang="ts">
import LandingLayout from '@/layouts/LandingLayout.vue'
import { faqCategories } from '@/faq/faqs'
</script>

<template>
  <LandingLayout>
    <!-- Hero -->
    <section class="max-w-6xl mx-auto px-4 sm:px-6 pt-16 pb-10 sm:pt-24 sm:pb-12">
      <h1 class="text-4xl sm:text-5xl font-bold tracking-tight mb-4" style="color: var(--text-strong)">
        Frequently asked questions
      </h1>
      <p class="text-lg max-w-2xl" style="color: var(--text-dim)">
        Everything in one place — getting started, billing, monitors and alerts, status pages, and privacy.
      </p>
    </section>

    <div class="max-w-6xl mx-auto px-4 sm:px-6 pb-24 flex flex-col lg:flex-row gap-12">
      <!-- Sidebar nav -->
      <nav class="lg:w-56 flex-shrink-0">
        <div class="lg:sticky lg:top-24 flex lg:flex-col gap-1 overflow-x-auto pb-2 lg:pb-0 -mx-1 px-1">
          <a
            v-for="c in faqCategories"
            :key="c.id"
            :href="`#${c.id}`"
            class="text-sm px-3 py-1.5 rounded-md whitespace-nowrap transition-colors hover:underline"
            style="color: var(--text-dim)"
          >
            {{ c.label }}
          </a>
        </div>
      </nav>

      <!-- Content -->
      <div class="flex-1 min-w-0 space-y-16">
        <section v-for="c in faqCategories" :key="c.id" :id="c.id" class="scroll-mt-24">
          <h2 class="text-2xl font-bold mb-4" style="color: var(--text-strong)">{{ c.label }}</h2>
          <div class="space-y-0 divide-y" style="border-color: var(--border)">
            <div v-for="faq in c.entries" :key="faq.q" class="py-5" style="border-color: var(--border)">
              <h3 class="text-sm font-semibold mb-2" style="color: var(--text-strong)">{{ faq.q }}</h3>
              <p class="text-sm leading-relaxed" style="color: var(--text-dim)">{{ faq.a }}</p>
            </div>
          </div>
        </section>
      </div>
    </div>
  </LandingLayout>
</template>
