<script setup lang="ts">
import Button from '@/components/ui/Button.vue'
import { useConsent } from '@/lib/consent'
import { loadGtm } from '@/lib/analytics'

const { status, grant, deny } = useConsent()

function accept() {
  grant()
  loadGtm()
}
</script>

<template>
  <div
    v-if="status === undefined"
    class="fixed inset-x-0 bottom-0 z-50 flex flex-col gap-3 border-t border-[var(--border)] bg-[var(--surface-raised)] p-4 text-sm sm:flex-row sm:items-center sm:justify-between"
  >
    <p class="text-[var(--text-dim)]">
      We use cookies for basic web analytics (page views, traffic sources). No ads, no cross-site tracking.
    </p>
    <div class="flex shrink-0 gap-2">
      <Button variant="secondary" size="sm" @click="deny">Decline</Button>
      <Button size="sm" @click="accept">Accept</Button>
    </div>
  </div>
</template>
