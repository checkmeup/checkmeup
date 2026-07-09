<script setup lang="ts">
import { useRouter } from 'vue-router'

interface AttentionItem {
  key: string
  title: string
  detail: string
  actionLabel: string
  severity: 0 | 1
  color: string
  wash: string
  routeName: string
  id: string
}

defineProps<{
  items: AttentionItem[]
}>()

const router = useRouter()
</script>

<template>
  <div v-if="items.length > 0" class="mb-7">
    <div
      class="text-[13px] font-semibold mb-3 flex items-center gap-2"
      style="color: var(--text-strong)"
    >
      <svg
        width="14"
        height="14"
        viewBox="0 0 24 24"
        fill="none"
        stroke="var(--status-degraded)"
        stroke-width="2.5"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path
          d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"
        />
        <line x1="12" y1="9" x2="12" y2="13" />
        <line x1="12" y1="17" x2="12.01" y2="17" />
      </svg>
      Needs attention
    </div>
    <div class="grid gap-3" style="grid-template-columns: repeat(auto-fit, minmax(280px, 1fr))">
      <div
        v-for="item in items"
        :key="item.key"
        class="rounded-xl border p-4 flex flex-col gap-2 cursor-pointer"
        :style="{
          borderColor: `color-mix(in srgb, ${item.color} 30%, transparent)`,
          backgroundColor: item.wash,
        }"
        @click="router.push({ name: item.routeName, params: { id: item.id } })"
      >
        <div class="text-[13.5px] font-semibold" style="color: var(--text-strong)">
          {{ item.title }}
        </div>
        <div class="text-xs leading-relaxed font-mono" style="color: var(--text-dim)">
          {{ item.detail }}
        </div>
        <span class="text-xs font-semibold mt-0.5" :style="{ color: item.color }"
          >{{ item.actionLabel }} →</span
        >
      </div>
    </div>
  </div>
</template>
