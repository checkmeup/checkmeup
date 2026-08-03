<script setup lang="ts">
import Input from '@/components/ui/Input.vue'
import Label from '@/components/ui/Label.vue'
import type { JsonAssertion } from '@/api/monitors'
import { comparatorOptions } from '@/lib/uptimeMonitorForm'

const assertions = defineModel<JsonAssertion[]>({ required: true })

function addAssertion() {
  assertions.value.push({ path: '', comparator: 'equals', expected: '' })
}

function removeAssertion(i: number) {
  assertions.value.splice(i, 1)
}
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-2">
      <Label>JSON assertions (optional)</Label>
      <button
        type="button"
        class="text-xs px-2 py-1 rounded"
        style="color: var(--text-dim); background-color: var(--surface-raised)"
        @click="addAssertion"
      >
        + Add
      </button>
    </div>
    <div v-if="assertions.length === 0" class="text-xs" style="color: var(--text-muted)">
      Assert on JSON response fields, e.g. <code>data.status</code> equals <code>ok</code>.
    </div>
    <div v-for="(a, i) in assertions" :key="i" class="flex items-center gap-2 mt-2">
      <Input v-model="a.path" placeholder="$.status" class="flex-1 min-w-0" />
      <select
        v-model="a.comparator"
        class="rounded-md border px-2 py-2 text-sm"
        style="background-color: var(--surface-raised); border-color: var(--border); color: var(--text)"
      >
        <option v-for="opt in comparatorOptions" :key="opt.value" :value="opt.value">
          {{ opt.label }}
        </option>
      </select>
      <Input v-model="a.expected" placeholder="ok" class="flex-1 min-w-0" />
      <button
        type="button"
        class="text-xs px-1.5 py-1 rounded flex-shrink-0"
        style="color: var(--text-muted); background-color: var(--surface-raised)"
        @click="removeAssertion(i)"
      >
        ✕
      </button>
    </div>
  </div>
</template>
