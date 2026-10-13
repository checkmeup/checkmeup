# Team management: invite teammates into an org

## Problem

An org has exactly one user, so a freelancer can't share monitoring with a partner or a client's developer without sharing a password.

## Users

Small teams and solo agencies adding one or two people to their own org. Not consultants who need one login across several orgs (ADR auth).

## Success signal

An owner invites a teammate by email and the teammate is working in the shared org within minutes of opening the link.

## Scope

- Inviting by email with a role, through a single-use link valid for 7 days; resending and revoking pending invites.
- Two roles: Owner (everything, including billing and team) and Member (monitors and status pages).
- Accepting an invite by setting a password.
- Listing and removing members; team size limited per plan.

## Non-goals

- One user in several orgs; an email registered in any org can't be invited (ADR auth).
- Transferring ownership; more than two roles.

## Acceptance criteria

- An owner invites an email with a role; the invitee gets an email with a link valid for 7 days.
- Inviting an email that already has an account, in any org, fails with "This email is already registered".
- Pending invites are listed in Settings and can be resent or revoked.
- Opening a valid invite shows a set-password form with the email filled in and read-only; submitting creates the user in the inviting org with the invited role and signs them in.
- An expired or already-used invite link shows a clear error.
- The member list shows each person's email, role, joining date and whether the invite is still pending.
- Removing a member signs them out everywhere at once; an owner can't remove themselves.
- Inviting past the plan's team size limit gets `402 plan_limit_reached`.
- A member can manage monitors and status pages like the owner and sees the same data; billing and team endpoints return `403` for them and their pages are hidden.
- An org always has exactly one owner, the user who signed up.
