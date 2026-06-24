<script setup lang="ts">
import { RouterLink } from 'vue-router'
import LandingLayout from '@/layouts/LandingLayout.vue'

const sections = [
  { id: 'getting-started', label: 'Getting started' },
  { id: 'cron', label: 'Cron job monitoring' },
  { id: 'uptime', label: 'Uptime monitoring' },
  { id: 'ssl', label: 'SSL expiry monitoring' },
  { id: 'domain', label: 'Domain expiry monitoring' },
  { id: 'telegram', label: 'Telegram alerts' },
  { id: 'email', label: 'Email alerts' },
  { id: 'slack', label: 'Slack alerts' },
  { id: 'status-pages', label: 'Status pages' },
  { id: 'maintenance', label: 'Maintenance windows' },
  { id: 'appearance', label: 'Appearance' },
  { id: 'plans', label: 'Plans & limits' },
  { id: 'help', label: 'Need help?' },
]
</script>

<template>
  <LandingLayout>
    <!-- Hero -->
    <section class="max-w-6xl mx-auto px-4 sm:px-6 pt-16 pb-10 sm:pt-24 sm:pb-12">
      <h1
        class="text-4xl sm:text-5xl font-bold tracking-tight mb-4"
        style="color: var(--text-strong)"
      >
        Documentation
      </h1>
      <p class="text-lg max-w-2xl" style="color: var(--text-dim)">
        Everything checkmeup does today, in plain language — how to set up a monitor, what triggers
        an alert, and what each plan actually includes.
      </p>
    </section>

    <div class="max-w-6xl mx-auto px-4 sm:px-6 pb-24 flex flex-col lg:flex-row gap-12">
      <!-- Sidebar nav -->
      <nav class="lg:w-56 flex-shrink-0">
        <div
          class="lg:sticky lg:top-24 flex lg:flex-col gap-1 overflow-x-auto pb-2 lg:pb-0 -mx-1 px-1"
        >
          <a
            v-for="s in sections"
            :key="s.id"
            :href="`#${s.id}`"
            class="text-sm px-3 py-1.5 rounded-md whitespace-nowrap transition-colors hover:underline"
            style="color: var(--text-dim)"
          >
            {{ s.label }}
          </a>
        </div>
      </nav>

      <!-- Content -->
      <div class="flex-1 min-w-0 space-y-16">
        <!-- Getting started -->
        <section id="getting-started" class="scroll-mt-24">
          <h2 class="text-2xl font-bold mb-4" style="color: var(--text-strong)">Getting started</h2>
          <p class="text-sm leading-relaxed mb-3" style="color: var(--text-dim)">
            Sign up, pick a monitor type from the dashboard, and fill in the form. There's no setup
            wizard and no required onboarding call — the whole thing takes under a minute for your
            first monitor.
          </p>
          <p class="text-sm leading-relaxed" style="color: var(--text-dim)">
            checkmeup watches four kinds of things: scheduled jobs that should run on a cadence
            (cron monitors), URLs that should always respond (uptime monitors), TLS certificates
            that shouldn't be allowed to quietly expire (SSL monitors), and domain registrations
            that shouldn't be allowed to lapse (domain monitors). Connect Telegram, add an alert
            email, or both — every monitor type alerts through whichever channels you enable.
          </p>
        </section>

        <!-- Cron -->
        <section id="cron" class="scroll-mt-24">
          <h2 class="text-2xl font-bold mb-4" style="color: var(--text-strong)">
            Cron job monitoring
          </h2>
          <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">
            Create a cron monitor and you'll get a unique ping URL. Call it at the end of your job —
            if the ping doesn't arrive on schedule, you get alerted.
          </p>
          <pre
            class="rounded-xl border p-4 text-xs overflow-x-auto font-mono leading-relaxed mb-4"
            style="
              background-color: var(--surface);
              border-color: var(--border);
              color: var(--color-green-300);
            "
          ><code># Add this to the end of any cron job, script, or pipeline step
