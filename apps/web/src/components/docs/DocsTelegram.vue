<template>
  <section id="telegram" class="scroll-mt-24">
    <h2 class="mb-4 text-[26px] font-semibold tracking-[-0.03em] text-[var(--text-strong)]">Telegram alerts</h2>
    <p class="text-sm leading-relaxed mb-4 text-[var(--text-dim)]">
      Every monitor type can alert through Telegram. Set it up in
      <strong class="text-[var(--text-strong)]">Settings → Notification channels</strong>:
    </p>
    <ol class="space-y-2 text-sm list-decimal list-inside text-[var(--text-dim)]">
      <li>
        Open
        <a
          href="https://t.me/checkmeupnet_bot"
          target="_blank"
          rel="noopener noreferrer"
          class="underline text-[var(--accent)]"
          >@checkmeupnet_bot</a
        >
        in Telegram and send
        <code class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
          >/start</code
        >.
      </li>
      <li>The bot replies with your Chat ID.</li>
      <li>
        In Settings → Notification channels, click
        <strong class="text-[var(--text-strong)]">Add channel</strong>, choose Telegram, paste
        your Chat ID, and click
        <strong class="text-[var(--text-strong)]">Send test message</strong> to confirm it
        works.
      </li>
    </ol>
    <p class="text-sm leading-relaxed mt-4 text-[var(--text-dim)]">
      Alerts are capped per incident — by default, 3 alerts and then silence until the monitor
      recovers, so a flapping check doesn't spam your phone. The recovery alert always sends,
      regardless of the cap. The alert limit is configurable per monitor.
    </p>
  </section>
</template>
