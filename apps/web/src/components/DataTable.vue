<script setup lang="ts" generic="Row">
// The one table style for logs and lists: a bordered box, a quiet header row,
// hairline rules. Each column's cell is a `#<key>` slot, falling back to
// `row[key]` as text.
defineProps<{
  columns: { key: string; label: string; class?: string }[]
  rows: Row[]
  rowKey: (row: Row) => string
  empty: string
}>()
</script>

<template>
  <p v-if="rows.length === 0" class="border-t border-[var(--border)] py-4 text-sm text-[var(--text-dim)]">
    {{ empty }}
  </p>
  <div v-else class="overflow-x-auto rounded-[14px] border border-[var(--border)]">
    <table class="w-full min-w-[560px] border-collapse text-sm">
      <thead>
        <tr class="bg-[var(--surface)] text-left text-[12.5px] text-[var(--text-dim)]">
          <th
            v-for="col in columns"
            :key="col.key"
            class="px-4 py-2.5 font-medium first:pl-4"
            :class="col.class"
          >
            {{ col.label }}
          </th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="row in rows" :key="rowKey(row)" class="border-t border-[var(--border)]">
          <td
            v-for="col in columns"
            :key="col.key"
            class="px-4 py-2.5 text-[var(--text)]"
            :class="col.class"
          >
            <slot :name="col.key" :row="row">{{ (row as Record<string, unknown>)[col.key] }}</slot>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
