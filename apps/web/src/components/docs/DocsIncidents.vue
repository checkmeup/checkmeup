<template>
  <section id="incidents" class="scroll-mt-24">
    <h2 class="text-2xl font-bold mb-4" style="color: var(--text-strong)">
      Incident management
    </h2>
    <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">
      Declare an incident from the
      <strong style="color: var(--text-strong)">Incidents</strong> page for anything affecting
      visitors that isn't a plain monitor-down — degraded performance, a partial outage, or
      something you want to narrate before a monitor even trips. Pick a title, an initial
      message, a severity (Minor / Major / Critical), and any combination of affected monitors.
    </p>
    <ul class="space-y-2 text-sm" style="color: var(--text-dim)">
      <li class="flex items-start gap-2">
        <span
          class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5"
          style="background-color: var(--color-green-500)"
        ></span>
        <span
          ><strong style="color: var(--text-strong)">Progress updates</strong> — post
          timestamped updates as you work the incident; status moves forward through
          Investigating → Identified → Monitoring → Resolved with each one.</span
        >
      </li>
      <li class="flex items-start gap-2">
        <span
          class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5"
          style="background-color: var(--color-green-500)"
        ></span>
        <span
          ><strong style="color: var(--text-strong)">Shown on the public status page</strong> —
          active incidents appear above the monitor list, and a Major or Critical incident
          escalates the overall status banner. Resolved incidents move to a paginated
          history.</span
        >
      </li>
      <li class="flex items-start gap-2">
        <span
          class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5"
          style="background-color: var(--color-green-500)"
        ></span>
        <span
          ><strong style="color: var(--text-strong)">Independent of monitor status</strong> —
          declaring, updating, or resolving an incident never changes a monitor's own up/down
          state, and doesn't trigger alerts on any channel.</span
        >
      </li>
      <li class="flex items-start gap-2">
        <span
          class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5"
          style="background-color: var(--color-green-500)"
        ></span>
        <span
          ><strong style="color: var(--text-strong)">Maintenance-aware</strong> — declaring an
          incident against a monitor already under an active maintenance window shows a warning
          first, so the two don't collide by accident.</span
        >
      </li>
    </ul>
  </section>
</template>
