<template>
  <section id="maintenance" class="scroll-mt-24">
    <h2 class="mb-4 text-[26px] font-semibold tracking-[-0.03em] text-[var(--text-strong)]">
      Maintenance windows
    </h2>
    <p class="text-sm leading-relaxed mb-4 text-[var(--text-dim)]">
      Schedule a maintenance window from the
      <strong class="text-[var(--text-strong)]">Maintenance</strong> page and pick any
      combination of cron, uptime, SSL, domain, port, and DNS monitors to cover. While a window is
      active, those monitors aren't checked at all — no alerts, no incidents, and your uptime
      stats stay untouched.
    </p>
    <ul class="space-y-2 text-sm text-[var(--text-dim)]">
      <li class="flex items-start gap-2">
        <span
          class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5 bg-[var(--accent)]"
        ></span>
        <span
          ><strong class="text-[var(--text-strong)]">Planned or unplanned</strong> — set a
          start and end time ahead of a deploy, or start one immediately with no end date and
          close it manually with
          <strong class="text-[var(--text-strong)]">End now</strong> once the incident is
          resolved.</span
        >
      </li>
      <li class="flex items-start gap-2">
        <span
          class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5 bg-[var(--accent)]"
        ></span>
        <span
          ><strong class="text-[var(--text-strong)]">Multiple monitors per window</strong> —
          one window can cover everything affected by a single deploy or maintenance
          task.</span
        >
      </li>
      <li class="flex items-start gap-2">
        <span
          class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5 bg-[var(--accent)]"
        ></span>
        <span
          ><strong class="text-[var(--text-strong)]">Visible on status pages</strong> — any
          covered monitor shows "Under maintenance" with an optional message instead of
          up/down, and doesn't trigger an outage banner.</span
        >
      </li>
      <li class="flex items-start gap-2">
        <span
          class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5 bg-[var(--accent)]"
        ></span>
        <span
          ><strong class="text-[var(--text-strong)]">Up to 100 per org</strong> — on every
          plan. Delete an old one to free up room for a new one.</span
        >
      </li>
    </ul>
  </section>
</template>
