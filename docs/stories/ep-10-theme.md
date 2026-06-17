# EP-10: Light & dark theme

The app ships dark-only today ([`docs/design.md`](../design.md) neutrals are dark-first). Add a light palette and a switch so users can pick the appearance they prefer.

---

### US-1001: Define a light theme palette

**As a** user, **I want** a light theme available **so that** I can use the app comfortably in bright environments.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] Light-mode values defined for every neutral token in `docs/design.md` (`--bg`, `--surface`, `--surface-raised`, `--border`, `--text-muted`, `--text-dim`, `--text`, `--text-strong`)
- [ ] Brand greens and semantic status colors (`--status-up/degraded/down/paused`) stay the same in both themes
- [ ] No hardcoded hex values introduced outside the token definitions

---

### US-1002: Switch between light and dark theme

**As a** user, **I want** a toggle to switch between light and dark theme **so that** I can pick the appearance I prefer without leaving the page.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] Toggle available from Settings and from a quick-access control in the app shell
- [ ] Switching applies instantly, no page reload
- [ ] Applies consistently across all authenticated views and the public status page

---

### US-1003: Remember theme preference

**As a** returning user, **I want** my theme choice remembered **so that** I don't have to reselect it every visit.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [ ] Preference persisted client-side (e.g. `localStorage`) and applied before first paint to avoid a flash of the wrong theme
- [ ] First-time visitors default to their OS-level `prefers-color-scheme`
- [ ] Preference is per-browser only — no `org_id`-scoped DB column needed for MVP
