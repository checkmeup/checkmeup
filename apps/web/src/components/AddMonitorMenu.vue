<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import Button from '@/components/ui/Button.vue'
import { type MonitorType, createRouteName, monitorTypes, typeLabel } from '@/lib/monitorRows'

const router = useRouter()
const open = ref(false)

const descriptions: Record<MonitorType, string> = {
  cron: 'Alert on missed scheduled jobs',
  uptime: 'Ping a URL and detect downtime',
  ssl: 'Track certificate expiry',
  domain: 'Track registration expiry',
  port: 'Raw TCP connect checks',
  dns: 'Alert on an unexpected record change',
}

function choose(type: MonitorType) {
  open.value = false
  router.push({ name: createRouteName[type] })
}
</script>

<template>
  <div class="relative flex-shrink-0">
    <Button :aria-expanded="open" aria-haspopup="menu" @click="open = !open">
      New monitor
    </Button>
    <template v-if="open">
      <div class="fixed inset-0 z-40" @click="open = false"></div>
      <div
        role="menu"
        class="absolute top-[calc(100%+8px)] right-0 z-50 w-[264px] overflow-hidden rounded-xl border border-[var(--border)] bg-[var(--bg)] py-1.5 shadow-[0_16px_40px_rgb(0_0_0/18%)]"
      >
        <button
          v-for="type in monitorTypes"
          :key="type"
          type="button"
          role="menuitem"
          class="block w-full px-4 py-2.5 text-left transition-colors hover:bg-[var(--surface-raised)] focus-visible:bg-[var(--surface-raised)] focus-visible:outline-none"
          @click="choose(type)"
        >
          <span class="block text-sm font-medium text-[var(--text-strong)]">
            {{ typeLabel[type] }} monitor
          </span>
          <span class="mt-0.5 block text-[13px] text-[var(--text-dim)]">{{
            descriptions[type]
          }}</span>
        </button>
      </div>
    </template>
  </div>
</template>
