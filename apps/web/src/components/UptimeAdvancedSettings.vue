<script setup lang="ts">
import { computed, ref } from 'vue'
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import type { HttpMethod } from '@/api/monitors'
import {
  DEFAULT_MAX_RESPONSE_TIME_MS,
  httpMethodOptions,
  statusCodeOptions,
} from '@/lib/uptimeMonitorForm'

const httpMethod = defineModel<HttpMethod>('httpMethod', { required: true })
const acceptedStatusCodes = defineModel<number[]>('acceptedStatusCodes', { required: true })
const maxResponseTimeMs = defineModel<number>('maxResponseTimeMs', { required: true })

const open = ref(false)

// Input's modelValue is string-typed (it doesn't implement Vue's
// modelModifiers convention, so a bare `v-model.number` silently does no
// numeric conversion) — bridge it to the numeric model explicitly.
const maxResponseTimeMsInput = computed({
  get: () => maxResponseTimeMs.value.toString(),
  set: (v: string) => {
    maxResponseTimeMs.value = v === '' ? DEFAULT_MAX_RESPONSE_TIME_MS : Number(v)
  },
})

function toggleStatusCode(code: number) {
  const i = acceptedStatusCodes.value.indexOf(code)
  if (i === -1) {
    acceptedStatusCodes.value.push(code)
  } else {
    acceptedStatusCodes.value.splice(i, 1)
  }
}
</script>

<template>
  <div>
    <button
      type="button"
      class="flex items-center gap-1.5 text-sm font-medium text-[var(--text-dim)]"
      :aria-expanded="open"
      @click="open = !open"
    >
      <span
        class="inline-block text-[10px] transition-transform duration-150"
        :style="{ transform: open ? 'rotate(90deg)' : 'rotate(0deg)' }"
      >
        ▶
      </span>
      Advanced check settings
    </button>

    <div
      v-if="open"
      class="mt-2.5 space-y-4 rounded-[14px] bg-[var(--surface)] p-5"
    >
      <div>
        <Label for="httpMethod">Request method</Label>
        <select
          id="httpMethod"
          v-model="httpMethod"
          class="mt-1 w-full h-10 rounded-lg border px-3 text-sm"
        >
          <option v-for="opt in httpMethodOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </div>

      <div>
        <Label>Accepted status codes</Label>
        <div class="flex flex-wrap gap-2 mt-1.5">
          <button
            v-for="code in statusCodeOptions"
            :key="code"
            type="button"
            :aria-pressed="acceptedStatusCodes.includes(code)"
            class="px-3 py-1 rounded-full text-xs font-medium border transition-colors"
            :style="{
              borderColor: acceptedStatusCodes.includes(code) ? 'var(--accent)' : 'var(--border)',
              backgroundColor: acceptedStatusCodes.includes(code) ? 'var(--accent-wash)' : 'transparent',
              color: acceptedStatusCodes.includes(code) ? 'var(--accent)' : 'var(--text-dim)',
            }"
            @click="toggleStatusCode(code)"
          >
            {{ code }}
          </button>
        </div>
        <p class="text-xs mt-1.5 text-[var(--text-muted)]">
          A response outside this set counts as down, regardless of body content.
        </p>
      </div>

      <div>
        <Label for="maxResponseTimeMs">Request timeout</Label>
        <div class="flex items-center gap-2 mt-1">
          <Input
            id="maxResponseTimeMs"
            v-model="maxResponseTimeMsInput"
            type="number"
            min="1000"
            max="30000"
            class="w-40"
            required
          />
          <span class="text-sm text-[var(--text-muted)]">ms</span>
        </div>
        <p class="text-xs mt-1 text-[var(--text-muted)]">
          Abort and fail the check if no response arrives within this window.
        </p>
      </div>
    </div>
  </div>
</template>
