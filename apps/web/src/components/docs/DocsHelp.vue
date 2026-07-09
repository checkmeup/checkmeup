<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<template>
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
</template>
