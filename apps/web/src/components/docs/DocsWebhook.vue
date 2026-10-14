<template>
  <section id="webhook" class="scroll-mt-24">
    <h2 class="mb-4 text-[26px] font-semibold tracking-[-0.03em] text-[var(--text-strong)]">Webhook alerts</h2>
    <p class="text-sm leading-relaxed mb-4 text-[var(--text-dim)]">
      A generic alert channel for wiring monitor events into your own automation — PagerDuty,
      a custom script, or anything else that isn't a first-class integration. Set it up in
      <strong class="text-[var(--text-strong)]">Settings → Notification channels</strong>:
    </p>
    <ol class="space-y-2 text-sm list-decimal list-inside text-[var(--text-dim)]">
      <li>
        Click <strong class="text-[var(--text-strong)]">Add channel</strong>, choose Webhook,
        and enter a URL that starts with
        <code class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
          >https://</code
        >.
      </li>
      <li>
        Click <strong class="text-[var(--text-strong)]">Send test webhook</strong> to verify
        delivery before saving.
      </li>
      <li>
        A signing secret is generated automatically once the channel is saved — use it to
        verify requests really came from Checkmeup (see below). Regenerate it any time from
        the channel's edit view.
      </li>
    </ol>
    <p class="text-sm leading-relaxed mt-4 mb-4 text-[var(--text-dim)]">
      On every down/recovery event (and for the test button), Checkmeup sends a single
      unretried
      <code class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
        >POST</code
      >
      with a JSON body:
    </p>
    <pre
      class="mb-4 overflow-x-auto rounded-[14px] bg-[var(--surface)] p-5 font-mono text-[13px] leading-relaxed text-[var(--text)]"
    ><code>{
  "eventType": "down",
  "monitorName": "api.example.com",
  "monitorType": "uptime",
  "reason": "HTTP 503",
  "timestamp": "2026-07-10T14:32:00Z"
}</code></pre>
    <p class="text-sm leading-relaxed text-[var(--text-dim)]">
      The request carries an
      <code class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
        >X-Checkmeup-Signature</code
      >
      header — a hex-encoded HMAC-SHA256 of the raw body, signed with your channel's secret —
      so you can verify it before acting on it. You can add multiple webhook channels and
      assign them to specific monitors rather than all of them.
    </p>
  </section>
</template>
