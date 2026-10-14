<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'
import Button from '@/components/ui/Button.vue'
import StatusLabel from '@/components/StatusLabel.vue'
import type { Tone } from '@/lib/monitorRows'

// The top of every monitor page: where it sits, what state it's in, its name,
// a plain-words summary, and the Edit / Pause / Delete actions.
defineProps<{
  backTo: RouteLocationRaw
  backLabel: string
  name: string
  tone: Tone
  statusLabel: string
  paused: boolean
}>()

const confirmDelete = defineModel<boolean>('confirmDelete', { default: false })

const emit = defineEmits<{
  edit: []
  togglePause: []
  delete: []
}>()
</script>

<template>
  <div class="flex flex-col gap-5">
    <RouterLink
      :to="backTo"
      class="self-start py-1 text-sm text-[var(--text-dim)] hover:text-[var(--text)]"
    >
      ← {{ backLabel }}
    </RouterLink>
    <div class="flex flex-wrap items-start justify-between gap-6">
      <div class="flex max-w-[720px] min-w-0 flex-col gap-2">
        <StatusLabel :tone="tone" :label="statusLabel" />
        <h1
          class="text-[34px] leading-[1.08] font-semibold tracking-[-0.035em] break-words text-[var(--text-strong)] md:text-[44px]"
        >
          {{ name }}
        </h1>
        <slot />
      </div>
      <div class="flex flex-wrap gap-2">
        <slot name="actions" />
        <Button variant="secondary" @click="emit('edit')">Edit</Button>
        <Button variant="secondary" @click="emit('togglePause')">
          {{ paused ? 'Resume' : 'Pause' }}
        </Button>
        <Button
          v-if="!confirmDelete"
          variant="secondary"
          class="text-[var(--status-down)]"
          @click="confirmDelete = true"
        >
          Delete
        </Button>
        <template v-else>
          <Button variant="destructive" @click="emit('delete')">Confirm delete</Button>
          <Button variant="secondary" @click="confirmDelete = false">Cancel</Button>
        </template>
      </div>
    </div>
  </div>
</template>
