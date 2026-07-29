<template>
  <section id="api" class="scroll-mt-24">
    <h2 class="text-2xl font-bold mb-4" style="color: var(--text-strong)">Public API</h2>
    <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">
      Read a monitor's current status from scripts, CI pipelines, or your own dashboard —
      useful for a build step that should surface its result somewhere else, or a status
      display that isn't Checkmeup itself (think a physical LED, an internal ops dashboard).
      Generate a key in
      <strong style="color: var(--text-strong)">Settings → API keys</strong> — the raw key is
      shown once, so copy it immediately — then send it as an
      <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
        >X-API-Key</code
      >
      header. Keys are read-only for now and rate-limited to 60 requests/minute.
    </p>
    <pre
      class="rounded-xl border p-4 text-xs overflow-x-auto font-mono leading-relaxed mb-4"
      style="
        background-color: var(--surface);
        border-color: var(--border);
        color: var(--color-green-300);
      "
    ><code>curl -H "X-API-Key: cmu_live_..." \
https://checkmeup.net/api/v1/public/monitors/cron/&lt;monitor-id&gt;/status</code></pre>
    <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">
      The monitor ID is the UUID in its detail page URL. Swap
      <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
        >cron</code
      >
      for
      <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
        >uptime</code
      >,
      <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
        >ssl</code
      >,
      <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
        >domain</code
      >, or
      <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
        >port</code
      >, or
      <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
        >dns</code
      >
      to match the monitor type. A cron monitor's response looks like:
    </p>
    <pre
      class="rounded-xl border p-4 text-xs overflow-x-auto font-mono leading-relaxed mb-4"
      style="
        background-color: var(--surface);
        border-color: var(--border);
        color: var(--color-green-300);
      "
    ><code>{
"id": "5e2b...",
"name": "Nightly export",
"type": "cron",
"status": "up",
"lastCheckedAt": "2026-07-03T20:33:47Z",
"lastPingMetadata": { "build": "142", "state": "success" }
}</code></pre>
    <p class="text-sm leading-relaxed" style="color: var(--text-dim)">
      <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
        >lastPingMetadata</code
      >
      only appears for cron monitors, and only once a ping has arrived — a CI job can attach
      its own key/value pairs (build number, exit state, anything short) as query params on
      the ping URL itself, e.g.
      <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
        >?build=142&amp;state=success</code
      >, and read them back through this endpoint. It always reflects the
      <strong style="color: var(--text-strong)">most recent</strong> ping — sending a new one
      overwrites it, it isn't a history. Capped at 20 key/value pairs per ping, 64 characters
      per key, and 256 characters per value; anything past that is silently dropped rather
      than rejected, since a ping must always succeed. SSL and domain monitors instead include
      <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
        >expiresAt</code
      >
      and
      <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
        >daysUntilExpiry</code
      >.
    </p>
    <p class="text-sm leading-relaxed mt-4" style="color: var(--text-dim)">
      Up to 100 active keys per org, on every plan — revoke an old one to free up room for a
      new one.
    </p>
  </section>
</template>
