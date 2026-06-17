# EP-10: Light & dark theme

The app ships dark-only today ([`docs/design.md`](../design.md) neutrals are dark-first). Add a light palette and a switch so users can pick the appearance they prefer.

---

### US-1001: Define a light theme palette

**As a** user, **I want** a light theme available **so that** I can use the app comfortably in bright environments.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] Light-mode values defined for every neutral token in `docs/design.md` (`--bg`, `--surface`, `--surface-raised`, `--border`, `--text-muted`, `--text-dim`, `--text`, `--text-strong`)
- [x] Brand greens and semantic status colors (`--status-up/degraded/down/paused`) stay the same in both themes
- [x] No hardcoded hex values introduced outside the token definitions — also surfaced and fixed real pre-existing bugs along the way: `LandingLayout.vue`'s hardcoded header backdrop and `hover:text-white` nav links (would've broken in light mode); `bg-[--token]`/`text-[--token]` Tailwind arbitrary-value classes compiling to invalid CSS (missing `var()`) in `Button.vue`, `Input.vue`, and `Label.vue` — silently broken in both themes already, just unnoticed (e.g. the sign-in button had no visible background in either theme); and the `HomeView`/`PricingView`/`AboutView` CTA banner gradient/text, fixed with dedicated `--cta-*` tokens so it's an intentional light wash in light mode rather than a washed-out dark box

---

### US-1002: Switch between light and dark theme

**As a** user, **I want** a toggle to switch between light and dark theme **so that** I can pick the appearance I prefer without leaving the page.

**Estimate:** 1 h

**Acceptance criteria:**

- [x] Toggle available from Settings and from a quick-access control in the app shell (also added to the marketing-site header, for free, since it's the same global toggle)
- [x] Switching applies instantly, no page reload
- [x]\* Applies consistently across all authenticated views, marketing pages, and auth pages (sign-in/up, forgot/reset password)

\* The **public status page** turned out to be a separate, server-rendered Go template (`status_public.go`) with its own fixed light palette, hardcoded independently of `docs/design.md`'s tokens — not a dark page that needed a light option. Giving it an actual dark variant would mean designing a second status-page theme from scratch, which is more than "switch my dashboard" implied. Descoped; track separately if there's real demand for a dark status page.

---

### US-1003: Remember theme preference

**As a** returning user, **I want** my theme choice remembered **so that** I don't have to reselect it every visit.

**Estimate:** 0.5 h

**Acceptance criteria:**

- [x] Preference persisted client-side (e.g. `localStorage`) and applied before first paint to avoid a flash of the wrong theme
- [x] First-time visitors default to their OS-level `prefers-color-scheme`
- [x] Preference is per-browser only — no `org_id`-scoped DB column needed for MVP
