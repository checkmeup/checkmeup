<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import StatusLabel from '@/components/StatusLabel.vue'
import { trackToolCheck } from '@/lib/analytics'
import { expiryStatus, runToolCheck } from '@/lib/toolCheck'

// The form and result of a free checker (PRD free-tools): one input, one
// live check, and under the result an offer to monitor it.
const props = defineProps<{
  tool: string
  endpoint: string
  param: string
  inputLabel: string
  placeholder: string
  detailKey: string
  detailLabel: string
  monitorCta: string
}>()

interface ExpiryResult {
  expiresAt: string
  daysLeft: number
  [key: string]: unknown
}

const value = ref('')
const checking = ref(false)
const result = ref<ExpiryResult | null>(null)
const error = ref('')

async function check() {
  if (!value.value.trim() || checking.value) return
  checking.value = true
  error.value = ''
  result.value = null
  const outcome = await runToolCheck<ExpiryResult>(props.endpoint, props.param, value.value)
  if (outcome.ok) result.value = outcome.data
  else error.value = outcome.message
  trackToolCheck(props.tool, outcome.ok)
  checking.value = false
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric', timeZone: 'UTC' })
}
</script>

<template>
  <div class="rounded-[14px] border border-[var(--border)] bg-[var(--surface)] p-5 sm:p-6">
    <form class="flex flex-col gap-3 sm:flex-row sm:items-end" @submit.prevent="check">
      <label class="flex flex-1 flex-col gap-2 text-sm font-medium text-[var(--text-strong)]">
        {{ inputLabel }}
        <Input v-model="value" :placeholder="placeholder" autocomplete="off" required />
      </label>
      <Button type="submit" :disabled="checking">{{ checking ? 'Checking…' : 'Check' }}</Button>
    </form>

    <p v-if="error" role="alert" class="mt-5 text-sm text-[var(--status-down)]">{{ error }}</p>

    <div v-if="result" class="mt-6 border-t border-[var(--border)] pt-5">
      <StatusLabel v-bind="expiryStatus(result.daysLeft)" />
      <dl class="mt-4 grid grid-cols-[auto_1fr] gap-x-6 gap-y-2 text-sm">
        <dt class="text-[var(--text-dim)]">Expires</dt>
        <dd class="text-[var(--text-strong)]">{{ formatDate(result.expiresAt) }}</dd>
        <dt class="text-[var(--text-dim)]">Days left</dt>
        <dd class="text-[var(--text-strong)]">{{ result.daysLeft }}</dd>
        <dt class="text-[var(--text-dim)]">{{ detailLabel }}</dt>
        <dd class="text-[var(--text-strong)]">{{ result[detailKey] || '—' }}</dd>
      </dl>
      <div class="mt-6 flex flex-wrap items-center gap-x-4 gap-y-2 rounded-[10px] bg-[var(--surface-raised)] p-4">
        <p class="flex-1 text-sm text-[var(--text-dim)]">{{ monitorCta }}</p>
        <RouterLink
          to="/sign-up"
          class="rounded-[10px] bg-[var(--accent)] px-4 py-2 text-sm font-semibold text-[var(--on-accent)] transition-colors hover:bg-[var(--accent-hover)]"
        >
          Monitor it free →
        </RouterLink>
      </div>
    </div>
  </div>
</template>
