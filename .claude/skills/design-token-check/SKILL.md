---
name: design-token-check
description: Scan apps/web/src for the two recurring color-token mistakes in this repo — Tailwind's bg-[--token] (compiles to invalid CSS, silently drops the style) and hardcoded hex values instead of design tokens. Use when asked to "check for hardcoded colors", "audit design tokens", or before/after any styling change, since these bugs are silent (no lint error, no visual diff in devtools until you inspect computed styles).
---

# Design token check

Two known failure modes, both from CLAUDE.md's Don't section and EP-10:

1. **`bg-[--token]` / `text-[--token]` / `border-[--token]`** — in this
   Tailwind v4 setup, this compiles to invalid CSS (`background-color:
   --token`, missing `var()`) and the style is silently dropped. The
   correct form is always `bg-[var(--token)]`. Found broken across
   `Button.vue`/`Input.vue`/`Label.vue`/`LandingLayout.vue` during EP-10 —
   the sign-in button had no background in *either* theme until fixed.
2. **Hardcoded hex instead of a token** — breaks automatically on theme
   switch (`data-theme` on `<html>`) since a literal hex doesn't flip
   between light/dark.

## Steps

**1. Check for the `bg-[--token]` anti-pattern** (should always be zero
hits — this was fully fixed in EP-10):

```bash
cd apps/web/src
grep -rnE "(bg|text|border|from|to|via|ring|fill|stroke)-\[--[a-z-]+\]" --include="*.vue" . | grep -v "var(--"
```

Any hit is a real bug: the token reference is missing `var(...)`. Fix by
wrapping it: `bg-[var(--token)]`.

**2. Check for hardcoded hex outside `style.css`** (where tokens are
*defined*, so hex there is expected):

```bash
grep -rnE "#[0-9a-fA-F]{3,8}\b" --include="*.vue" . | grep -v "style.css"
```

**Not every hit is a bug** — verify each one against
`docs/reference/design.md`'s token table before flagging it:

- A hex value that matches a documented token (e.g. `#1D9E75` ==
  `--green-500`) should be replaced with the token.
- A hex value with **no** token match may be a deliberate fixed color —
  e.g. `HomeView.vue`'s macOS-traffic-light mockup dots
  (`#ff5f57`/`#febc2e`/`#28c840`) are supposed to look like real macOS
  chrome in both themes, not shift with `--status-*` tokens. Don't
  "fix" those; note them as intentional.

**3. For real hits**, replace with the matching token from
`docs/reference/design.md` (36 tokens documented — brand greens, status
colors, and per-theme neutrals/surfaces/accents), using the
`bg-[var(--token)]` / `text-[var(--token)]` Tailwind form, never a raw
CSS custom-property reference.

**4. Verify the fix actually renders** in both themes — toggle
`data-theme` on `<html>` via devtools or the in-app theme toggle. A
missing `var()` produces no console error and no lint failure, so
grep + a visual check is the only reliable catch.
