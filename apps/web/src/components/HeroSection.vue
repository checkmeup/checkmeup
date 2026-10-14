<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<template>
  <!-- The hero shows the product's one moment of value: the line you add to a
       cron job, and the alert you get when that job stops. -->
  <section class="mx-auto max-w-[1200px] px-4 pt-16 pb-20 sm:px-6 md:pt-24">
    <div class="grid items-center gap-14 lg:grid-cols-2">
      <div class="flex min-w-0 flex-col gap-6">
        <p class="font-mono text-[13px] text-[var(--text-dim)]">
          cron · uptime · ssl · domain · port · dns
        </p>
        <h1
          class="text-[44px] leading-[1.0] font-semibold tracking-[-0.045em] text-[var(--text-strong)] sm:text-[64px]"
        >
          Know before your client does.
        </h1>
        <p class="max-w-[480px] text-[17px] leading-relaxed text-[var(--text-dim)] sm:text-lg">
          Add one line to the end of a cron job, or paste a URL. When something stops, you hear
          about it on Telegram, Slack or email — before the person who pays you does.
        </p>
        <div class="flex flex-col gap-3 sm:flex-row">
          <RouterLink
            to="/sign-up"
            class="rounded-[10px] bg-[var(--accent)] px-6 py-3 text-center text-[15px] font-semibold text-[var(--on-accent)] transition-colors hover:bg-[var(--accent-hover)]"
          >
            Start free — no credit card
          </RouterLink>
          <RouterLink
            to="/docs"
            class="rounded-[10px] border border-[var(--border-strong)] px-6 py-3 text-center text-[15px] font-medium text-[var(--text)] transition-colors hover:bg-[var(--surface-raised)]"
          >
            Read the docs
          </RouterLink>
        </div>
      </div>

      <div
        aria-label="Example: a cron job and its missed-ping alert"
        role="img"
        class="flex min-w-0 flex-col gap-3.5"
      >
        <div
          class="overflow-x-auto rounded-[14px] bg-[var(--terminal-bg)] px-5 py-5 font-mono text-[13.5px] leading-[1.75] text-[var(--terminal-text)]"
        >
          <div class="mb-2.5 flex justify-between text-xs text-[var(--terminal-dim)]">
            <span>crontab -e</span>
            <span>server-01</span>
          </div>
          <div class="text-[var(--terminal-dim)]"># nightly database backup</div>
          <div class="whitespace-nowrap">
            0 3 * * * /opt/backup.sh
            <span class="text-[var(--terminal-accent)]"
              >&amp;&amp; curl -fsS checkmeup.net/ping/k7f2c9</span
            >
          </div>
        </div>

        <div class="flex items-center gap-2.5 pl-5 font-mono text-xs text-[var(--text-muted)]">
          <svg
            aria-hidden="true"
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <line x1="12" y1="4" x2="12" y2="20" />
            <polyline points="6 14 12 20 18 14" />
          </svg>
          03:00 passes, no ping arrives
        </div>

        <div
          class="flex items-start gap-3.5 rounded-[14px] border border-[var(--border)] bg-[var(--bg)] px-5 py-[18px] shadow-[0_18px_40px_rgb(0_0_0/7%)]"
        >
          <img src="/favicon.svg" alt="" class="h-9 w-9 flex-shrink-0" />
          <div class="flex min-w-0 flex-1 flex-col gap-1">
            <div class="flex justify-between gap-3 text-[13px] text-[var(--text-muted)]">
              <span class="font-semibold text-[var(--text-strong)]">Checkmeup</span>
              <span class="font-mono">03:05</span>
            </div>
            <div
              class="flex items-center gap-2 text-[15px] font-semibold text-[var(--text-strong)]"
            >
              <span class="h-2 w-2 flex-shrink-0 rounded-full bg-[var(--status-down)]"></span>
              nightly-backup missed its 03:00 ping
            </div>
            <div class="text-sm text-[var(--text-dim)]">
              Last ping yesterday at 03:00 · grace period 5 min · Telegram
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
