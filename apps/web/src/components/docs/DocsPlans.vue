<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<template>
  <section id="plans" class="scroll-mt-24">
    <h2 class="text-2xl font-bold mb-4" style="color: var(--text-strong)">Plans & limits</h2>
    <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">
      Limits are per organization and apply across all monitor types combined. The free Hobby
      plan is meant to be genuinely usable, not a trial.
    </p>
    <div class="rounded-xl border overflow-hidden" style="border-color: var(--border)">
      <table class="w-full text-sm">
        <thead>
          <tr style="background-color: var(--surface)">
            <th class="text-left px-4 py-3 font-medium" style="color: var(--text-muted)">
              Plan
            </th>
            <th class="text-left px-4 py-3 font-medium" style="color: var(--text-muted)">
              Monitors
            </th>
            <th class="text-left px-4 py-3 font-medium" style="color: var(--text-muted)">
              Status pages
            </th>
            <th class="text-left px-4 py-3 font-medium" style="color: var(--text-muted)">
              SMS credits/mo
            </th>
          </tr>
        </thead>
        <tbody>
          <tr style="border-top: 1px solid var(--border)">
            <td class="px-4 py-3" style="color: var(--text-strong)">Hobby — Free</td>
            <td class="px-4 py-3" style="color: var(--text-dim)">10</td>
            <td class="px-4 py-3" style="color: var(--text-dim)">1</td>
            <td class="px-4 py-3" style="color: var(--text-dim)">—</td>
          </tr>
          <tr style="border-top: 1px solid var(--border)">
            <td class="px-4 py-3" style="color: var(--text-strong)">Solo — $9/mo</td>
            <td class="px-4 py-3" style="color: var(--text-dim)">30</td>
            <td class="px-4 py-3" style="color: var(--text-dim)">3</td>
            <td class="px-4 py-3" style="color: var(--text-dim)">10</td>
          </tr>
          <tr style="border-top: 1px solid var(--border)">
            <td class="px-4 py-3" style="color: var(--text-strong)">Startup — $29/mo</td>
            <td class="px-4 py-3" style="color: var(--text-dim)">100</td>
            <td class="px-4 py-3" style="color: var(--text-dim)">10</td>
            <td class="px-4 py-3" style="color: var(--text-dim)">30</td>
          </tr>
          <tr style="border-top: 1px solid var(--border)">
            <td class="px-4 py-3" style="color: var(--text-strong)">Enterprise — $99/mo</td>
            <td class="px-4 py-3" style="color: var(--text-dim)">1000</td>
            <td class="px-4 py-3" style="color: var(--text-dim)">100</td>
            <td class="px-4 py-3" style="color: var(--text-dim)">100</td>
          </tr>
        </tbody>
      </table>
    </div>
    <p class="text-sm leading-relaxed mt-4" style="color: var(--text-dim)">
      Keyword monitoring is available on every plan, including Hobby. See the full feature
      comparison for everything else that's the same across every plan.
    </p>
    <p class="text-sm leading-relaxed mt-4" style="color: var(--text-dim)">
      Paid plans can also be billed <strong style="color: var(--text-strong)">annually</strong>
      — two months free compared to paying monthly — toggle it in
      <strong style="color: var(--text-strong)">Settings → Billing</strong>.
    </p>
    <RouterLink
      to="/pricing"
      class="inline-flex items-center gap-1 mt-4 text-sm transition-colors"
      style="color: var(--color-green-500)"
    >
      Full pricing details →
    </RouterLink>
  </section>
</template>
