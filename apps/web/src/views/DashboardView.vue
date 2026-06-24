<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import { useAuthStore } from '@/stores/auth'
import { useCronMonitors } from '@/composables/useCronMonitors'
import { useUptimeMonitors } from '@/composables/useUptimeMonitors'
import { useSSLMonitors } from '@/composables/useSSLMonitors'
import { useDomainMonitors } from '@/composables/useDomainMonitors'
import { useStatusPages } from '@/composables/useStatusPages'

const auth = useAuthStore()
const router = useRouter()

// Per-card counts fail independently (each query has its own error state),
// matching the original Promise.allSettled behavior — one slow/broken
// resource doesn't block the others from showing their counts.
const { data: cronData } = useCronMonitors()
const { data: uptimeData } = useUptimeMonitors()
const { data: sslData } = useSSLMonitors()
const { data: domainData } = useDomainMonitors()
const { data: statusPageData } = useStatusPages()

const cronCount = computed(() => cronData.value?.length ?? null)
const uptimeCount = computed(() => uptimeData.value?.length ?? null)
const sslCount = computed(() => sslData.value?.length ?? null)
const domainCount = computed(() => domainData.value?.length ?? null)
const statusPageCount = computed(() => statusPageData.value?.length ?? null)
</script>

<template>
  <AppLayout>
    <div class="p-8 max-w-4xl mx-auto">
      <div class="mb-8">
        <h1 class="text-2xl font-semibold" style="color: var(--text-strong)">
          Dashboard
        </h1>
        <p class="mt-1 text-sm" style="color: var(--text-muted)">
          Welcome back{{ auth.user?.email ? `, ${auth.user.email}` : '' }}.
        </p>
      </div>

      <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <!-- Cron monitors -->
        <div
          class="rounded-xl border p-6 space-y-4 cursor-pointer transition-colors hover:border-[var(--color-green-700)]"
          style="background-color: var(--surface); border-color: var(--border)"
          @click="router.push({ name: 'cron-monitors' })"
        >
          <div class="flex items-center justify-between">
            <span class="text-sm font-medium" style="color: var(--text-dim)">Cron monitors</span>
            <span class="text-2xl font-bold" style="color: var(--text-strong)">
              {{ cronCount ?? '—' }}
            </span>
          </div>
          <p class="text-xs" style="color: var(--text-muted)">
            Get alerted when a scheduled job stops running.
          </p>
          <Button
            variant="secondary"
            size="sm"
            class="w-full"
            @click.stop="router.push({ name: 'cron-monitor-create' })"
          >
            Add cron monitor
          </Button>
        </div>

        <!-- Uptime monitors -->
        <div
          class="rounded-xl border p-6 space-y-4 cursor-pointer transition-colors hover:border-[var(--color-green-700)]"
          style="background-color: var(--surface); border-color: var(--border)"
          @click="router.push({ name: 'uptime-monitors' })"
        >
          <div class="flex items-center justify-between">
            <span class="text-sm font-medium" style="color: var(--text-dim)">Uptime monitors</span>
            <span class="text-2xl font-bold" style="color: var(--text-strong)">
              {{ uptimeCount ?? '—' }}
            </span>
          </div>
          <p class="text-xs" style="color: var(--text-muted)">
            Ping your URLs and detect downtime in seconds.
          </p>
          <Button
            variant="secondary"
            size="sm"
            class="w-full"
            @click.stop="router.push({ name: 'uptime-monitor-create' })"
          >
            Add uptime monitor
          </Button>
        </div>

        <!-- SSL monitors -->
        <div
          class="rounded-xl border p-6 space-y-4 cursor-pointer transition-colors hover:border-[var(--color-green-700)]"
          style="background-color: var(--surface); border-color: var(--border)"
          @click="router.push({ name: 'ssl-monitors' })"
        >
          <div class="flex items-center justify-between">
            <span class="text-sm font-medium" style="color: var(--text-dim)">SSL monitors</span>
            <span class="text-2xl font-bold" style="color: var(--text-strong)">
              {{ sslCount ?? '—' }}
            </span>
          </div>
          <p class="text-xs" style="color: var(--text-muted)">
            Know before your certificates expire.
          </p>
          <Button
            variant="secondary"
            size="sm"
            class="w-full"
            @click.stop="router.push({ name: 'ssl-monitor-create' })"
          >
            Add SSL monitor
          </Button>
        </div>

        <!-- Domain monitors -->
        <div
          class="rounded-xl border p-6 space-y-4 cursor-pointer transition-colors hover:border-[var(--color-green-700)]"
          style="background-color: var(--surface); border-color: var(--border)"
          @click="router.push({ name: 'domain-monitors' })"
        >
          <div class="flex items-center justify-between">
            <span class="text-sm font-medium" style="color: var(--text-dim)">Domain monitors</span>
            <span class="text-2xl font-bold" style="color: var(--text-strong)">
              {{ domainCount ?? '—' }}
            </span>
          </div>
          <p class="text-xs" style="color: var(--text-muted)">
            Know before your domain registration lapses.
          </p>
          <Button
            variant="secondary"
            size="sm"
            class="w-full"
            @click.stop="router.push({ name: 'domain-monitor-create' })"
          >
            Add domain monitor
          </Button>
        </div>

        <!-- Status pages -->
        <div
          class="rounded-xl border p-6 space-y-4 cursor-pointer transition-colors hover:border-[var(--color-green-700)]"
          style="background-color: var(--surface); border-color: var(--border)"
          @click="router.push({ name: 'status-pages' })"
        >
          <div class="flex items-center justify-between">
            <span class="text-sm font-medium" style="color: var(--text-dim)">Status pages</span>
            <span class="text-2xl font-bold" style="color: var(--text-strong)">
              {{ statusPageCount ?? '—' }}
            </span>
          </div>
          <p class="text-xs" style="color: var(--text-muted)">
            Public dashboards your users can bookmark.
          </p>
          <Button
            variant="secondary"
            size="sm"
            class="w-full"
            @click.stop="router.push({ name: 'status-page-create' })"
          >
            Create status page
          </Button>
        </div>
      </div>

      <div
        class="mt-8 rounded-xl border p-6"
        style="background-color: var(--surface); border-color: var(--border)"
      >
        <h2 class="font-medium mb-3" style="color: var(--text-strong)">Getting started</h2>
        <ol class="space-y-2 text-sm" style="color: var(--text-dim)">
          <li class="flex items-start gap-2">
            <span class="font-mono text-xs px-1.5 py-0.5 rounded" style="background-color: var(--surface-raised); color: var(--text-muted)">1</span>
            Add a monitor — cron, uptime, SSL, or domain expiry
          </li>
          <li class="flex items-start gap-2">
            <span class="font-mono text-xs px-1.5 py-0.5 rounded" style="background-color: var(--surface-raised); color: var(--text-muted)">2</span>
            Set up an alert channel — Telegram, email, or webhook
          </li>
          <li class="flex items-start gap-2">
            <span class="font-mono text-xs px-1.5 py-0.5 rounded" style="background-color: var(--surface-raised); color: var(--text-muted)">3</span>
            Create a status page to share uptime with your clients
          </li>
        </ol>
      </div>
    </div>
  </AppLayout>
</template>
