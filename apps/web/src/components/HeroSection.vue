<template>
  <!-- Hero -->
  <section class="relative pt-16 pb-0 overflow-hidden">
    <!-- Grid texture -->
    <div
      class="absolute inset-0 pointer-events-none"
      style="
        background-image:
          linear-gradient(var(--border) 1px, transparent 1px),
          linear-gradient(90deg, var(--border) 1px, transparent 1px);
        background-size: 48px 48px;
      "
    ></div>
    <!-- Accent glow, top-right -->
    <div
      class="absolute pointer-events-none"
      style="
        top: -100px;
        right: -120px;
        width: 700px;
        height: 600px;
        background: radial-gradient(ellipse at 70% 30%, var(--accent-wash) 0%, transparent 65%);
      "
    ></div>
    <!-- Bottom fade -->
    <div
      class="absolute inset-x-0 bottom-0 pointer-events-none"
      style="height: 120px; background: linear-gradient(to bottom, transparent, var(--bg))"
    ></div>

    <div
      class="relative max-w-[1160px] mx-auto px-4 sm:px-7 pt-16 pb-20 grid lg:grid-cols-[1.05fr_1fr] gap-14 items-center"
    >
      <!-- Left: copy -->
      <div class="text-center lg:text-left">
        <div
          class="inline-flex items-center gap-2 text-xs font-normal px-3 py-1.5 rounded-full mb-6"
          style="
            border: 1px solid var(--border);
            background-color: var(--card);
            color: var(--text-dim);
          "
        >
          <span
            class="w-1.5 h-1.5 rounded-full flex-shrink-0 animate-pulse"
            style="background-color: var(--accent)"
          ></span>
          Built for freelancers &amp; solo devs with client sites
        </div>

        <h1
          class="text-4xl sm:text-5xl lg:text-6xl font-extrabold tracking-tight mb-5 leading-[1.08]"
          style="color: var(--text-strong)"
        >
          Know before<br />
          <span style="color: var(--accent-emphasis)">your client does.</span>
        </h1>

        <p
          class="text-[16.5px] font-normal mx-auto lg:mx-0 mb-8 max-w-[480px]"
          style="color: var(--text-muted)"
        >
          Cron job monitoring, uptime checks, SSL and domain expiry alerts, and raw TCP port
          checks — for every client site you maintain. Get alerted on Telegram, Slack, or email
          before your client ever notices something's wrong.
        </p>

        <div
          class="flex flex-col sm:flex-row items-center lg:items-start justify-center lg:justify-start gap-3"
        >
          <RouterLink
            to="/sign-up"
            class="w-full sm:w-auto text-[15px] font-semibold px-6 py-3 rounded-md transition-opacity hover:opacity-90"
            style="background-color: var(--accent); color: var(--on-accent)"
          >
            Start free — no credit card
          </RouterLink>
          <RouterLink
            to="/sign-in"
            class="w-full sm:w-auto text-sm px-6 py-3 rounded-md border transition-colors bg-[var(--bg)] hover:bg-[var(--surface-raised)]"
            style="color: var(--text-dim); border-color: var(--border)"
          >
            Sign in →
          </RouterLink>
        </div>
      </div>

      <!-- Right: product screenshot -->
      <div
        class="rounded-2xl border overflow-hidden text-left"
        style="border-color: var(--border); background-color: var(--surface)"
      >
        <!-- Fake browser chrome -->
        <div
          class="flex items-center gap-1.5 px-4 py-3 border-b"
          style="border-color: var(--border); background-color: var(--surface-raised)"
        >
          <span class="w-2.5 h-2.5 rounded-full" style="background-color: #ff5f57"></span>
          <span class="w-2.5 h-2.5 rounded-full" style="background-color: #febc2e"></span>
          <span class="w-2.5 h-2.5 rounded-full" style="background-color: #28c840"></span>
          <span
            class="flex-1 text-center font-mono text-xs"
            style="color: var(--text-muted); margin-right: 76px"
          >
            checkmeup.net — Monitor
          </span>
        </div>

        <div class="p-5 sm:p-6">
          <div class="flex items-center gap-3 mb-4">
            <span class="font-mono text-xs" style="color: var(--text-muted)">← Back</span>
            <span class="text-base font-semibold" style="color: var(--text-strong)"
              >checkmeup.net</span
            >
          </div>

          <div
            class="flex items-start justify-between pb-4 mb-4 border-b"
            style="border-color: var(--border)"
          >
            <div>
              <div class="flex items-center gap-2 mb-1">
                <span class="relative inline-block w-2.5 h-2.5">
                  <span
                    class="absolute inset-0 rounded-full animate-ping"
                    style="background-color: var(--status-up); opacity: 0.55"
                  ></span>
                  <span
                    class="absolute inset-0 rounded-full"
                    style="background-color: var(--status-up)"
                  ></span>
                </span>
                <span class="text-sm font-semibold" style="color: var(--status-up)">Up</span>
              </div>
              <div class="font-mono text-xs" style="color: var(--text-muted)">
                https://checkmeup.net
              </div>
              <div class="font-mono text-xs mt-0.5" style="color: var(--text-muted)">
                Every 10 min · Last checked 7m ago
              </div>
            </div>
            <div class="flex gap-2">
              <button
                class="px-2.5 py-1.5 rounded-md border text-xs"
                style="
                  border-color: var(--border);
                  background-color: var(--card);
                  color: var(--text-dim);
                "
              >
                Edit
              </button>
              <button
                class="px-2.5 py-1.5 rounded-md border text-xs"
                style="
                  border-color: var(--border);
                  background-color: var(--card);
                  color: var(--text-dim);
                "
              >
                Pause
              </button>
              <button
                class="px-2.5 py-1.5 rounded-md border text-xs"
                style="
                  border-color: var(--status-down-wash);
                  background-color: var(--status-down-wash);
                  color: var(--status-down);
                "
              >
                Delete
              </button>
            </div>
          </div>

          <div class="grid grid-cols-3 mb-4">
            <div class="text-center py-3 border-r" style="border-color: var(--border)">
              <div class="font-mono text-xl font-medium" style="color: var(--text-strong)">
                100.00%
              </div>
              <div class="text-xs mt-1" style="color: var(--text-muted)">Uptime 24h</div>
            </div>
            <div class="text-center py-3 border-r" style="border-color: var(--border)">
              <div class="font-mono text-xl font-medium" style="color: var(--text-strong)">
                100.00%
              </div>
              <div class="text-xs mt-1" style="color: var(--text-muted)">Uptime 7d</div>
            </div>
            <div class="text-center py-3">
              <div class="font-mono text-xl font-medium" style="color: var(--text-strong)">
                99.91%
              </div>
              <div class="text-xs mt-1" style="color: var(--text-muted)">Uptime 30d</div>
            </div>
          </div>

          <div class="pt-4 border-t" style="border-color: var(--border)">
            <div class="text-xs font-medium mb-2.5" style="color: var(--text-dim)">
              Response time — last 24h
            </div>
            <svg viewBox="0 0 640 48" fill="none" class="w-full h-12 block">
              <path
                d="M0 34 C10 32 18 27 32 30 C46 33 54 24 68 27 C82 30 92 37 106 34 C120 31 128 22 142 25 C156 28 166 34 180 31 C194 28 202 20 216 23 C230 26 240 33 254 29 C268 25 278 17 292 21 C306 25 316 33 330 29 C344 25 352 17 366 21 C380 25 392 33 406 29 C420 25 428 17 442 21 C456 25 468 33 482 29 C496 25 504 17 518 21 C532 25 544 33 558 29 C572 25 580 17 594 21 C608 25 620 33 634 29 C638 27 640 26 640 27"
                style="stroke: var(--status-up); stroke-width: 1.5px"
                fill="none"
              />
            </svg>
            <div class="font-mono text-xs mt-1.5" style="color: var(--text-muted)">
              Red dots = failed checks. Y axis = response time (max 77ms).
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
