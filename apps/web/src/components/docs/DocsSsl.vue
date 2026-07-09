<template>
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
</template>
