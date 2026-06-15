<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import { useAuthStore } from '@/stores/auth'
import { monitorsApi } from '@/api/monitors'

const auth = useAuthStore()
const router = useRouter()

const cronCount = ref<number | null>(null)

onMounted(async () => {
  try {
    const monitors = await monitorsApi.listCron()
    cronCount.value = monitors.length
  } catch {
    cronCount.value = 0
  }
})
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

      <div class="grid gap-4 sm:grid-cols-3">
        <!-- Cron monitors -->
        <div
          class="rounded-xl border p-6 space-y-4 cursor-pointer transition-colors"
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

        <!-- Uptime monitors (Phase 3 — not yet built) -->
        <div
          class="rounded-xl border p-6 space-y-4"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div class="flex items-center justify-between">
            <span class="text-sm font-medium" style="color: var(--text-dim)">Uptime monitors</span>
            <span class="text-2xl font-bold" style="color: var(--text-strong)">0</span>
          </div>
          <p class="text-xs" style="color: var(--text-muted)">
            Ping your URLs and detect downtime in seconds.
          </p>
          <Button variant="secondary" size="sm" class="w-full" disabled>
            Add uptime monitor
          </Button>
        </div>

        <!-- SSL monitors (Phase 4 — not yet built) -->
        <div
          class="rounded-xl border p-6 space-y-4"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div class="flex items-center justify-between">
            <span class="text-sm font-medium" style="color: var(--text-dim)">SSL monitors</span>
            <span class="text-2xl font-bold" style="color: var(--text-strong)">0</span>
          </div>
          <p class="text-xs" style="color: var(--text-muted)">
            Know before your certificates expire.
          </p>
          <Button variant="secondary" size="sm" class="w-full" disabled>
            Add SSL monitor
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
            Add a cron monitor and copy the ping URL
          </li>
          <li class="flex items-start gap-2">
            <span class="font-mono text-xs px-1.5 py-0.5 rounded" style="background-color: var(--surface-raised); color: var(--text-muted)">2</span>
            Call the ping URL at the end of your cron job
          </li>
          <li class="flex items-start gap-2">
            <span class="font-mono text-xs px-1.5 py-0.5 rounded" style="background-color: var(--surface-raised); color: var(--text-muted)">3</span>
            Connect Telegram to receive alerts when a job misses
          </li>
        </ol>
      </div>
    </div>
  </AppLayout>
</template>
