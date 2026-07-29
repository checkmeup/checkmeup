<template>
  <section id="dns" class="scroll-mt-24">
    <h2 class="text-2xl font-bold mb-4" style="color: var(--text-strong)">
      DNS record monitoring
    </h2>
    <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">
      A hijacked or accidentally-changed DNS record doesn't fail the way a downed server does —
      traffic just starts going somewhere else. Give us a hostname and a record type (A, AAAA,
      CNAME, MX, TXT, or NS) and we resolve it on your chosen interval, comparing every check
      against a stored value.
    </p>
    <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">
      Set an <strong style="color: var(--text-strong)">expected value</strong> to pin exactly what
      the record should resolve to — you'll be alerted from the moment the monitor is created if
      reality doesn't match. Leave it blank and the first successful check captures whatever it
      currently resolves to as a
      <strong style="color: var(--text-strong)">baseline</strong> instead; every later check
      compares against that captured value the same way.
    </p>
    <p class="text-sm leading-relaxed" style="color: var(--text-dim)">
      A lookup failure (NXDOMAIN, SERVFAIL, timeout) alerts as a distinct "can't resolve" state,
      separate from a value mismatch's "changed from X to Y" — those are different problems with
      different fixes. Editing a monitor to accept a new value (or clearing it back to blank) is
      how you tell Checkmeup an intentional change is fine going forward.
    </p>
  </section>
</template>
