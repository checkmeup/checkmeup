import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "A database that stopped accepting connections at 2 AM doesn't show up in an uptime check — nothing was polling a URL on that host to begin with. Neither does the admin panel someone opened to the internet during a rushed deploy and forgot to close. Both are the same kind of blind spot: a service that lives on a port, not behind a URL, and nobody's watching whether that port is doing what it's supposed to.",
  },
  {
    type: 'h2',
    text: 'What port (TCP) monitoring actually does',
  },
  {
    type: 'p',
    text: "A port monitor opens a raw TCP connection to a host and port on a fixed interval and records whether the connection succeeded — no HTTP request, no protocol handshake, nothing sent or read once it's open. It's the simplest possible check there is: can something on the other end accept a connection right now, yes or no. That simplicity is the point — it works for anything that listens on a socket, not just things that speak HTTP.",
  },
  {
    type: 'h2',
    text: "Reachability, not correctness — and why that's still useful",
  },
  {
    type: 'p',
    text: "A TCP connect succeeding only proves one thing: something is listening. It doesn't confirm a mail server actually speaks SMTP, that a database will authenticate a real query, or that a Redis instance responds correctly to a command — a socket can accept a connection while the service behind it is completely broken. That's a real limit, not a hidden one, and it's why port monitoring is a complement to application-level checks, not a replacement for them. What it does catch, reliably, is the failure mode those checks can't: the process crashed, the firewall rule changed, the host is unreachable — anything where the port itself stopped answering, for a service that has no HTTP endpoint to poll in the first place.",
  },
  {
    type: 'h2',
    text: 'The other direction: watching for a port that should stay closed',
  },
  {
    type: 'p',
    text: 'Most monitoring assumes you want to know when something goes down. Port monitoring can run the opposite check just as easily: point it at a port that\'s supposed to be firewalled off — a database, an admin interface, a debug port that should never face the internet — and set the expected state to "closed" instead of "open." Now the alert fires the moment that port unexpectedly starts accepting connections, which is usually the first sign of a misconfigured security group or a firewall rule that got reverted. It\'s a monitor for a security posture, not an outage, using the same connect check either way.',
  },
  {
    type: 'h2',
    text: 'Setting up port monitoring in three steps',
  },
  {
    type: 'h3',
    text: '1. Decide which state you actually want to watch for',
  },
  {
    type: 'p',
    text: 'A service that should be reachable — a database your app connects to, a mail server, a game server, anything without an HTTP front end — gets an "open" check: alert when it stops accepting connections. Anything that should never be reachable from wherever the monitor runs gets a "closed" check: alert when it starts. Picking the wrong one means either missing the outage or missing the exposure.',
  },
  {
    type: 'h3',
    text: "2. Point it at the host and port that's actually load-bearing",
  },
  {
    type: 'p',
    text: "Monitor the specific host and port the failure would actually happen on — the database's real endpoint, not a load balancer in front of three replicas that could mask one of them dying. A monitor watching the wrong layer will stay green through a real problem.",
  },
  {
    type: 'h3',
    text: '3. Set a failure threshold before you trust the alert',
  },
  {
    type: 'p',
    text: "A single dropped connection on an otherwise healthy host shouldn't wake anyone up. Set the monitor to alert after more than one consecutive failure, and cap how many alerts a single incident can send, so a port that's genuinely down for an hour pages you once with useful information instead of every few minutes for the whole outage.",
  },
  {
    type: 'h2',
    text: 'What to look for in a port monitoring tool',
  },
  {
    type: 'ul',
    items: [
      'Both directions — a check for "should be reachable" and one for "should not be," not just the first',
      'A configurable check interval, short enough for whatever is actually behind that port',
      'A consecutive-failure threshold before alerting, so one dropped connection does not become constant noise',
      'A cap on alerts per incident, so a long outage does not turn into an alert every few minutes for its whole duration',
      'An incident log with start time and duration — not just a live badge that forgets the moment it goes green again',
    ],
  },
  {
    type: 'h2',
    text: 'The mistake worth avoiding',
  },
  {
    type: 'p',
    text: "Treating a green port monitor as proof the service behind it works is the easiest way to get burned. A TCP connect succeeding means the socket answered — it says nothing about whether the database will actually authenticate a query or the mail server will actually accept a message. For anything where the application layer matters as much as reachability, pair the port check with a real uptime or content check on top of it, and let the port monitor do the one thing it's actually built for: knowing the instant a socket stops (or starts) answering.",
  },
  {
    type: 'divider',
  },
  {
    type: 'p',
    text: "Checkmeup's port monitor covers both directions — open or closed expected state, configurable check intervals, a consecutive-failure threshold, a per-incident alert cap, and full incident history. The Hobby plan is free for up to 10 monitors of any type, no credit card required.",
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
