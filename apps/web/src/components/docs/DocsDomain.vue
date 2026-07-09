<template>
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
</template>
