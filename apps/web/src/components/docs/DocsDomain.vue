<template>
  <section id="domain" class="scroll-mt-24">
    <h2 class="mb-4 text-[26px] font-semibold tracking-[-0.03em] text-[var(--text-strong)]">
      Domain expiry monitoring
    </h2>
    <p class="text-sm leading-relaxed mb-4 text-[var(--text-dim)]">
      Separate from your SSL certificate — this tracks the domain
      <em>registration</em> itself, which is rarer to lapse but far more catastrophic when it does
      (your site disappears entirely, and the domain can get squatted before you notice). Give us an
      apex domain, same format as an SSL monitor's hostname, and we check its registration once a
      day and alert you at 30, 14, and 7 days before it expires, plus immediately if it's already
      expired.
    </p>
    <p class="text-sm leading-relaxed mb-4 text-[var(--text-dim)]">
      The dashboard shows the current registrar and exact expiry date for every monitored domain.
    </p>
    <p class="text-sm leading-relaxed text-[var(--text-dim)]">
      Registration data comes from the registry over RDAP, which covers .com, .net, .org and most
      other domains. A few registries don't run RDAP, so for .il, .io, .co, .me, .us, .it, .ru, .jp,
      .cn and .se domains we read their WHOIS instead. The .de, .eu, .ch, .li, .at and .es
      registries don't publish expiry dates at all, so domains under them can't be monitored; the
      form says so when you add one.
    </p>
  </section>
</template>
