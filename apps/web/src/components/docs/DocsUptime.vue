<template>
  <section id="uptime" class="scroll-mt-24">
    <h2 class="text-2xl font-bold mb-4" style="color: var(--text-strong)">
      Uptime monitoring
    </h2>
    <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">
      Point an uptime monitor at any URL. We send a GET request on your chosen interval — 5,
      10, or 30 minutes on Hobby, plus 1 minute on paid plans — with a 10-second timeout, and
      expect an HTTP 200 back. Anything else (a different status code, a timeout, a connection
      error) opens an incident.
    </p>
    <ul class="space-y-2 text-sm" style="color: var(--text-dim)">
      <li class="flex items-start gap-2">
        <span
          class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5"
          style="background-color: var(--color-green-500)"
        ></span>
        <span
          ><strong style="color: var(--text-strong)">Response time tracking</strong> — every
          check records how long your server took to respond.</span
        >
      </li>
      <li class="flex items-start gap-2">
        <span
          class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5"
          style="background-color: var(--color-green-500)"
        ></span>
        <span
          ><strong style="color: var(--text-strong)">Incident history</strong> — start time,
          resolution time, and duration for every outage, so you can see patterns over
          time.</span
        >
      </li>
      <li class="flex items-start gap-2">
        <span
          class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5"
          style="background-color: var(--color-green-500)"
        ></span>
        <span
          ><strong style="color: var(--text-strong)">Uptime percentage</strong> — rolling
          24-hour, 7-day, and 30-day uptime shown on the monitor detail page.</span
        >
      </li>
      <li class="flex items-start gap-2">
        <span
          class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5"
          style="background-color: var(--color-green-500)"
        ></span>
        <span>
          <strong style="color: var(--text-strong)">Keyword monitoring</strong> (every plan,
          including Hobby) — add an optional keyword check alongside the status code, so a
          maintenance page served with a 200 or an error embedded in a JSON response still
          trips an alert. Set it to require the keyword (<code
            class="px-1 rounded text-xs"
            style="background-color: var(--surface-raised)"
            >Contains</code
          >) or to fail when it's present (<code
            class="px-1 rounded text-xs"
            style="background-color: var(--surface-raised)"
            >Does not contain</code
          >), with an optional case-sensitive toggle. We search the first 512 KB of the
          response body — no extra request, and the body itself is never stored, only the
          pass/fail reason.
        </span>
      </li>
    </ul>
  </section>
</template>
