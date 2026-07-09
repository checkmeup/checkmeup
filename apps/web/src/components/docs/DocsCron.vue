<template>
  <section id="cron" class="scroll-mt-24">
    <h2 class="text-2xl font-bold mb-4" style="color: var(--text-strong)">
      Cron job monitoring
    </h2>
    <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">
      Create a cron monitor and you'll get a unique ping URL. Call it at the end of your job —
      if the ping doesn't arrive on schedule, you get alerted.
    </p>
    <pre
      class="rounded-xl border p-4 text-xs overflow-x-auto font-mono leading-relaxed mb-4"
      style="
        background-color: var(--surface);
        border-color: var(--border);
        color: var(--color-green-300);
      "
    ><code># Add this to the end of any cron job, script, or pipeline step
curl -s https://checkmeup.net/ping/&lt;your-monitor-token&gt;</code></pre>
    <ul class="space-y-2 text-sm" style="color: var(--text-dim)">
      <li class="flex items-start gap-2">
        <span
          class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5"
          style="background-color: var(--color-green-500)"
        ></span>
        <span
          ><strong style="color: var(--text-strong)">Schedule presets</strong> — every hour,
          every 30 minutes, daily at a fixed time, or weekdays only. Standard cron expressions
          also work.</span
        >
      </li>
      <li class="flex items-start gap-2">
        <span
          class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5"
          style="background-color: var(--color-green-500)"
        ></span>
        <span
          ><strong style="color: var(--text-strong)">Grace period</strong> — how late a ping
          can arrive before it counts as missed. Choose from 1 minute up to 1 hour, depending
          on how tight your job's timing is.</span
        >
      </li>
      <li class="flex items-start gap-2">
        <span
          class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5"
          style="background-color: var(--color-green-500)"
        ></span>
        <span
          ><strong style="color: var(--text-strong)">Execution history</strong> — every ping
          is logged, so you can see exactly when a job ran and how that compares to its
          schedule.</span
        >
      </li>
      <li class="flex items-start gap-2">
        <span
          class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5"
          style="background-color: var(--color-green-500)"
        ></span>
        <span
          ><strong style="color: var(--text-strong)">Attach custom data to a ping</strong> —
          add query params to the ping URL and they're stored with that ping, e.g.
          <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
            >curl -s
            "https://checkmeup.net/ping/&lt;token&gt;?build=142&amp;state=success"</code
          >. Capped at 20 key/value pairs, 64 characters per key, 256 per value — extra pairs
          are silently dropped rather than failing the ping. Only the most recent ping's data
          is kept (each ping overwrites it, it isn't a history). Read it back through the
          <a href="#api" class="underline" style="color: var(--color-green-500)">public API</a
          >.</span
        >
      </li>
    </ul>
  </section>
</template>
