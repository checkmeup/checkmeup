import type { BlogPost } from '../types'

export const post: BlogPost = {
  slug: 'v1-18-release-notes',
  title: "v1.18: A Public API, So Checkmeup Isn't the Only Place Status Lives",
  date: 'July 3, 2026',
  readTime: '4 min read',
  excerpt:
    "Every monitor's status has lived exclusively behind the dashboard's session cookie until now. v1.18 adds a read-only public API — generate a key, send it as X-API-Key, and pull a monitor's status into a CI pipeline, an internal ops dashboard, or a physical status LED. Cron pings can now also carry their own tags (a build number, a pass/fail state), which show up in both the dashboard and the API.",
  content: [
    {
      type: 'p',
      text: "Checkmeup has always assumed the only consumer of a monitor's status is a human looking at the dashboard. That assumption broke the moment a third-party integration came up as a real request rather than a hypothetical: someone wants their CI pipeline to report a build's outcome somewhere Checkmeup can see it, and someone else wants a physical LED on a desk to turn red the moment a job stops checking in. Neither of those is a browser with a session cookie, so v1.18's actual work wasn't the status endpoint itself — it was deciding how a non-browser client gets to authenticate at all.",
    },
    {
      type: 'h3',
      text: 'Why this needed its own auth mechanism, not a workaround',
    },
    {
      type: 'p',
      text: "ADR-003 has said since the very first auth commit that the httpOnly session cookie is the only auth mechanism, full stop — no Authorization header, ever. That rule exists for good reasons specific to browsers: an httpOnly cookie can't be read by injected JavaScript, and SameSite=Strict stops another site from riding a victim's ambient session. A CI job has neither a cookie jar nor a DOM for either of those threats to exploit, but the rule as originally written didn't carve out an exception — it just said no. Rather than quietly reinterpreting a hard rule because it was inconvenient, this got an actual amendment (ADR-028): the cookie-only rule is now scoped to browser session auth specifically, and a second, independent mechanism — a dedicated X-API-Key header — exists for everything else. The two are kept deliberately distinct in code, not just in intent, down to using separate request-context keys internally, so a request authenticated one way can never be mistaken for the other.",
    },
    {
      type: 'p',
      text: "Generate a key from Settings → API keys and it's shown exactly once — only its hash is ever stored, same pattern the codebase already uses for passwords and refresh tokens. From there:",
    },
    {
      type: 'code',
      lang: 'bash',
      text: 'curl -H "X-API-Key: cmu_live_..." \\\n  https://checkmeup.net/api/v1/public/monitors/cron/your-monitor-id/status',
    },
    {
      type: 'p',
      text: "Swap cron for uptime, ssl, domain, or port to match the monitor type. Keys are read-only for now and rate-limited to 60 requests/minute per key — deliberately per-key rather than per-org or per-IP, so one leaked or misbehaving key can't hammer the API on anyone else's behalf.",
    },
    {
      type: 'h3',
      text: 'Cron pings can carry their own data now',
    },
    {
      type: 'p',
      text: "The second half of this release is narrower but was the actual trigger for wanting a public API at all: a cron ping can now carry arbitrary query-string data. Add ?build=142&state=success to the existing ping URL and it's stored with that ping — capped at 20 key/value pairs, 64 characters per key, 256 per value, silently truncated past that rather than failing the ping, since a monitoring check-in has to succeed no matter what garbage gets thrown at it. Only the most recent ping's data is kept per monitor; it's a snapshot, not a log. It shows up two places: in the dashboard's existing execution log, next to the ping it came from, and in the public API's cron status response as lastPingMetadata — which is the piece that actually closes the loop for a CI pipeline or a status LED asking \"what happened last time?\" without needing to log into Checkmeup at all.",
    },
    {
      type: 'h2',
      text: 'Also this release',
    },
    {
      type: 'ul',
      items: [
        'Settings has a new API keys card: generate (shown once, copy it or lose it), list with a masked prefix and last-used timestamp, and revoke — which takes effect on the very next request.',
        'The docs site has a new Public API section with the curl example above, the full response shape per monitor type, and the metadata caps written out precisely instead of hand-waved.',
      ],
    },
    {
      type: 'p',
      text: "One thing this release deliberately didn't build: per-key scope (read-only vs. read-write). ADR-028 describes it, but every route under /api/v1/public is a GET today, so there was nothing to scope yet — adding a schema column with no enforcement behind it seemed worse than adding it later when a write endpoint actually exists. It's tracked in the decision backlog specifically so it isn't forgotten and quietly assumed-done by whoever builds that write endpoint.",
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
  ],
}
