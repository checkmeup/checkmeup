# Projects: group monitors by client or project

## Problem

Every monitor sits in one flat list. A freelancer or agency with a dozen clients, each with an uptime check, an SSL certificate, a domain and a nightly backup, can't see one client's monitors together, and an alert doesn't say whose site it is about.

## Users

Freelancers and small agencies running monitors for several clients (ADR product); also anyone with several projects of their own.

## Success signal

A user with monitors for several clients can see each client's monitors as one group, find a client's problem at a glance, and tell from an alert which client it concerns without opening the app.

## Scope

- A project has a name, unique within the org. Creating, renaming, deleting and listing projects on a `/projects` page.
- Each monitor, of any of the six types, belongs to at most one project. It is set on the monitor's form, where a new project can also be created, or by assigning several monitors to a project at once.
- Deleting a project keeps its monitors running, without a project.
- The `/monitors` list groups monitors by project, with a summary of what needs attention in each group and a "No project" group last. A `?project=` filter works alongside `?type=`.
- The dashboard names the project of each monitor that needs attention.
- Every alert for a monitor in a project starts its headline with the project name in brackets, `[Acme]`, on every channel. Webhook events gain `projectId` and `projectName`. An alert for a monitor without a project is unchanged.
- No per-plan limit on projects. A flat cap of 1000 per org, the same on every plan (ADR limits).

## Non-goals

- A monitor in several projects, nested projects, or tags (ADR projects).
- Per-project settings: default notification channels, maintenance windows by project, creating a status page from a project. These are a later step with their own PRD.
- Per-project access for team members (PRD team-management).
- Projects in the public API (`X-API-Key`).

## Acceptance criteria

- A project needs a name; a second project with the same name in the org is refused.
- Another org's project can't be read, renamed, deleted or assigned, and another org's monitor can't be put into a project.
- A monitor of each type can be put into a project and taken out of it; the change shows in the list.
- Several monitors of different types can be assigned to a project in one action.
- Deleting a project leaves its monitors in place, without a project, still checked.
- The `/monitors` list shows one group per project plus "No project", each with how many of its monitors need attention; `?project=` shows one project's monitors.
- An alert for a monitor in project "Acme" reads `[Acme]` before the monitor name on Telegram, email (body and subject), Slack and SMS, and its webhook event carries `projectId` and `projectName`.
- An alert for a monitor without a project is byte for byte what it was before projects.
- A 1001st project in one org is refused with `409 too_many_projects` (ADR limits).
