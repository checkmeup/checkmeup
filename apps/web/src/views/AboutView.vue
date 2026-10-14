<script setup lang="ts">
import LandingLayout from '@/layouts/LandingLayout.vue'
import CtaBanner from '@/components/CtaBanner.vue'
import LabeledSection from '@/components/LabeledSection.vue'
import PageHero from '@/components/PageHero.vue'
import { useSeo } from '@/composables/useSeo'

useSeo({
  title: 'About — Checkmeup',
  description: 'Who builds Checkmeup, the stack it runs on, and the values behind it.',
  path: '/about',
})

const stack = [
  { label: 'Backend', value: 'Go · Chi · sqlc · goose' },
  { label: 'Frontend', value: 'Vue 3 · Vite · Pinia · Tailwind' },
  { label: 'Database', value: 'PostgreSQL' },
  { label: 'Infrastructure', value: 'Hetzner · Kamal · Traefik' },
  { label: 'Alerts', value: 'Telegram Bot API' },
  { label: 'Billing', value: 'Paddle (Merchant of Record)' },
]

const values = [
  {
    title: 'Built by developers, for developers',
    body: 'No product managers, no UX researchers, no marketing-driven roadmaps. Every feature starts with a real problem a developer has actually hit. If the API is clunky, the person who built it hears about it — and fixes it.',
  },
  {
    title: 'Transparency by default',
    body: 'We publish what we know: the tech stack, the architecture decisions, and eventually the business metrics. If something breaks, we say so and say why. No vague status page updates. No PR-managed incidents.',
  },
  {
    title: 'Shared success',
    body: 'When you run a successful product, your monitors run quietly. When you grow, we grow with you. Pricing is designed so the free plan is genuinely useful, not a bait-and-switch — and paid plans scale with real value, not seat counts.',
  },
]
</script>

