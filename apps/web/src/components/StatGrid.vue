<script setup lang="ts">
// A row of key numbers: big mono value, plain-words label under it.
defineProps<{
  items: { label: string; value: string; color?: string }[]
}>()
</script>

<template>
  <dl class="grid grid-cols-[repeat(auto-fit,minmax(150px,1fr))] gap-6">
    <div v-for="item in items" :key="item.label" class="flex flex-col-reverse gap-0.5">
      <dt class="text-sm text-[var(--text-dim)]">{{ item.label }}</dt>
      <dd
        class="font-mono text-[30px] font-medium tracking-[-0.04em] break-words"
        :style="{ color: item.color ?? 'var(--text-strong)' }"
      >
        {{ item.value }}
      </dd>
    </div>
  </dl>
</template>