curl -s https://checkmeup.net/ping/&lt;your-monitor-token&gt;</code></pre>
          <ul class="space-y-2 text-sm" style="color: var(--text-dim)">
            <li class="flex items-start gap-2">
              <span
                class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5"
                style="background-color: var(--color-green-500)"
              ></span>
              <span
                ><strong style="color: var(--text-strong)">Schedule presets</strong> — every hour,
                every 30 minutes, daily at a fixed time, or weekdays only. Standard cron expressions
                also work.</span
              >
            </li>
            <li class="flex items-start gap-2">
              <span
                class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5"
                style="background-color: var(--color-green-500)"
              ></span>
              <span
                ><strong style="color: var(--text-strong)">Grace period</strong> — how late a ping
                can arrive before it counts as missed. Choose from 1 minute up to 1 hour, depending
                on how tight your job's timing is.</span
              >
            </li>
            <li class="flex items-start gap-2">
              <span
                class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5"
                style="background-color: var(--color-green-500)"
              ></span>
              <span
                ><strong style="color: var(--text-strong)">Execution history</strong> — every ping
                is logged, so you can see exactly when a job ran and how that compares to its
                schedule.</span
              >
            </li>
          </ul>
        </section>

        <!-- Uptime -->
        <section id="uptime" class="scroll-mt-24">
          <h2 class="text-2xl font-bold mb-4" style="color: var(--text-strong)">
            Uptime monitoring
          </h2>
          <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">
            Point an uptime monitor at any URL. We send a GET request on your chosen interval — 5,
            10, or 30 minutes on Hobby, plus 1 minute on paid plans — with a 10-second timeout, and
            expect an HTTP 200 back. Anything else (a different status code, a timeout, a connection
            error) opens an incident.
          </p>
          <ul class="space-y-2 text-sm" style="color: var(--text-dim)">
            <li class="flex items-start gap-2">
              <span
                class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5"
                style="background-color: var(--color-green-500)"
              ></span>
              <span
                ><strong style="color: var(--text-strong)">Response time tracking</strong> — every
                check records how long your server took to respond.</span
              >
            </li>
            <li class="flex items-start gap-2">
              <span
                class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5"
                style="background-color: var(--color-green-500)"
              ></span>
              <span
                ><strong style="color: var(--text-strong)">Incident history</strong> — start time,
                resolution time, and duration for every outage, so you can see patterns over
                time.</span
              >
            </li>
            <li class="flex items-start gap-2">
              <span
                class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5"
                style="background-color: var(--color-green-500)"
              ></span>
              <span
                ><strong style="color: var(--text-strong)">Uptime percentage</strong> — rolling
                24-hour, 7-day, and 30-day uptime shown on the monitor detail page.</span
              >
            </li>
            <li class="flex items-start gap-2">
              <span
                class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5"
                style="background-color: var(--color-green-500)"
              ></span>
              <span>
                <strong style="color: var(--text-strong)">Keyword monitoring</strong> (every plan,
                including Hobby) — add an
                optional keyword check alongside the status code, so a maintenance page served with
                a 200 or an error embedded in a JSON response still trips an alert. Set it to
                require the keyword (<code
                  class="px-1 rounded text-xs"
                  style="background-color: var(--surface-raised)"
                  >Contains</code
                >) or to fail when it's present (<code
                  class="px-1 rounded text-xs"
                  style="background-color: var(--surface-raised)"
                  >Does not contain</code
                >), with an optional case-sensitive toggle. We search the first 512 KB of the
                response body — no extra request, and the body itself is never stored, only the
                pass/fail reason.
              </span>
            </li>
          </ul>
        </section>

        <!-- SSL -->
        <section id="ssl" class="scroll-mt-24">
          <h2 class="text-2xl font-bold mb-4" style="color: var(--text-strong)">
            SSL expiry monitoring
          </h2>
          <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">
            Give us a hostname — no
            <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
              >https://</code
            >
            and no path, just the domain. We check the certificate on port 443 once a day and alert
            you at 30, 14, and 7 days before it expires, plus immediately if a check fails or the
            cert is already invalid.
          </p>
          <p class="text-sm leading-relaxed" style="color: var(--text-dim)">
            The dashboard shows the current issuer and exact expiry date for every monitored
            certificate, so renewal is never a surprise.
          </p>
        </section>

        <!-- Domain -->
        <section id="domain" class="scroll-mt-24">
          <h2 class="text-2xl font-bold mb-4" style="color: var(--text-strong)">
            Domain expiry monitoring
          </h2>
          <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">
            Separate from your SSL certificate — this tracks the domain
            <em>registration</em> itself, which is rarer to lapse but far more catastrophic when it
            does (your site disappears entirely, and the domain can get squatted before you notice).
            Give us an apex domain, same format as an SSL monitor's hostname, and we check its
            registration once a day and alert you at 30, 14, and 7 days before it expires, plus
            immediately if it's already expired.
          </p>
          <p class="text-sm leading-relaxed" style="color: var(--text-dim)">
            The dashboard shows the current registrar and exact expiry date for every monitored
            domain.
          </p>
        </section>

        <!-- Telegram -->
        <section id="telegram" class="scroll-mt-24">
          <h2 class="text-2xl font-bold mb-4" style="color: var(--text-strong)">Telegram alerts</h2>
          <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">
            Every monitor type can alert through Telegram. Set it up in
            <strong style="color: var(--text-strong)">Settings → Notification channels</strong>:
          </p>
          <ol class="space-y-2 text-sm list-decimal list-inside" style="color: var(--text-dim)">
            <li>
              Open
              <a
                href="https://t.me/checkmeupnet_bot"
                target="_blank"
                rel="noopener noreferrer"
                class="underline"
                style="color: var(--color-green-500)"
                >@checkmeupnet_bot</a
              >
              in Telegram and send
              <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
                >/start</code
              >.
            </li>
            <li>The bot replies with your Chat ID.</li>
            <li>
              In Settings → Notification channels, click
              <strong style="color: var(--text-strong)">Add channel</strong>, choose Telegram,
              paste your Chat ID, and click
              <strong style="color: var(--text-strong)">Send test message</strong> to confirm it
              works.
            </li>
          </ol>
          <p class="text-sm leading-relaxed mt-4" style="color: var(--text-dim)">
            Alerts are capped per incident — by default, 3 alerts and then silence until the monitor
            recovers, so a flapping check doesn't spam your phone. The recovery alert always sends,
            regardless of the cap. The alert limit is configurable per monitor.
          </p>
        </section>

        <!-- Email -->
        <section id="email" class="scroll-mt-24">
          <h2 class="text-2xl font-bold mb-4" style="color: var(--text-strong)">Email alerts</h2>
          <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">
            Send alerts to any email address — use alongside Telegram and Slack, or on its own. Set
            it up in
            <strong style="color: var(--text-strong)">Settings → Notification channels</strong>:
          </p>
          <ol class="space-y-2 text-sm list-decimal list-inside" style="color: var(--text-dim)">
            <li>
              Click <strong style="color: var(--text-strong)">Add channel</strong> and choose
              Email.
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
            Down and recovery alerts include monitor name, reason, and timestamp, with a subject line
            containing the monitor name and "DOWN" for inbox filtering. A single alert event fires
            on each assigned channel independently; the per-incident alert cap counts it as one
            notification regardless of how many channels fire.
          </p>
        </section>

        <!-- Slack -->
        <section id="slack" class="scroll-mt-24">
          <h2 class="text-2xl font-bold mb-4" style="color: var(--text-strong)">Slack alerts</h2>
          <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">
            Post alerts to any Slack channel via an Incoming Webhook. Set it up in
            <strong style="color: var(--text-strong)">Settings → Notification channels</strong>:
          </p>
          <ol class="space-y-2 text-sm list-decimal list-inside" style="color: var(--text-dim)">
            <li>
              In Slack, go to
              <strong style="color: var(--text-strong)">Apps → Incoming Webhooks</strong> and
              create a new webhook for your target channel.
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
              the URL, and click <strong style="color: var(--text-strong)">Send test message</strong>
              to verify, then Save.
            </li>
          </ol>
          <p class="text-sm leading-relaxed mt-4" style="color: var(--text-dim)">
            You can add multiple Slack channels — name each after the workspace or channel it posts
            to — and assign them to specific monitors rather than all of them.
          </p>
        </section>

        <!-- Status pages -->
        <section id="status-pages" class="scroll-mt-24">
          <h2 class="text-2xl font-bold mb-4" style="color: var(--text-strong)">Status pages</h2>
          <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">
            Create a public status page at
            <code class="px-1 rounded text-xs" style="background-color: var(--surface-raised)"
              >checkmeup.net/status/your-slug</code
            >
            and add any of your monitors to it. Visitors see live status and incident history — no
            login required, nothing to install on their end.
          </p>
          <p class="text-sm leading-relaxed" style="color: var(--text-dim)">
            There's no subdomain or DNS setup. Status pages live on the same domain by design, which
            means agencies can share a clean link with clients in seconds instead of provisioning
            DNS for every account.
          </p>
        </section>

        <!-- Maintenance windows -->
        <section id="maintenance" class="scroll-mt-24">
          <h2 class="text-2xl font-bold mb-4" style="color: var(--text-strong)">
            Maintenance windows
          </h2>
          <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">
            Schedule a maintenance window from the
            <strong style="color: var(--text-strong)">Maintenance</strong> page and pick any
            combination of cron, uptime, SSL, and domain monitors to cover. While a window is active, those
            monitors aren't checked at all — no alerts, no incidents, and your uptime stats stay
            untouched.
          </p>
          <ul class="space-y-2 text-sm" style="color: var(--text-dim)">
            <li class="flex items-start gap-2">
              <span
                class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5"
                style="background-color: var(--color-green-500)"
              ></span>
              <span
                ><strong style="color: var(--text-strong)">Planned or unplanned</strong> — set a
                start and end time ahead of a deploy, or start one immediately with no end date and
                close it manually with
                <strong style="color: var(--text-strong)">End now</strong> once the incident is
                resolved.</span
              >
            </li>
            <li class="flex items-start gap-2">
              <span
                class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5"
                style="background-color: var(--color-green-500)"
              ></span>
              <span
                ><strong style="color: var(--text-strong)">Multiple monitors per window</strong> —
                one window can cover everything affected by a single deploy or maintenance
                task.</span
              >
            </li>
            <li class="flex items-start gap-2">
              <span
                class="flex-shrink-0 w-1.5 h-1.5 rounded-full mt-1.5"
                style="background-color: var(--color-green-500)"
              ></span>
              <span
                ><strong style="color: var(--text-strong)">Visible on status pages</strong> — any
                covered monitor shows "Under maintenance" with an optional message instead of
                up/down, and doesn't trigger an outage banner.</span
              >
            </li>
          </ul>
        </section>

        <!-- Appearance -->
        <section id="appearance" class="scroll-mt-24">
          <h2 class="text-2xl font-bold mb-4" style="color: var(--text-strong)">Appearance</h2>
          <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">
            checkmeup supports light and dark themes. Switch anytime from the toggle in the app
            shell or
            <strong style="color: var(--text-strong)">Settings → Appearance</strong> — it applies
            instantly, no reload. First visit defaults to your OS-level light/dark preference; after
            that, your choice is remembered on this device.
          </p>
        </section>

        <!-- Plans -->
        <section id="plans" class="scroll-mt-24">
          <h2 class="text-2xl font-bold mb-4" style="color: var(--text-strong)">Plans & limits</h2>
          <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">
            Limits are per organization and apply across all monitor types combined. The free Hobby
            plan is meant to be genuinely usable, not a trial.
          </p>
          <div class="rounded-xl border overflow-hidden" style="border-color: var(--border)">
            <table class="w-full text-sm">
              <thead>
                <tr style="background-color: var(--surface)">
                  <th class="text-left px-4 py-3 font-medium" style="color: var(--text-muted)">
                    Plan
                  </th>
                  <th class="text-left px-4 py-3 font-medium" style="color: var(--text-muted)">
                    Monitors
                  </th>
                  <th class="text-left px-4 py-3 font-medium" style="color: var(--text-muted)">
                    Status pages
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr style="border-top: 1px solid var(--border)">
                  <td class="px-4 py-3" style="color: var(--text-strong)">Hobby — Free</td>
                  <td class="px-4 py-3" style="color: var(--text-dim)">10</td>
                  <td class="px-4 py-3" style="color: var(--text-dim)">1</td>
                </tr>
                <tr style="border-top: 1px solid var(--border)">
                  <td class="px-4 py-3" style="color: var(--text-strong)">Solo — $9/mo</td>
                  <td class="px-4 py-3" style="color: var(--text-dim)">30</td>
                  <td class="px-4 py-3" style="color: var(--text-dim)">3</td>
                </tr>
                <tr style="border-top: 1px solid var(--border)">
                  <td class="px-4 py-3" style="color: var(--text-strong)">Startup — $29/mo</td>
                  <td class="px-4 py-3" style="color: var(--text-dim)">100</td>
                  <td class="px-4 py-3" style="color: var(--text-dim)">10</td>
                </tr>
                <tr style="border-top: 1px solid var(--border)">
                  <td class="px-4 py-3" style="color: var(--text-strong)">Enterprise — $99/mo</td>
                  <td class="px-4 py-3" style="color: var(--text-dim)">1000</td>
                  <td class="px-4 py-3" style="color: var(--text-dim)">100</td>
                </tr>
              </tbody>
            </table>
          </div>
          <p class="text-sm leading-relaxed mt-4" style="color: var(--text-dim)">
            Keyword monitoring is available on every plan, including Hobby. See the full feature
            comparison for everything else that's the same across every plan.
          </p>
          <RouterLink
            to="/pricing"
            class="inline-flex items-center gap-1 mt-4 text-sm transition-colors"
            style="color: var(--color-green-500)"
          >
            Full pricing details →
          </RouterLink>
        </section>

        <!-- Help -->
        <section id="help" class="scroll-mt-24">
          <h2 class="text-2xl font-bold mb-4" style="color: var(--text-strong)">Need help?</h2>
          <p class="text-sm leading-relaxed mb-6" style="color: var(--text-dim)">
            There's no support ticket queue — questions, bug reports, and feature requests all reach
            an engineer directly.
          </p>
          <div class="flex flex-wrap gap-4">
            <RouterLink
              to="/settings"
              class="inline-flex items-center gap-2 text-sm px-4 py-2 rounded-md border transition-colors"
              style="
                color: var(--text-dim);
                border-color: var(--border);
                background-color: var(--surface);
              "
            >
              <svg
                width="14"
                height="14"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
              >
                <path
                  d="M12 2a7 7 0 0 0-4 12.74V17a2 2 0 0 0 2 2h4a2 2 0 0 0 2-2v-2.26A7 7 0 0 0 12 2z"
                />
                <line x1="9" y1="22" x2="15" y2="22" />
              </svg>
              Suggest a feature
            </RouterLink>
            <a
              href="mailto:andrew@checkmeup.net"
              class="inline-flex items-center gap-2 text-sm px-4 py-2 rounded-md border transition-colors"
              style="
                color: var(--text-dim);
                border-color: var(--border);
                background-color: var(--surface);
              "
            >
              <svg
                width="14"
                height="14"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
              >
                <path
                  d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"
                />
                <polyline points="22,6 12,13 2,6" />
              </svg>
              andrew@checkmeup.net
            </a>
            <a
              href="https://github.com/checkmeup/checkmeup/issues"
              target="_blank"
              rel="noopener noreferrer"
              class="inline-flex items-center gap-2 text-sm px-4 py-2 rounded-md border transition-colors"
              style="
                color: var(--text-dim);
                border-color: var(--border);
                background-color: var(--surface);
              "
            >
              <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
                <path
                  d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z"
                />
              </svg>
              Open an issue
            </a>
          </div>
        </section>
      </div>
    </div>
  </LandingLayout>
</template>
