# Public status pages

Date: 2026-10-13

A status page lives at `checkmeup.net/status/<slug>`, not on a subdomain, and is rendered by Go's `html/template` in the API process, so it works without JavaScript (curl, feed readers, bots). Its route is registered before the SPA catch-all. Styling is inline CSS in the template, outside the Tailwind pipeline. Custom domains are deferred to Enterprise.

Manual incidents, which a user declares to tell visitors something (degraded performance, say), live in their own tables, separate from the worker's automatic incidents: an incident with a title, severity and status, a join table linking it to monitors of any type, and an append-only list of timestamped updates. Declaring or resolving one never changes a monitor's status; the page's overall banner escalates on either source.

Two settings are per page, since one org can run pages for different audiences: `hide_branding` removes the "Powered by Checkmeup" footer and its links, paid plans only (ADR billing); `layout` is `classic` (one column) or `grid` (two columns with incidents in a sidebar), on every plan, `classic` by default. Both layouts share the same `{{define}}` blocks for monitor and incident cards, so they can't drift apart on a field.

## Considered options

- A subdomain per org: needs a wildcard certificate and per-tenant DNS.
- Rendering the page in the Vue SPA, or through Nuxt SSR: the first needs JavaScript, the second a second server process.
- Adding manual incidents to the automatic incident tables: those are one row per monitor and per check-failure window, and every worker path assumes a row means a detected transition.
- Branding and layout as org-wide settings: one org's pages serve different audiences.
- A client-side layout switch: the page must render without JavaScript.

## Consequences

- The `checkmeup.net` URL is a gap for agencies that want white-label pages, until custom domains ship.
- The public render reads `hide_branding` and `layout` from the page row; enforcing branding on downgrade happens at webhook time, not at render time.
