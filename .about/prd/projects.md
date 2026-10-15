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
  → test: apps/api/internal/handler/projects_test.go "TestCreateProject/a name is required and at most 100 characters"
  → test: apps/api/internal/handler/projects_test.go "TestCreateProject/a second project with the same name, in any case, is refused"
  → test: apps/api/internal/handler/projects_test.go "TestRenameProject/renaming onto another project's name is refused"
  → test: apps/web/src/components/ProjectField.test.ts "needs a name, and Cancel goes back to the list"
  → test: apps/web/src/components/ProjectField.test.ts "shows why a new project was refused and keeps the name field open"
- Another org's project can't be read, renamed, deleted or assigned, and another org's monitor can't be put into a project.
  → test: apps/api/internal/handler/projects_test.go "TestListProjects/lists only the org's projects, by name, with monitor counts"
  → test: apps/api/internal/handler/projects_test.go "TestRenameProject/another org's project can't be renamed"
  → test: apps/api/internal/handler/projects_test.go "TestDeleteProject/another org's project can't be deleted"
  → test: apps/api/internal/handler/projects_test.go "TestMonitorProjectField/cron monitor can't be created in another org's project"
  → test: apps/api/internal/handler/projects_test.go "TestSetProjectMonitors/monitors can't be put into another org's project"
  → test: apps/api/internal/handler/projects_test.go "TestSetProjectMonitors/another org's monitor can't be put into a project, and nothing changes"
  → test: apps/api/internal/handler/projects_test.go "TestProjectForeignKeyIsOrgScoped"
- A monitor of each type can be put into a project and taken out of it; the change shows in the list.
  → test: apps/api/internal/handler/projects_test.go "TestMonitorProjectField/cron monitor is created without a project by default"
  → test: apps/api/internal/handler/projects_test.go "TestMonitorProjectField/cron monitor can be created in a project"
  → test: apps/api/internal/handler/projects_test.go "TestMonitorProjectField/an update with projectId null takes the monitor out of its project"
  → test: apps/api/internal/handler/projects_test.go "TestMonitorProjectField/an update without projectId keeps the project"
  → test: apps/api/internal/handler/projects_test.go "TestSetProjectMonitors/assigns monitors of every type in one request, and takes them out again"
  → test: apps/web/src/components/ProjectField.test.ts "selects a project, and \"No project\" clears it"
  → test: apps/web/src/components/ProjectField.test.ts "creates a new project from the form and selects it"
  → test: apps/web/src/views/CronMonitorEditView.test.ts "keeps the monitor's project, and saves a new one picked on the form (PRD projects)"
- Several monitors of different types can be assigned to a project in one action.
  → test: apps/api/internal/handler/projects_test.go "TestSetProjectMonitors/assigns monitors of every type in one request, and takes them out again"
  → test: apps/api/internal/handler/projects_test.go "TestSetProjectMonitors/replaces the project's monitors and moves a monitor out of its old project"
  → test: apps/web/src/views/ProjectListView.test.ts "opens the monitor picker with the project's monitors and saves the new selection"
- Deleting a project leaves its monitors in place, without a project, still checked.
  → test: apps/api/internal/handler/projects_test.go "TestDeleteProject/deleting a project leaves its monitors in place, without a project"
  → test: apps/web/src/views/ProjectListView.test.ts "deletes a project after confirming, saying its monitors stay"
- The `/monitors` list shows one group per project plus "No project", each with how many of its monitors need attention; `?project=` shows one project's monitors.
  → test: apps/web/src/lib/monitorRows.test.ts "groups rows by project name, \"No project\" last, leaving out empty projects"
  → test: apps/web/src/lib/monitorRows.test.ts "counts the monitors that need attention in each group"
  → test: apps/web/src/views/MonitorListView.test.ts "groups monitors by project with \"No project\" last and a summary of what needs attention"
  → test: apps/web/src/views/MonitorListView.test.ts "shows only one project's monitors, ungrouped, for ?project="
  → test: apps/web/src/views/MonitorListView.test.ts "shows the monitors without a project for ?project=none"
- An alert for a monitor in project "Acme" reads `[Acme]` before the monitor name on Telegram, email (body and subject), Slack and SMS, and its webhook event carries `projectId` and `projectName`.
  → test: apps/api/internal/worker/worker_project_test.go "TestProjectAlert_NamesTheProjectOnEveryChannel"
  → test: apps/api/internal/worker/worker_project_test.go "TestProjectAlert_SlackHeader"
  → test: apps/api/internal/worker/worker_project_test.go "TestProjectAlert_EscapesTheProjectNameInHTML"
  → test: apps/api/internal/worker/worker_project_test.go "TestDispatchAlert_Project/a monitor in a project sends its project on the webhook event"
- An alert for a monitor without a project is byte for byte what it was before projects.
  → test: apps/api/internal/worker/worker_project_test.go "TestDispatchAlert_Project/a monitor without a project gets the message exactly as built"
  → test: apps/api/internal/worker/worker_project_test.go "TestDispatchAlert_Project/a monitor without a project sends the event without project fields"
- A 1001st project in one org is refused with `409 too_many_projects` (ADR limits).
  → test: apps/api/internal/handler/projects_test.go "TestCreateProject/the 1001st project is refused"
