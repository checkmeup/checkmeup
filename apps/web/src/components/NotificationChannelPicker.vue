<script setup lang="ts">
import { computed } from 'vue'
import { useNotificationChannels } from '@/composables/useNotificationChannels'

const props = defineProps<{ modelValue: string[] }>()
const emit = defineEmits<{ 'update:modelValue': [value: string[]] }>()

const { data, isPending: loading } = useNotificationChannels()
const channels = computed(() => data.value ?? [])

const typeLabel: Record<string, string> = { telegram: 'Telegram', email: 'Email' }
const selected = computed(() => new Set(props.modelValue))

// Disabled channels are excluded from delivery the same way unselected ones
// are (worker.go's ListMonitorNotificationChannels only joins enabled
// channels), so "selected but all disabled" hits the same fallback as
// "nothing selected" — surface that here too, not just the empty case.
const hasEnabledSelection = computed(() =>
  channels.value.some((c) => c.enabled && selected.value.has(c.id)),
)

function toggle(id: string) {
  if (selected.value.has(id)) {
    emit('update:modelValue', props.modelValue.filter((v) => v !== id))
  } else {
    emit('update:modelValue', [...props.modelValue, id])
  }
}
</script>

<template>
  <div class="rounded-xl border" style="background-color: var(--surface); border-color: var(--border)">
    <div
      class="px-4 py-3 border-b text-sm font-medium"
      style="border-color: var(--border); color: var(--text-strong)"
    >
      Notification channels ({{ modelValue.length }} selected)
    </div>
    <div v-if="loading" class="px-4 py-3 text-xs" style="color: var(--text-muted)">Loading…</div>
    <div v-else-if="channels.length === 0" class="px-4 py-3 text-xs" style="color: var(--text-muted)">
      No channels yet — add one in
      <RouterLink to="/settings" class="underline" style="color: var(--color-green-500)">Settings</RouterLink>
      first. Until then, alerts fall back to your account email.
    </div>
    <ul v-else class="max-h-72 overflow-y-auto">
      <li
        v-for="c in channels"
        :key="c.id"
        class="flex items-center gap-3 px-4 py-2.5 cursor-pointer transition-colors border-b last:border-0"
        style="border-color: var(--border)"
        @click="toggle(c.id)"
      >
        <div
          class="w-4 h-4 rounded border flex-shrink-0 flex items-center justify-center text-xs font-bold"
          :style="{
            backgroundColor: selected.has(c.id) ? 'var(--accent)' : 'transparent',
            borderColor: selected.has(c.id) ? 'var(--accent)' : 'var(--border)',
            color: 'var(--on-accent)',
          }"
        >
          {{ selected.has(c.id) ? '✓' : '' }}
        </div>
        <span class="text-sm flex-1 truncate" style="color: var(--text)">{{ c.name }}</span>
        <span v-if="!c.enabled" class="text-xs flex-shrink-0" style="color: var(--status-down)">disabled</span>
        <span class="text-xs flex-shrink-0" style="color: var(--text-muted)">{{ typeLabel[c.type] }}</span>
      </li>
    </ul>
    <p
      v-if="channels.length > 0 && !hasEnabledSelection"
      class="px-4 py-2 text-xs border-t"
      style="border-color: var(--border); color: var(--text-muted)"
    >
      {{
        modelValue.length === 0
          ? 'No channels selected'
          : 'All selected channels are disabled'
      }}
      — alerts will fall back to your account email.
    </p>
  </div>
</template>