<template>
  <LandingLayout>
    <PageHero
      title="Built by a developer who got paged at 2 am."
      eyebrow="Independent · Bootstrapped · Built in the open"
    >
      <div class="space-y-4">
        <p>
          Every developer has been there. A cron job silently fails for three days. A certificate
          expires over a holiday weekend. A site goes down and you find out from a client, not from
          your own tooling.
        </p>
        <p>
          Checkmeup started as the tool I wished existed — simple, reliable, and built the way
          developers actually want to use software.
        </p>
      </div>
    </PageHero>

    <div
      class="mx-auto max-w-[1200px] px-4 pb-16 text-[15px] leading-relaxed text-[var(--text-dim)] sm:px-6"
    >
      <LabeledSection title="Who's building this">
        <div class="max-w-[680px] space-y-3">
          <div class="flex flex-wrap items-center gap-3">
            <span class="text-lg font-semibold tracking-[-0.02em] text-[var(--text-strong)]"
              >Andrew Molyuk</span
            >
            <div class="flex items-center gap-2">
              <a
                href="https://github.com/andrewmolyuk"
                target="_blank"
                rel="noopener noreferrer"
                class="text-[var(--text-dim)] transition-colors hover:text-[var(--text)]"
                aria-label="GitHub"
              >
                <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
                  <path
                    d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z"
                  />
                </svg>
              </a>
              <a
                href="https://linkedin.com/in/andrewmolyuk"
                target="_blank"
                rel="noopener noreferrer"
                class="text-[var(--text-dim)] transition-colors hover:text-[var(--text)]"
                aria-label="LinkedIn"
              >
                <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
                  <path
                    d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433a2.062 2.062 0 0 1-2.063-2.065 2.064 2.064 0 1 1 2.063 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z"
                  />
                </svg>
              </a>
              <a
                href="https://andrew.molyuk.com"
                target="_blank"
                rel="noopener noreferrer"
                class="text-[var(--text-dim)] transition-colors hover:text-[var(--text)]"
                aria-label="Personal site"
              >
                <svg
                  width="16"
                  height="16"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <circle cx="12" cy="12" r="10" />
                  <line x1="2" y1="12" x2="22" y2="12" />
                  <path
                    d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"
                  />
                </svg>
              </a>
            </div>
          </div>
          <p>
            25+ years writing software — from enterprise video platforms processing 23 billion
            events a day, to voting systems handling 30 million concurrent connections, to building
            and leading remote engineering teams across two continents.
          </p>
          <p>
            After spending years in the high-scale world — AWS, Kubernetes, microservices, the whole
            catalogue — I kept coming back to the same observation: the boring, reliable tools are
            the ones that actually stay running. Checkmeup is built on that principle. Go backend,
            PostgreSQL, a handful of Hetzner servers. No distributed systems unless the problem
            actually demands it.
          </p>
          <p>
            I'm based in Israel. Checkmeup is independent and bootstrapped — no VC, no growth
            hacking, no hockey-stick pressure. Just a developer building a tool that should exist.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="How we work">
        <div class="max-w-[680px] space-y-6">
          <div v-for="v in values" :key="v.title">
            <h3 class="mb-1.5 font-semibold text-[var(--text-strong)]">{{ v.title }}</h3>
            <p>{{ v.body }}</p>
          </div>
        </div>
      </LabeledSection>

      <LabeledSection title="The stack">
        <template #description>
          Boring by design. Every choice here was made to maximize reliability and minimize
          operational complexity — not to be impressive on a conference slide.
        </template>
        <div class="max-w-[680px]">
          <dl>
            <div
              v-for="item in stack"
              :key="item.label"
              class="grid grid-cols-[130px_minmax(0,1fr)] gap-4 border-b border-[var(--border)] py-2.5 first:pt-0"
            >
              <dt class="text-sm text-[var(--text-dim)]">{{ item.label }}</dt>
              <dd class="font-mono text-[13.5px] text-[var(--text)]">{{ item.value }}</dd>
            </div>
          </dl>
          <p class="mt-6 text-sm">
            Architecture decisions are documented in the
            <a
              href="https://github.com/checkmeup/checkmeup/tree/main/.about/adr/"
              target="_blank"
              rel="noopener noreferrer"
              class="text-[var(--accent)] hover:underline"
              >ADR log</a
            >
            in the public repo — including the choices that weren't obvious.
          </p>
          <p class="mt-3 text-sm">
            The code itself is source-available under the
            <a
              href="https://github.com/checkmeup/checkmeup/blob/main/LICENSE.md"
              target="_blank"
              rel="noopener noreferrer"
              class="text-[var(--accent)] hover:underline"
              >Business Source License</a
            >
            — free to read, self-host, and modify; the one restriction is reselling Checkmeup itself
            as a competing hosted monitoring service. It converts to Apache 2.0 in 2030.
          </p>
        </div>
      </LabeledSection>

      <LabeledSection title="Things we won't do">
        <div class="max-w-[680px]">
          <ul class="space-y-3">
            <li class="flex items-start gap-3">
              <span class="mt-2.5 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-[var(--accent)]"></span>
              <span
                >Sell or share your data. Your monitors, your alert history, your email — none of
                it.</span
              >
            </li>
            <li class="flex items-start gap-3">
              <span class="mt-2.5 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-[var(--accent)]"></span>
              <span
                >Raise VC money and change the pricing model six months later. Bootstrapped means
                the incentives stay aligned with you.</span
              >
            </li>
            <li class="flex items-start gap-3">
              <span class="mt-2.5 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-[var(--accent)]"></span>
              <span
                >Add features just because competitors have them. Every addition has a maintenance
                cost. We'd rather do three things well than fifteen things adequately.</span
              >
            </li>
            <li class="flex items-start gap-3">
              <span class="mt-2.5 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-[var(--accent)]"></span>
              <span
                >Write vague status page updates. If something breaks, we explain what broke, why,
                and what changed.</span
              >
            </li>
          </ul>
        </div>
      </LabeledSection>

      <LabeledSection title="Get in touch">
        <template #description>
          Questions, bug reports, feature ideas, complaints about the docs — all welcome. There's no
          support ticket queue. You'll reach an engineer.
        </template>
        <div class="flex flex-wrap gap-4">
          <a
            href="mailto:andrew@checkmeup.net"
            class="inline-flex items-center gap-2 text-sm px-4 py-2 rounded-[10px] border border-[var(--border-strong)] text-[var(--text)] transition-colors hover:bg-[var(--surface-raised)]"
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
            class="inline-flex items-center gap-2 text-sm px-4 py-2 rounded-[10px] border border-[var(--border-strong)] text-[var(--text)] transition-colors hover:bg-[var(--surface-raised)]"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
              <path
                d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0 0 24 12c0-6.63-5.37-12-12-12z"
              />
            </svg>
            Open an issue
          </a>
        </div>
      </LabeledSection>
    </div>

    <CtaBanner
      title="Try it. It's free."
      subtitle="10 monitors, no credit card, no expiry date on the free plan."
    />
  </LandingLayout>
</template>
