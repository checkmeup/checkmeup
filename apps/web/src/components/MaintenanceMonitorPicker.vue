<script setup lang="ts">
import { computed } from 'vue'
import { useCronMonitors } from '@/composables/useCronMonitors'
import { useUptimeMonitors } from '@/composables/useUptimeMonitors'
import { useSSLMonitors } from '@/composables/useSSLMonitors'
import { useDomainMonitors } from '@/composables/useDomainMonitors'
import { usePortMonitors } from '@/composables/usePortMonitors'

interface SelectedMonitor {
  monitorType: 'cron' | 'uptime' | 'ssl' | 'domain' | 'port'
  monitorId: string
  name: string
}

const props = defineProps<{ modelValue: SelectedMonitor[] }>()
const emit = defineEmits<{ 'update:modelValue': [value: SelectedMonitor[]] }>()

const { data: cronData, isPending: cronLoading } = useCronMonitors()
const { data: uptimeData, isPending: uptimeLoading } = useUptimeMonitors()
const { data: sslData, isPending: sslLoading } = useSSLMonitors()
const { data: domainData, isPending: domainLoading } = useDomainMonitors()
const { data: portData, isPending: portLoading } = usePortMonitors()
const cronMonitors = computed(() => cronData.value ?? [])
const uptimeMonitors = computed(() => uptimeData.value ?? [])
const sslMonitors = computed(() => sslData.value ?? [])
const domainMonitors = computed(() => domainData.value ?? [])
const portMonitors = computed(() => portData.value ?? [])
const loading = computed(
  () => cronLoading.value || uptimeLoading.value || sslLoading.value || domainLoading.value || portLoading.value,
)

const typeLabel: Record<string, string> = { cron: 'Cron', uptime: 'Uptime', ssl: 'SSL', domain: 'Domain', port: 'Port' }

const allMonitors = computed(() => {
  const result: { key: string; type: 'cron' | 'uptime' | 'ssl' | 'domain' | 'port'; id: string; name: string }[] = []
  cronMonitors.value.forEach((m) => result.push({ key: `cron:${m.id}`, type: 'cron', id: m.id, name: m.name }))
  uptimeMonitors.value.forEach((m) => result.push({ key: `uptime:${m.id}`, type: 'uptime', id: m.id, name: m.name }))
  sslMonitors.value.forEach((m) => result.push({ key: `ssl:${m.id}`, type: 'ssl', id: m.id, name: m.name }))
  domainMonitors.value.forEach((m) => result.push({ key: `domain:${m.id}`, type: 'domain', id: m.id, name: m.name }))
  portMonitors.value.forEach((m) => result.push({ key: `port:${m.id}`, type: 'port', id: m.id, name: m.name }))
  return result
})

const selectedKeys = computed(() => new Set(props.modelValue.map((m) => `${m.monitorType}:${m.monitorId}`)))

function toggle(m: { key: string; type: 'cron' | 'uptime' | 'ssl' | 'domain' | 'port'; id: string; name: string }) {
  if (selectedKeys.value.has(m.key)) {
    emit(
      'update:modelValue',
      props.modelValue.filter((e) => `${e.monitorType}:${e.monitorId}` !== m.key),
    )
  } else {
    emit('update:modelValue', [...props.modelValue, { monitorType: m.type, monitorId: m.id, name: m.name }])
  }
}
</script>

<template>
  <div class="rounded-xl border" style="background-color: var(--surface); border-color: var(--border)">
    <div
      class="px-4 py-3 border-b text-sm font-medium"
      style="border-color: var(--border); color: var(--text-strong)"
    >
      Monitors ({{ modelValue.length }} selected)
    </div>
    <div v-if="loading" class="px-4 py-3 text-xs" style="color: var(--text-muted)">Loading…</div>
    <div v-else-if="allMonitors.length === 0" class="px-4 py-3 text-xs" style="color: var(--text-muted)">
      No monitors yet — create a cron, uptime, SSL, domain, or port monitor first.
    </div>
    <ul v-else class="max-h-72 overflow-y-auto">
      <li
        v-for="m in allMonitors"
        :key="m.key"
        class="flex items-center gap-3 px-4 py-2.5 cursor-pointer transition-colors border-b last:border-0"
        style="border-color: var(--border)"
        @click="toggle(m)"
      >
        <div
          class="w-4 h-4 rounded border flex-shrink-0 flex items-center justify-center text-xs font-bold"
          :style="{
            backgroundColor: selectedKeys.has(m.key) ? 'var(--accent)' : 'transparent',
            borderColor: selectedKeys.has(m.key) ? 'var(--accent)' : 'var(--border)',
            color: 'var(--on-accent)',
          }"
        >
          {{ selectedKeys.has(m.key) ? '✓' : '' }}
        </div>
        <span class="text-sm flex-1 truncate" style="color: var(--text)">{{ m.name }}</span>
        <span class="text-xs flex-shrink-0" style="color: var(--text-muted)">{{ typeLabel[m.type] }}</span>
      </li>
    </ul>
  </div>
</template>
