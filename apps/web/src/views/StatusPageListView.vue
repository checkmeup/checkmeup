<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import AppLayout from '@/layouts/AppLayout.vue'
import Button from '@/components/ui/Button.vue'
import type { StatusPage } from '@/api/statusPages'
import { useStatusPages } from '@/composables/useStatusPages'

const router = useRouter()
const { data, isPending: loading, error: queryError, refetch } = useStatusPages()
const pages = computed<StatusPage[]>(() => data.value ?? [])
const error = computed(() => queryError.value?.message ?? '')

function load() {
  refetch()
}
</script>

<template>
  <AppLayout>
    <div class="mx-auto max-w-[1200px] px-4 pt-8 pb-20 md:px-6">
      <div class="flex items-center justify-between mb-6">
        <h1 class="text-[30px] leading-[1.1] font-semibold tracking-[-0.035em] text-[var(--text-strong)] md:text-[34px]">Status pages</h1>
        <Button @click="router.push({ name: 'status-page-create' })">Create page</Button>
      </div>

      <div v-if="loading" class="text-sm" style="color: var(--text-muted)">Loading…</div>

      <div v-else-if="error" class="rounded-[14px] border border-[var(--border)] p-6 text-center">
        <p class="text-sm mb-4" style="color: var(--status-down)">{{ error }}</p>
        <Button variant="secondary" size="sm" @click="load">Try again</Button>
      </div>

      <div
        v-else-if="pages.length === 0"
        class="rounded-[14px] border border-[var(--border)] p-12 text-center"
      >
        <p class="text-sm mb-4" style="color: var(--text-muted)">
          No status pages yet. Create one to give your users a public health dashboard.
        </p>
        <Button @click="router.push({ name: 'status-page-create' })">Create your first page</Button>
      </div>

      <template v-else>
        <!-- Mobile cards -->
        <div class="md:hidden space-y-2">
          <div
            v-for="p in pages"
            :key="p.id"
            class="rounded-[14px] border border-[var(--border)] p-4 cursor-pointer"
            @click="router.push({ name: 'status-page-detail', params: { id: p.id } })"
          >
            <div class="flex items-center justify-between mb-1">
              <span class="font-medium text-sm" style="color: var(--text-strong)">{{ p.title }}</span>
              <a
                :href="p.publicUrl"
                target="_blank"
                rel="noopener"
                class="text-xs underline"
                style="color: var(--text-muted)"
                @click.stop
              >
                View →
              </a>
            </div>
            <span class="font-mono text-xs" style="color: var(--text-dim)">/status/{{ p.slug }}</span>
          </div>
        </div>

        <!-- Desktop table -->
        <div class="hidden md:block rounded-xl border overflow-hidden" style="border-color: var(--border)">
          <table class="w-full text-sm">
            <thead>
              <tr class="bg-[var(--surface)]">
                <th class="text-left px-4 py-2.5 font-medium text-[12.5px] text-[var(--text-dim)]">Title</th>
                <th class="text-left px-4 py-2.5 font-medium text-[12.5px] text-[var(--text-dim)]">Slug</th>
                <th class="text-left px-4 py-2.5 font-medium text-[12.5px] text-[var(--text-dim)]">Public URL</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="p in pages"
                :key="p.id"
                class="cursor-pointer border-t border-[var(--border)] transition-colors hover:bg-[var(--surface)]"
                @click="router.push({ name: 'status-page-detail', params: { id: p.id } })"
              >
                <td class="px-4 py-3 font-medium" style="color: var(--text-strong)">{{ p.title }}</td>
                <td class="px-4 py-3 font-mono text-xs" style="color: var(--text-dim)">/status/{{ p.slug }}</td>
                <td class="px-4 py-3">
                  <a
                    :href="p.publicUrl"
                    target="_blank"
                    rel="noopener"
                    class="text-xs underline"
                    style="color: var(--text-muted)"
                    @click.stop
                  >
                    View →
                  </a>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </template>
    </div>
  </AppLayout>
</template>
