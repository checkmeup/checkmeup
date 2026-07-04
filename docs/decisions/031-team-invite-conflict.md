# ADR-031: Team invites — reject email already registered under another org (no multi-org membership)

**Date:** 2026-07-04
**Status:** Accepted

---

## Context

[EP-12](../stories/ep-12-team-management.md) adds inviting a teammate into an org. Today `users.email` is globally unique and a user belongs to exactly one org via `users.org_id` — there's no `org_members` join table ([`001_initial.sql`](../../apps/api/migrations/001_initial.sql)). Per the [decision backlog](backlog.md), the open question: what happens when an invited email already has an account under a *different* org — reject the invite outright, or support one user belonging to multiple orgs (a materially bigger schema change: an `org_members` join table plus a role column, moving away from the current 1:1 `users.org_id`, an "active org" concept threaded through the session/JWT, an org switcher in the UI, and re-checking every existing `org_id`-scoped query's assumption that a session maps to exactly one org)?

## Decision

**Reject the invite** when the email already belongs to a user in any org — no multi-org-per-user support. This is what US-1201's acceptance criteria already specified; this ADR formalizes it as an actual, examined decision rather than an unexamined default, and closes the backlog item.

Rationale:

- checkmeup's target users for EP-12 are small teams and solo agencies inviting one or two teammates into their own org — not consultants juggling membership across many unrelated client orgs. There's no validated demand for the multi-org case.
- Supporting multi-org membership is a materially bigger change than EP-12's scope: a new join table, an "active org" concept threaded through the session/JWT and every `org_id`-scoped query (today implicitly one-per-session), and new UI (an org switcher). Building that speculatively contradicts this project's general bias against adding structure ahead of demand — the same reasoning [ADR-023](023-notification-channels.md) used to decline a "channel group" entity, and the same reasoning multi-region checking ([EP-32](../stories/ep-32-multi-region-checking.md)) is deferred behind.
- Keeping the existing 1:1 `users.org_id` model intact means EP-12 is purely additive — no changes to multi-tenancy enforcement ([ADR-002](002-multi-tenancy.md)), no session/JWT changes, and no changes to any existing `org_id`-scoped query.

**User-facing behavior:** inviting an email that already has an account (in any org, including the inviter's own) shows a clear rejection error — "This email is already registered" — rather than silently failing or attempting to merge accounts. There is no path to join a second org with the same email; the invited person would need a different email address to join elsewhere. This is stated as a known limitation, not hidden from the user.

## Consequences

- No schema changes beyond what EP-12 already needed for invites and roles themselves (an `invites` table and the Owner/Member role, per US-1201–US-1205) — no `org_members` table, no active-org/session change.
- **If multi-org-per-user demand ever materializes** (e.g. an agency consultant wanting one login across several client orgs), it requires its own future ADR and is a genuinely bigger lift than a quick follow-up to EP-12. Flagging here so it isn't mistaken for an oversight later.
- Removes the "Multi-user orgs / invite conflict" entry from the [decision backlog](backlog.md). EP-12 is unblocked.
