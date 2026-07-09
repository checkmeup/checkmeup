<template>
  <section id="email" class="scroll-mt-24">
    <h2 class="text-2xl font-bold mb-4" style="color: var(--text-strong)">Email alerts</h2>
    <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">
      Send alerts to any email address — use alongside Telegram and Slack, or on its own. Set
      it up in
      <strong style="color: var(--text-strong)">Settings → Notification channels</strong>:
    </p>
    <ol class="space-y-2 text-sm list-decimal list-inside" style="color: var(--text-dim)">
      <li>
        Click <strong style="color: var(--text-strong)">Add channel</strong> and choose Email.
      </li>
      <li>
        Enter an address — can be a shared inbox like
        <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
          >alerts@yourteam.com</code
        >
        rather than your login email.
      </li>
      <li>
        Click <strong style="color: var(--text-strong)">Send test email</strong> to confirm
        it's deliverable, then <strong style="color: var(--text-strong)">Save</strong>.
      </li>
    </ol>
    <p class="text-sm leading-relaxed mt-4" style="color: var(--text-dim)">
      Down and recovery alerts include monitor name, reason, and timestamp, with a subject
      line containing the monitor name and "DOWN" for inbox filtering. A single alert event
      fires on each assigned channel independently; the per-incident alert cap counts it as
      one notification regardless of how many channels fire.
    </p>
  </section>
</template>
