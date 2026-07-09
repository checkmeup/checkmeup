<template>
  <section id="slack" class="scroll-mt-24">
    <h2 class="text-2xl font-bold mb-4" style="color: var(--text-strong)">Slack alerts</h2>
    <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">
      Post alerts to any Slack channel via an Incoming Webhook. Set it up in
      <strong style="color: var(--text-strong)">Settings → Notification channels</strong>:
    </p>
    <ol class="space-y-2 text-sm list-decimal list-inside" style="color: var(--text-dim)">
      <li>
        In Slack, go to
        <strong style="color: var(--text-strong)">Apps → Incoming Webhooks</strong> and create
        a new webhook for your target channel.
      </li>
      <li>
        Copy the Webhook URL (starts with
        <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
          >https://hooks.slack.com/services/</code
        >).
      </li>
      <li>
        In Settings → Notification channels, click
        <strong style="color: var(--text-strong)">Add channel</strong>, choose Slack, paste
        the URL, and click
        <strong style="color: var(--text-strong)">Send test message</strong>
        to verify, then Save.
      </li>
    </ol>
    <p class="text-sm leading-relaxed mt-4" style="color: var(--text-dim)">
      You can add multiple Slack channels — name each after the workspace or channel it posts
      to — and assign them to specific monitors rather than all of them.
    </p>
  </section>
</template>
