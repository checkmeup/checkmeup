<template>
  <section id="webhook" class="scroll-mt-24">
    <h2 class="text-2xl font-bold mb-4" style="color: var(--text-strong)">Webhook alerts</h2>
    <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">
      A generic alert channel for wiring monitor events into your own automation — PagerDuty,
      a custom script, or anything else that isn't a first-class integration. Set it up in
      <strong style="color: var(--text-strong)">Settings → Notification channels</strong>:
    </p>
    <ol class="space-y-2 text-sm list-decimal list-inside" style="color: var(--text-dim)">
      <li>
        Click <strong style="color: var(--text-strong)">Add channel</strong>, choose Webhook,
        and enter a URL that starts with
        <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
          >https://</code
        >.
      </li>
      <li>
        Click <strong style="color: var(--text-strong)">Send test webhook</strong> to verify
        delivery before saving.
      </li>
      <li>
        A signing secret is generated automatically once the channel is saved — use it to
        verify requests really came from Checkmeup (see below). Regenerate it any time from
        the channel's edit view.
      </li>
    </ol>
    <p class="text-sm leading-relaxed mt-4 mb-4" style="color: var(--text-dim)">
      On every down/recovery event (and for the test button), Checkmeup sends a single
      unretried
      <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
        >POST</code
      >
      with a JSON body:
    </p>
    <pre
      class="rounded-xl border p-4 text-xs overflow-x-auto font-mono leading-relaxed mb-4"
      style="
        background-color: var(--surface);
        border-color: var(--border);
        color: var(--color-green-300);
      "
    ><code>{
  "eventType": "down",
  "monitorName": "api.example.com",
  "monitorType": "uptime",
  "reason": "HTTP 503",
  "timestamp": "2026-07-10T14:32:00Z"
}</code></pre>
    <p class="text-sm leading-relaxed" style="color: var(--text-dim)">
      The request carries an
      <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
        >X-Checkmeup-Signature</code
      >
      header — a hex-encoded HMAC-SHA256 of the raw body, signed with your channel's secret —
      so you can verify it before acting on it. You can add multiple webhook channels and
      assign them to specific monitors rather than all of them.
    </p>
  </section>
</template>
