<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink } from 'vue-router'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import { trackToolCheck } from '@/lib/analytics'
import { runToolCheck } from '@/lib/toolCheck'

// The form and result of the free DNS lookup (PRD free-tools): one hostname,
// every record type a DNS monitor checks, and under them an offer to monitor
// them.

interface RecordSet {
  type: string
  values: string[]
  error?: string
}

interface DNSResult {
  host: string
  records: RecordSet[]
}

const value = ref('')
const checking = ref(false)
const result = ref<DNSResult | null>(null)
const error = ref('')

async function check() {
  if (!value.value.trim() || checking.value) return
  checking.value = true
  error.value = ''
  result.value = null
  const outcome = await runToolCheck<DNSResult>('/api/v1/tools/dns', 'host', value.value)
  if (outcome.ok) result.value = outcome.data
  else error.value = outcome.message
  trackToolCheck('dns', outcome.ok)
  checking.value = false
}
</script>

<template>
  <div class="rounded-[14px] border border-[var(--border)] bg-[var(--surface)] p-5 sm:p-6">
    <form class="flex flex-col gap-3 sm:flex-row sm:items-end" @submit.prevent="check">
      <label class="flex flex-1 flex-col gap-2 text-sm font-medium text-[var(--text-strong)]">
        Hostname
        <Input v-model="value" placeholder="example.com" autocomplete="off" required />
      </label>
      <Button type="submit" :disabled="checking">{{ checking ? 'Checking…' : 'Look up' }}</Button>
    </form>

    <p v-if="error" role="alert" class="mt-5 text-sm text-[var(--status-down)]">{{ error }}</p>

    <div v-if="result" class="mt-6 border-t border-[var(--border)] pt-5">
      <table class="w-full text-sm">
        <tbody>
          <tr v-for="set in result.records" :key="set.type" class="border-b border-[var(--border)] last:border-0">
            <td class="w-16 py-2 pr-4 align-top font-mono font-semibold text-[var(--text-strong)]">{{ set.type }}</td>
            <td class="py-2 align-top font-mono">
              <span v-if="set.error" class="text-[var(--status-down)]">{{ set.error }}</span>
              <span v-else-if="!set.values.length" class="text-[var(--text-dim)]">No records</span>
              <div v-for="v in set.values" v-else :key="v" class="break-all text-[var(--text-strong)]">{{ v }}</div>
            </td>
          </tr>
        </tbody>
      </table>
      <div class="mt-6 flex flex-wrap items-center gap-x-4 gap-y-2 rounded-[10px] bg-[var(--surface-raised)] p-4">
        <p class="flex-1 text-sm text-[var(--text-dim)]">
          Get an alert when one of these records changes or stops resolving.
        </p>
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
