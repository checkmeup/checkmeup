# EP-06: Status page

Public status pages served at `checkmeup.net/status/{slug}`. No subdomain, no custom domain on MVP (see [ADR-005](../decisions/005-status-page-same-domain.md)).

---

### US-0601: Create a status page

**As a** user, **I want** to create a public status page **so that** my users can check service health without contacting me.

**Acceptance criteria:**

- [x] Fields: page name, slug (URL-safe, globally unique)
- [x] Slug validated: lowercase letters, numbers, hyphens only; 3–48 chars
- [x] Slug unavailable message shown in real-time if taken
- [x] Page created empty — no monitors attached yet
- [x] Public URL shown immediately after creation

---

### US-0602: Add monitors to the status page

**As a** user, **I want** to choose which monitors appear on my status page **so that** I control what is public.

**Acceptance criteria:**

- [x] Multi-select from the org's monitors (all types)
- [x] Custom display name per monitor (defaults to monitor name)
- [x] Display order configurable (up/down arrows on MVP; drag-and-drop post-MVP)

---

### US-0603: View the public status page

**As a** visitor, **I want** to see the current status of a service's components **so that** I know if there's an incident affecting me.

**Acceptance criteria:**

- [x] Publicly accessible — no login required
- [x] Overall banner: "All systems operational" / "Partial outage" / "Major outage"
- [x] Each monitor shows: display name, current status, 90-day uptime bar
- [x] Last updated timestamp visible
- [x] Page renders correctly without JavaScript (SSR via Go html/template — see [ADR-017](../decisions/017-status-page-ssr.md))

---

### US-0604: Customise the status page

**As a** user, **I want** to add my name and description to the status page **so that** it represents my service, not checkmeup.

**Acceptance criteria:**

- [x] Editable: page title, description (shown below the title)
- [x] Logo: URL input on MVP; file upload post-MVP
- [x] Layout: choice of "classic" (single-column) or "grid" (monitor grid + incident sidebar), per page — see [ADR-038](../decisions/038-status-page-layout-option.md)

---

### US-0605: Delete a status page

**As a** user, **I want** to delete a status page **so that** it's no longer publicly accessible.

**Acceptance criteria:**

- [x] Confirmation dialog required
- [x] Public URL returns 404 immediately after deletion
- [x] Monitors themselves are not deleted — only removed from this page
