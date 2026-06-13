<script setup lang="ts">
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()

const monitors = [
  {
    title: 'Cron monitors',
    description: 'Get alerted when a scheduled job stops running.',
    cta: 'Add cron monitor',
  },
  {
    title: 'Uptime monitors',
    description: 'Ping your URLs and detect downtime in seconds.',
    cta: 'Add uptime monitor',
  },
  {
    title: 'SSL monitors',
    description: 'Know before your certificates expire.',
    cta: 'Add SSL monitor',
  },
]
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
        <div
          v-for="monitor in monitors"
          :key="monitor.title"
          class="rounded-xl border p-6 space-y-4"
          style="background-color: var(--surface); border-color: var(--border)"
        >
          <div class="flex items-center justify-between">
            <span class="text-sm font-medium" style="color: var(--text-dim)">{{ monitor.title }}</span>
            <span class="text-2xl font-bold" style="color: var(--text-strong)">0</span>
          </div>
          <p class="text-xs" style="color: var(--text-muted)">{{ monitor.description }}</p>
          <Button variant="secondary" size="sm" class="w-full" disabled>
            {{ monitor.cta }}
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
