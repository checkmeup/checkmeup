import type { ContentBlock } from '../types'

export const content: ContentBlock[] = [
  {
    type: 'p',
    text: "Uptime monitoring only covers services that speak HTTP. A mail server on port 25, a database exposed on 5432, or any custom daemon listening on a raw socket has never had a home in Checkmeup — you'd have to run your own connectivity check or fall back to pinging a health endpoint that may not exist. Port monitors close that gap: a raw TCP connect to a host and port, no HTTP request, no response body, on the same interval/status/incident/alert model as every other monitor type.",
  },
  {
    type: 'h3',
    text: 'Adding a port monitor',
  },
  {
    type: 'p',
    text: 'Give it a name, a host, and a port (1–65535), and the first check runs within one interval of creation — same pattern as uptime monitoring. Unlike SSL and domain monitors, the host and port are editable after creation rather than requiring delete-and-recreate; a TCP endpoint changes for more mundane reasons (a service moves to a new port) than a certificate\'s hostname does. Port monitors count toward the org\'s aggregate monitor limit alongside cron, uptime, SSL, and domain — the plan limits table now reads "cron + uptime + SSL + domain + port" — and plug into maintenance windows and status pages the same way the other four types do.',
  },
  {
    type: 'h3',
    text: 'Expected state: open vs. closed',
  },
  {
    type: 'p',
    text: 'Every port monitor has an expected state, defaulting to "open" — the familiar case: alert if the port stops accepting connections. The second option, "closed," inverts that: alert if the port unexpectedly starts accepting connections. That\'s not an uptime check anymore, it\'s a security check — confirming that a port which should be firewalled off (a database bound to a public interface, an admin panel, a debug port left on by mistake) actually stays unreachable, and catching the exact moment it doesn\'t. Both modes share the same check, the same worker loop, and the same alert plumbing; only the interpretation of a successful connect flips.',
  },
  {
    type: 'h3',
    text: 'How the check works',
  },
  {
    type: 'p',
    text: "A plain TCP dial with a 10-second timeout, closed immediately on success — no data sent or received, no protocol handshake, so it works identically for SMTP, Postgres, or anything else listening on a socket. Connect time is recorded on every check the same way uptime monitoring records response time, and the check log shows the same up/down/reason shape. Down transitions follow the same alert-after-N-failures filter as every other monitor type (default: alert on the first failure), so a flapping port isn't treated any differently than a flapping HTTP endpoint.",
  },
  {
    type: 'h3',
    text: 'Alerts',
  },
  {
    type: 'p',
    text: 'Alert wording adapts to the expected state: an open-state monitor going down reads the familiar "is down," while a closed-state monitor reads "port unexpectedly open" — because that\'s the actual problem, and "down" would be a confusing thing to say about a port that just became reachable. Recovery alerts, the per-incident cap, and every notification channel (Telegram, email, Slack, webhook) work exactly as they do for the other four monitor types.',
  },
  {
    type: 'h3',
    text: "What's not in this release",
  },
  {
    type: 'p',
    text: 'A DNS or host-resolution failure is treated identically to a connection refusal or timeout — both simply mean "couldn\'t connect." A distinct "error" state for an unresolvable host (the way SSL and domain monitors distinguish "error" from "expired") wasn\'t built: `monitor_status` for port monitors reuses the same waiting/up/down/paused enum as uptime monitoring, which has no error state to put it in, and reliably telling a DNS failure apart from a refused connection needs an extra lookup call that didn\'t seem worth the complexity for a first release. Reasonable follow-up if it turns out to matter in practice.',
  },
  {
    type: 'h2',
    text: 'Follow along',
  },
  {
    type: 'p',
    text: 'Microsoft Teams alerts are next on the board. Releases land on this blog as they ship; the GitHub repo has the full commit history and architecture decision records if you want the why behind any of this.',
  },
  {
    type: 'signature',
    text: '— Andrew',
  },
]
