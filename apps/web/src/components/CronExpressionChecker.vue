<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import { trackToolCheck } from '@/lib/analytics'
import { runToolCheck } from '@/lib/toolCheck'
import { commonCronExpressions } from '@/content/cronExpression'

// The input and result of the free cron explainer (PRD free-tools): the API
// reads the expression with a cron monitor's own parser, so the result
// follows the input as the visitor types, a moment after they stop.

interface CronResult {
  expression: string
  description: string
  runs: string[]
}

const DEBOUNCE_MS = 400

const value = ref(commonCronExpressions[0])
const result = ref<CronResult | null>(null)
const error = ref('')
const localZone = ref('')
let timer: ReturnType<typeof setTimeout> | undefined
let latest = 0
let tracked = false

async function explain() {
  const expression = value.value.trim()
  if (!expression) {
    result.value = null
    error.value = ''
    return
  }
  const request = ++latest
  const outcome = await runToolCheck<CronResult>('/api/v1/tools/cron', 'expression', expression)
  // An answer to an older keystroke arriving late doesn't replace a newer one.
  if (request !== latest) return
  if (outcome.ok) {
    result.value = outcome.data
    error.value = ''
  } else {
    result.value = null
    error.value = outcome.message
  }
  // Once per page view: every pause in typing is a request, and GA4 should
  // count visitors who used the tool, not their keystrokes.
  if (!tracked) {
    trackToolCheck('cron', outcome.ok)
    tracked = true
  }
}

function pick(expression: string) {
  clearTimeout(timer)
  value.value = expression
  explain()
}

watch(value, () => {
  clearTimeout(timer)
  timer = setTimeout(explain, DEBOUNCE_MS)
})

onMounted(() => {
  localZone.value = Intl.DateTimeFormat().resolvedOptions().timeZone
  explain()
})

onBeforeUnmount(() => clearTimeout(timer))

const runFormat: Intl.DateTimeFormatOptions = {
  weekday: 'short',
  month: 'short',
  day: 'numeric',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
  hourCycle: 'h23',
}

function inUTC(run: string) {
  return new Date(run).toLocaleString('en-US', { ...runFormat, timeZone: 'UTC' })
}

function inLocal(run: string) {
  return new Date(run).toLocaleString('en-US', runFormat)
}
</script>

<template>
  <div class="rounded-[14px] border border-[var(--border)] bg-[var(--surface)] p-5 sm:p-6">
    <label class="flex flex-col gap-2 text-sm font-medium text-[var(--text-strong)]">
      Cron expression
      <Input v-model="value" placeholder="*/5 * * * *" autocomplete="off" class="font-mono" />
    </label>

    <div class="mt-3 flex flex-wrap gap-2">
      <Button
        v-for="expression in commonCronExpressions"
        :key="expression"
        type="button"
        variant="secondary"
        size="sm"
        class="font-mono"
        @click="pick(expression)"
      >
        {{ expression }}
      </Button>
    </div>

    <p v-if="error" role="alert" class="mt-5 text-sm text-[var(--status-down)]">{{ error }}</p>

    <div v-if="result" class="mt-6 border-t border-[var(--border)] pt-5">
      <p class="text-[20px] font-semibold tracking-[-0.01em] text-[var(--text-strong)]" data-test="description">
        {{ result.description }}
      </p>
      <table class="mt-4 w-full text-sm">
        <thead>
          <tr class="text-left text-[var(--text-dim)]">
            <th class="pb-2 pr-4 font-medium">Next runs (UTC)</th>
            <th class="pb-2 font-medium">Your time<span v-if="localZone"> ({{ localZone }})</span></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="run in result.runs" :key="run" class="border-t border-[var(--border)]">
            <td class="py-2 pr-4 font-mono text-[var(--text-strong)]">{{ inUTC(run) }}</td>
            <td class="py-2 font-mono text-[var(--text)]">{{ inLocal(run) }}</td>
          </tr>
        </tbody>
      </table>
      <div class="mt-6 flex flex-wrap items-center gap-x-4 gap-y-2 rounded-[10px] bg-[var(--surface-raised)] p-4">
        <p class="flex-1 text-sm text-[var(--text-dim)]">
          Get an alert when a job on this schedule doesn't run.
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
