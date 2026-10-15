# Projects: grouping monitors

Date: 2026-10-15

A project groups monitors, typically one per client of a freelancer or agency (ADR product). It belongs to an org and has a name unique within it. Each monitor belongs to at most one project, recorded as a nullable `project_id` column on each of the six monitor tables, referencing `projects` with `ON DELETE SET NULL`: deleting a project leaves its monitors running, ungrouped. Assigning a monitor checks that the project is in the monitor's org, like every other tenant-scoped write (ADR database), and the database enforces it too: the foreign key is on `(project_id, org_id)`, and deleting a project nulls only `project_id`.

One project per monitor, not tags: the list groups unambiguously, and a project can later carry settings for its monitors (default channels, maintenance windows, a status page) without asking which of a monitor's several groups wins. The UI calls it "Project": "Client" fits agencies but not developers grouping their own work.

An alert names the project in brackets before the monitor, `[Acme] nightly-backup missed its ping`, on every channel, and the webhook event carries `projectId` and `projectName` next to the existing fields (ADR alerting). The project is looked up once when the alert is dispatched, rather than joined into each monitor type's worker query: alerts are rare, and dispatch is the one place every alert passes through.

Projects aren't limited per plan; they cost nothing to keep. A flat cap of 1000 per org (ADR limits) bounds growth by human action.

## Considered options

- Tags, many per monitor: flexible, but a monitor in three groups has no single place in a grouped list and no single source for per-project settings. Can be added on top of projects if cross-cutting groups (prod/staging) are asked for.
- Projects as workspaces scoping everything (monitors, channels, status pages, API keys), as in Sentry or Vercel: a second tenant boundary on every query, next to `org_id`, for a user base of mostly one-person orgs.
- Grouping by status page: mixes the public page with internal order; internal monitors would have to be published to be grouped, and a monitor can be on several pages or none.
- A polymorphic `project_monitors` join table like `status_page_monitors` (ADR database): one migration instead of six columns, but no foreign key, and every monitor list query would need a join to show its project.
- "Client" or "Group" as the name: see above.

## Consequences

- Six tables gain a column, and each monitor type's create, update and list paths carry `project_id`.
- Per-project access for team members (PRD team-management) has a natural place to attach, without being built now.
- A monitor that should sit in two groups can't, until tags exist.
