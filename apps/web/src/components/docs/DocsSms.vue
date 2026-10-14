<template>
  <section id="sms" class="scroll-mt-24">
    <h2 class="mb-4 text-[26px] font-semibold tracking-[-0.03em] text-[var(--text-strong)]">SMS alerts</h2>
    <p class="text-sm leading-relaxed mb-4 text-[var(--text-dim)]">
      Text alerts to a phone number — available on Solo, Startup, and Enterprise (not Hobby),
      since carrier fees give SMS a real per-message cost the other channels don't have. Set
      it up in
      <strong class="text-[var(--text-strong)]">Settings → Notification channels</strong>:
    </p>
    <ol class="space-y-2 text-sm list-decimal list-inside text-[var(--text-dim)]">
      <li>
        Click <strong class="text-[var(--text-strong)]">Add channel</strong> and choose SMS.
      </li>
      <li>
        Enter the number in E.164 format (e.g.
        <code class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
          >+14155551234</code
        >) and check the consent box — required before any automated text can be sent to that
        number.
      </li>
      <li>
        Click <strong class="text-[var(--text-strong)]">Send test SMS</strong> to confirm
        delivery, then <strong class="text-[var(--text-strong)]">Save</strong>.
      </li>
    </ol>
    <p class="text-sm leading-relaxed mt-4 text-[var(--text-dim)]">
      Down and recovery texts are kept to a single message segment. Each plan includes a
      monthly SMS credit allowance (Solo 10 / Startup 30 / Enterprise 100 — one credit per
      message sent) that resets on the 1st of each calendar month and doesn't roll over; if a
      month's credits run out, that monitor's other alert channels (or your account email)
      still fire as normal.
    </p>
  </section>
</template>
