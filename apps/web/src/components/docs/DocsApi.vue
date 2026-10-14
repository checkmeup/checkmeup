<template>
  <section id="api" class="scroll-mt-24">
    <h2 class="mb-4 text-[26px] font-semibold tracking-[-0.03em] text-[var(--text-strong)]">Public API</h2>
    <p class="text-sm leading-relaxed mb-4 text-[var(--text-dim)]">
      Read a monitor's current status from scripts, CI pipelines, or your own dashboard —
      useful for a build step that should surface its result somewhere else, or a status
      display that isn't Checkmeup itself (think a physical LED, an internal ops dashboard).
      Generate a key in
      <strong class="text-[var(--text-strong)]">Settings → API keys</strong> — the raw key is
      shown once, so copy it immediately — then send it as an
      <code class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
        >X-API-Key</code
      >
      header. Keys are read-only for now and rate-limited to 60 requests/minute.
    </p>
    <pre
      class="mb-4 overflow-x-auto rounded-[14px] bg-[var(--surface)] p-5 font-mono text-[13px] leading-relaxed text-[var(--text)]"
    ><code>curl -H "X-API-Key: cmu_live_..." \
https://checkmeup.net/api/v1/public/monitors/cron/&lt;monitor-id&gt;/status</code></pre>
    <p class="text-sm leading-relaxed mb-4 text-[var(--text-dim)]">
      The monitor ID is the UUID in its detail page URL. Swap
      <code class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
        >cron</code
      >
      for
      <code class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
        >uptime</code
      >,
      <code class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
        >ssl</code
      >,
      <code class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
        >domain</code
      >, or
      <code class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
        >port</code
      >, or
      <code class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
        >dns</code
      >
      to match the monitor type. A cron monitor's response looks like:
    </p>
    <pre
      class="mb-4 overflow-x-auto rounded-[14px] bg-[var(--surface)] p-5 font-mono text-[13px] leading-relaxed text-[var(--text)]"
    ><code>{
"id": "5e2b...",
"name": "Nightly export",
"type": "cron",
"status": "up",
"lastCheckedAt": "2026-07-03T20:33:47Z",
"lastPingMetadata": { "build": "142", "state": "success" }
}</code></pre>
    <p class="text-sm leading-relaxed text-[var(--text-dim)]">
      <code class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
        >lastPingMetadata</code
      >
      only appears for cron monitors, and only once a ping has arrived — a CI job can attach
      its own key/value pairs (build number, exit state, anything short) as query params on
      the ping URL itself, e.g.
      <code class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
        >?build=142&amp;state=success</code
      >, and read them back through this endpoint. It always reflects the
      <strong class="text-[var(--text-strong)]">most recent</strong> ping — sending a new one
      overwrites it, it isn't a history. Capped at 20 key/value pairs per ping, 64 characters
      per key, and 256 characters per value; anything past that is silently dropped rather
      than rejected, since a ping must always succeed. SSL and domain monitors instead include
      <code class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
        >expiresAt</code
      >
      and
      <code class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
        >daysUntilExpiry</code
      >.
    </p>
    <p class="text-sm leading-relaxed mt-4 text-[var(--text-dim)]">
      Up to 100 active keys per org, on every plan — revoke an old one to free up room for a
      new one.
    </p>
  </section>
</template>
