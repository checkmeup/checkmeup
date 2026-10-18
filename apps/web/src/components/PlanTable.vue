<script setup lang="ts">
import { RouterLink } from 'vue-router'
import IncludedIcon from '@/components/IncludedIcon.vue'
import { plans, planRows } from '@/content/home'

// The plan comparison: one column per plan, one row per thing that differs,
// with a sign-up button under each plan. The landing page and the monitor
// type pages show the same table.
</script>

<template>
  <div class="overflow-x-auto rounded-[14px] border border-[var(--border)]">
    <!-- Fixed layout so the four plan columns are equal, whatever their
         descriptions' lengths; the label column takes a set share. On
         phones the label column collapses and each label gets its own
         full-width line above its values, so the plan columns get the
         whole width ("Enterprise" overflows a quarter of what's left). -->
    <table class="w-full table-fixed border-collapse text-left text-[13px] sm:text-sm">
      <colgroup>
        <col class="w-0 sm:w-[24%]" />
        <col v-for="plan in plans" :key="plan.name" />
      </colgroup>
      <thead>
        <tr class="align-bottom">
          <th scope="col" class="px-2 py-4 max-sm:hidden sm:px-5">
            <span class="sr-only">Plan</span>
          </th>
          <th
            v-for="plan in plans"
            :key="plan.name"
            scope="col"
            class="px-2 py-4 font-normal sm:px-5"
            :class="plan.highlight ? 'bg-[color-mix(in_srgb,var(--accent)_6%,var(--bg))]' : ''"
          >
            <span
              v-if="plan.highlight"
              class="mb-1 block text-xs font-semibold text-[var(--accent)]"
              >Most popular</span
            >
            <span class="block text-xs font-semibold text-[var(--text-strong)] sm:text-sm">{{
              plan.name
            }}</span>
            <span
              class="block text-[18px] font-semibold tracking-[-0.035em] text-[var(--text-strong)] sm:text-[28px]"
            >
              {{ plan.price === 0 ? 'Free' : `$${plan.price}`
              }}<span
                v-if="plan.price > 0"
                class="text-xs font-normal tracking-normal text-[var(--text-dim)]"
                >/mo</span
              >
            </span>
            <span class="block text-xs text-[var(--text-dim)] max-sm:hidden">{{
              plan.description
            }}</span>
          </th>
        </tr>
      </thead>
      <tbody>
        <template v-for="row in planRows" :key="row.label">
          <tr class="border-t border-[var(--border)] sm:hidden">
            <td colspan="5" class="px-2 pt-3 text-xs text-[var(--text-dim)]">
              {{ row.label }}
            </td>
          </tr>
          <tr class="sm:border-t sm:border-[var(--border)]">
            <th
              scope="row"
              class="px-2 py-3 font-normal text-[var(--text-dim)] max-sm:hidden sm:px-5"
            >
              {{ row.label }}
            </th>
            <td
              v-for="(value, i) in row.values"
              :key="i"
              class="px-2 py-3 font-mono text-[var(--text-strong)] sm:px-5"
              :class="
                plans[i].highlight ? 'bg-[color-mix(in_srgb,var(--accent)_6%,var(--bg))]' : ''
              "
            >
              <IncludedIcon v-if="value === '✓'" class="align-middle" />
              <template v-else>{{ value }}</template>
            </td>
          </tr>
        </template>
        <tr class="border-t border-[var(--border)] max-sm:hidden">
          <td class="px-5 py-4"></td>
          <td
            v-for="plan in plans"
            :key="plan.name"
            class="px-5 py-4"
            :class="plan.highlight ? 'bg-[color-mix(in_srgb,var(--accent)_6%,var(--bg))]' : ''"
          >
            <RouterLink
              to="/sign-up"
              class="block rounded-[10px] px-3 py-2 text-center text-sm font-semibold whitespace-nowrap transition-colors"
              :class="
                plan.highlight
                  ? 'bg-[var(--accent)] text-[var(--on-accent)] hover:bg-[var(--accent-hover)]'
                  : 'border border-[var(--border-strong)] text-[var(--text-strong)] hover:bg-[var(--surface-raised)]'
              "
            >
              {{ plan.cta }}
            </RouterLink>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
