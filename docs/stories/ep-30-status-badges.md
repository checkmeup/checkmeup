# EP-30: Public status badges

A small embeddable status image, generated per status page ([EP-06](ep-06-status-page.md)) and per monitor on that page, for dropping into a README, footer, or docs site. Cheap to build (server-rendered SVG, no new infra) and doubles as free marketing — the badge links back to the public status page (see `docs/bucket-list.md`).

---

### US-3001: Generate an overall status badge for a status page

**As a** user, **I want** an embeddable badge showing my status page's overall status **so that** I can put it in my README or site footer without visitors clicking through.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [ ] Public, unauthenticated endpoint, e.g. `GET /status/:slug/badge.svg`
- [ ] SVG generated server-side (Go, no external badge-rendering service — keeps to the no-new-infra principle)
- [ ] Badge text reflects current page status, same wording as the page banner: "operational" / "degraded" / "outage"
- [ ] 404 if the slug doesn't exist or the page was deleted, matching EP-06 US-0605

---

### US-3002: Per-monitor status badge

**As a** user, **I want** a badge for a single monitor on my status page **so that** I can show "API: operational" in a project README without embedding the whole page's status.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] Endpoint per monitor on a page, e.g. `GET /status/:slug/badge/:monitor_id.svg`
- [ ] 404 if the monitor isn't attached to that status page (no leaking monitor status outside its page's visibility)
- [ ] Same visual style and status wording as the page-level badge (US-3001)

---

### US-3003: Copy an embeddable snippet from the UI

**As a** user, **I want** ready-made embed code for my badges **so that** I don't have to hand-build the URL or markup.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [ ] Status page editor (EP-06) gets a "Badges" section listing the page-level badge and one row per attached monitor
- [ ] Live preview of each badge rendered in the UI
- [ ] One-click copy for both Markdown (`![status](...)`) and HTML (`<img>` wrapped in `<a>` linking to the public page) snippet forms

---

### US-3004: Badge caching

**As a** platform, **I want** badge responses to be cacheable **so that** embedding sites and CDNs don't hammer the endpoint on every page load.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] Badge responses set `Cache-Control: max-age=60` (or similar short TTL) — reflects a status change within a minute without unbounded re-generation
- [ ] Badge endpoints are public and subject to the same rate limiting as other public status-page routes ([ADR-013](../decisions/013-rate-limiting.md)), not exempted
