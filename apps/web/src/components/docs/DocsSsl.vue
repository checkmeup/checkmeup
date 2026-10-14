<template>
  <section id="ssl" class="scroll-mt-24">
    <h2 class="mb-4 text-[26px] font-semibold tracking-[-0.03em] text-[var(--text-strong)]">
      SSL expiry monitoring
    </h2>
    <p class="text-sm leading-relaxed mb-4 text-[var(--text-dim)]">
      Give us a hostname — no
      <code class="rounded-md bg-[var(--surface-raised)] px-1.5 py-0.5 font-mono text-[0.85em]"
        >https://</code
      >
      and no path, just the domain. We check the certificate on port 443 once a day and alert
      you at 30, 14, and 7 days before it expires, plus immediately if a check fails or the
      cert is already invalid.
    </p>
    <p class="text-sm leading-relaxed text-[var(--text-dim)]">
      The dashboard shows the current issuer and exact expiry date for every monitored
      certificate, so renewal is never a surprise.
    </p>
  </section>
</template>
