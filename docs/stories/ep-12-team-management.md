# EP-12: Team management

Today an org has exactly one user, created together at sign-up ([EP-01](ep-01-auth.md)) — there's no way to add a second person. This epic adds inviting teammates into the same org with a restricted role, matching "Multi-user organizations — invite teammates, shared monitors" already promised on the blog's roadmap.

Invite-conflict handling decided in [ADR-031](../decisions/031-team-invite-conflict.md): reject the invite when the email already belongs to a user in any org — no multi-org-per-user support. `users.email` stays globally unique and a user still belongs to exactly one org via `users.org_id` (`apps/api/migrations/001_initial.sql`); no `org_members` table.

---

### US-1201: Invite a teammate by email

**As an** org owner, **I want** to invite a teammate by email **so that** they can access our shared monitors without sharing a login.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [ ] "Invite" action in Settings — takes an email and a role (see US-1202)
- [ ] Sends an invite email via Resend with a single-use, time-limited link (7-day TTL)
- [ ] Rejected with a clear error if the email already belongs to a user in any org (see the design decision above)
- [ ] Pending invites listed in Settings with resend/revoke actions

---

### US-1202: Owner and Member roles

**As an** org owner, **I want** teammates to have a restricted role by default **so that** they can't change billing or remove the org.

**Estimate:** 1.5 h

**Acceptance criteria:**

- [ ] Two roles for MVP: Owner (full access, including billing and team management) and Member (monitors and status pages only)
- [ ] Org always has exactly one Owner — the user who signed up; ownership transfer is out of scope for MVP
- [ ] Role enforced server-side on every billing/team endpoint, not just hidden in the UI

---

### US-1203: Accept an invite

**As an** invited teammate, **I want** to accept an invite and set a password **so that** I can sign in and access the org's monitors.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] Invite link opens a "set password" form, pre-filled with the invited email (read-only)
- [ ] On submit: creates the user under the inviting org with the assigned role, signs them in immediately
- [ ] Expired or already-used invite links show a clear error — no retry on the same link
- [ ] Invite is single-use — accepting marks it consumed even if the link is opened again

---

### US-1204: View and remove team members

**As an** org owner, **I want** to see everyone with access to my org and remove access when needed **so that** the team list stays accurate.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] List shows: name/email, role, joined date, status (active / invite pending)
- [ ] Owner can remove a Member — immediately revokes their sessions (existing refresh tokens revoked, see [ADR-003](../decisions/003-auth-jwt-httponly-cookie.md))
- [ ] Owner cannot remove themselves through this flow (no self-service ownership transfer or org deletion here)
- [ ] Team size capped per plan tier, same enforcement pattern as monitors/status pages ([ADR-019](../decisions/019-plan-limits.md))

---

### US-1205: Member-scoped access

**As a** Member, **I want** to use shared monitors and status pages without seeing billing or team settings **so that** the product reflects my role.

**Estimate:** 1 h

**Acceptance criteria:**

- [ ] Members can create/edit/delete monitors and status pages within the org, same as Owner — monitors are shared, not partitioned per user
- [ ] Billing and Team settings pages return 403 (API) / are hidden (UI) for Members
- [ ] Every existing `org_id`-scoped query applies unchanged — a Member sees exactly what the Owner sees
