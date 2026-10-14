<script setup lang="ts">
// One group of settings: its name and a line of explanation on the left, the
// controls on the right; groups are divided by hairlines, not boxes. Stacks
// into one column on phones.
defineProps<{
  title: string
}>()
</script>

<template>
  <section
    class="grid gap-x-12 gap-y-4 border-t border-[var(--border)] py-10 md:grid-cols-[260px_minmax(0,1fr)]"
  >
    <div class="flex flex-col gap-1.5">
      <h2 class="text-lg font-semibold tracking-[-0.015em] text-[var(--text-strong)]">
        {{ title }}
      </h2>
      <div v-if="$slots.description" class="text-sm leading-relaxed text-[var(--text-dim)]">
        <slot name="description" />
      </div>
    </div>
    <div class="min-w-0">
      <slot />
    </div>
  </section>
</template>
