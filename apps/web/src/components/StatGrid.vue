<script setup lang="ts">
// A row of key numbers: big mono value, plain-words label under it. A long
// text value (a registrar or issuer name) drops to body size so it doesn't
// wrap into a tower of giant words.
defineProps<{
  items: { label: string; value: string; color?: string }[]
}>()

const longValue = 16
</script>

<template>
  <dl class="grid grid-cols-[repeat(auto-fit,minmax(150px,1fr))] gap-6">
    <div v-for="item in items" :key="item.label" class="flex flex-col-reverse justify-end gap-0.5">
      <dt class="text-sm text-[var(--text-dim)]">{{ item.label }}</dt>
      <dd
        class="break-words"
        :class="
          item.value.length > longValue
            ? 'pt-2 text-lg leading-snug font-medium tracking-[-0.01em]'
            : 'font-mono text-[30px] font-medium tracking-[-0.04em]'
        "
        :style="{ color: item.color ?? 'var(--text-strong)' }"
      >
        {{ item.value }}
      </dd>
    </div>
  </dl>
</template>
