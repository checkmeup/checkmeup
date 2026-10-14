<template>
  <section id="slack" class="scroll-mt-24">
    <h2 class="mb-4 text-[26px] font-semibold tracking-[-0.03em] text-[var(--text-strong)]">Slack alerts</h2>
    <p class="text-sm leading-relaxed mb-4 text-[var(--text-dim)]">
      Post alerts to any Slack channel via an Incoming Webhook. Set it up in
      <strong class="text-[var(--text-strong)]">Settings → Notification channels</strong>:
    </p>
    <ol class="space-y-2 text-sm list-decimal list-inside text-[var(--text-dim)]">
      <li>
        In Slack, go to
        <strong class="text-[var(--text-strong)]">Apps → Incoming Webhooks</strong> and create
        a new webhook for your target channel.
      </li>
      <li>
        Copy the Webhook URL (starts with
        <code class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
          >https://hooks.slack.com/services/</code
        >).
      </li>
      <li>
        In Settings → Notification channels, click
        <strong class="text-[var(--text-strong)]">Add channel</strong>, choose Slack, paste
        the URL, and click
        <strong class="text-[var(--text-strong)]">Send test message</strong>
        to verify, then Save.
      </li>
    </ol>
    <p class="text-sm leading-relaxed mt-4 text-[var(--text-dim)]">
      You can add multiple Slack channels — name each after the workspace or channel it posts
      to — and assign them to specific monitors rather than all of them.
    </p>
  </section>
</template>
