# ADR-025: License changed from MIT to Business Source License 1.1

**Date:** 2026-07-01  
**Status:** Accepted

---

## Context

checkmeup has been MIT-licensed since the repo went public. MIT places no restriction on commercial use — anyone can take the code, self-host it, and resell it as a competing hosted monitoring service under their own brand, with no obligation back to the project. That risk grows with the codebase's surface area (five monitor types, alerting, status pages, billing) rather than shrinks.

The repo being public is a deliberate choice (`docs/design.md`, the About page's "transparency by default" value, the ADR log itself) — it's not up for reconsideration. The question is only whether the *license terms* should keep allowing a verbatim commercial clone.

---

## Alternatives considered

| Option | Public source | Self-host allowed | Blocks a competing SaaS clone | Ruled out because |
|---|---|---|---|---|
| Keep MIT | ✅ | ✅ | ❌ | No protection against a direct commercial fork |
| GPL v3 | ✅ | ✅ | Partial (copyleft, not a use restriction) | Copyleft affects downstream derivative licensing, not commercial hosting directly; also a bigger philosophical shift than needed |
| Fully closed source | ❌ | ❌ | ✅ | Contradicts the project's existing transparency commitment; loses the ADR-log/architecture-in-the-open value proposition |
| Business Source License 1.1 | ✅ | ✅ | ✅ (via Additional Use Grant) | **Chosen** |

BUSL is the model used by MariaDB (which authored it), Sentry, CockroachDB, and HashiCorp's early Terraform/Vault license changes — a known, well-understood pattern for exactly this situation: source-available, self-hostable, but not resellable as a competing hosted product.

---

## Decision

Relicense under **Business Source License 1.1**, with these parameters (`LICENSE.md`):

- **Licensor:** Andrew Molyuk
- **Additional Use Grant:** any use is permitted — including production use and monitoring third parties' (e.g. agency clients') systems — **except** offering the Software, in original or modified form, to third parties as a hosted or managed monitoring service competing with checkmeup.net
- **Change Date:** 2030-07-01 (4 years out)
- **Change License:** Apache License, Version 2.0

On the Change Date, the license automatically converts to Apache 2.0 and the Additional Use Grant's restriction disappears — this isn't a permanent lock, just a multi-year head start.

The About page (`apps/web/src/views/AboutView.vue`) now links to `LICENSE.md` and states the practical effect in plain language, in the same section that already links to the ADR log — consistent with the site's existing "we publish what we know, including the choices that weren't obvious" framing.

---

## Consequences

- Self-hosters, contributors, and agencies using checkmeup for their own or clients' monitoring are unaffected — this is exactly the Additional Use Grant's permitted use.
- A verbatim or lightly-modified commercial clone offered as a hosted service is now a license violation, not just a competitive annoyance.
- GitHub's license detector may not label the repo with a recognized SPDX badge the way MIT did — cosmetic, not functional.
- Any future contributor PR is accepted under BUSL terms, same as the rest of the codebase; no separate CLA was introduced.
- Revisit before 2030-07-01 only if the Change Date/License need adjusting for a new version of the Software (BUSL allows a different Change Date per version) — no action needed otherwise, the conversion is automatic.
