<script setup lang="ts">
import { computed } from 'vue'
import { type Tone, toneColor } from '@/lib/monitorRows'

const props = defineProps<{
  tone: Tone
  label: string
}>()

// An "up" label stays quiet (dim text, green dot); every other state shows
// its color in the text too, so problems stand out in a long list.
const textColor = computed(() => (props.tone === 'up' ? 'var(--text-dim)' : toneColor[props.tone]))
</script>

<template>
  <span
    class="inline-flex items-center gap-2 text-[13px] font-medium whitespace-nowrap"
    :style="{ color: textColor }"
  >
    <span
      class="h-2 w-2 flex-shrink-0 rounded-full"
      :style="{ backgroundColor: toneColor[tone] }"
    ></span>
    {{ label }}
  </span>
</template>
