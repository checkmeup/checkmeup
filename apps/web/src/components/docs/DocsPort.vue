<template>
  <section id="port" class="scroll-mt-24">
    <h2 class="text-2xl font-bold mb-4" style="color: var(--text-strong)">
      Port (TCP) monitoring
    </h2>
    <p class="text-sm leading-relaxed mb-4" style="color: var(--text-dim)">
      For anything that doesn't speak HTTP — a mail server, a database, a custom daemon — give
      us a host and a port and we open a raw TCP connection on your chosen interval, no data
      sent or received. Host and port are editable after creation, unlike SSL and domain
      monitors' hostname field.
    </p>
    <p class="text-sm leading-relaxed" style="color: var(--text-dim)">
      Each monitor has an expected state. The default,
      <strong style="color: var(--text-strong)">open</strong>, alerts if the port stops
      accepting connections — the familiar uptime check. The other option,
      <strong style="color: var(--text-strong)">closed</strong>, inverts that: alerts if the
      port unexpectedly starts accepting connections. That's a security check, not an uptime
      one — confirming a port that should be firewalled off (a database bound to a public
      interface, an admin panel, a debug port left on by mistake) actually stays unreachable,
      and catching the exact moment it doesn't.
    </p>
  </section>
</template>
