# Maintenance windows

## Problem

Planned work takes services down on purpose; without a way to say so, monitors raise false alarms, record fake incidents and drag uptime down.

## Users

Anyone doing planned or emergency maintenance on monitored services.

## Success signal

During a window nobody is paged, uptime stats stay untouched, and the status page says "Under maintenance".

## Scope

- A window with a title, an optional message, a start and an optional end, covering monitors of any type.
- No checks, incidents or alerts for covered monitors while it is active (ADR workers).
- "Under maintenance" on the public status page.
- Listing upcoming, active and ended windows; editing, ending now and deleting.

## Non-goals

- Recurring windows (ADR workers).

## Acceptance criteria

- A window needs a title, a start, at least one monitor, and an end unless "no end date" is checked; an end must be after the start.
  → test: apps/web/src/views/MaintenanceWindowCreateView.test.ts "shows a validation error when title is missing"
  → test: apps/web/src/views/MaintenanceWindowCreateView.test.ts "does not require an end time when \"no end date\" is checked"
  → test: apps/web/src/views/MaintenanceWindowCreateView.test.ts "shows a validation error when no monitors are selected"
  → test: apps/api/internal/handler/maintenance_test.go "TestCreateMaintenanceWindow/validation"
- Another org's monitor can't be added to a window.
  → test: apps/api/internal/handler/maintenance_test.go "TestCreateMaintenanceWindow/a monitor belonging to another org is rejected (tenant isolation)"
- While a window is active, its monitors of every type are not checked, no incident opens and no alert is sent.
  → test: apps/api/internal/worker/worker_test.go "TestCheckOverdue/a monitor under an active maintenance window is excluded"
  → test: apps/api/internal/worker/worker_test.go "TestCheckUptimeMonitors/a monitor under an active maintenance window is excluded"
  → test: apps/api/internal/worker/worker_test.go "TestCheckSSLMonitors/a monitor under an active maintenance window is excluded"
- The public status page shows covered monitors as "Under maintenance", which doesn't raise the outage banner.
  → test: apps/api/internal/handler/status_public_test.go "TestStatusPublicServeHTTP/an active maintenance window overrides the monitor's real status"
- The list shows each window's status (upcoming, active, ended), dates and monitor count.
  → test: apps/web/src/views/MaintenanceWindowListView.test.ts "renders window rows with status, dates, and monitor counts"
- "End now" closes an open window; it isn't offered for ended ones.
  → test: apps/api/internal/handler/maintenance_test.go "TestEndMaintenanceWindowNow/ends an open-ended window, and ending it again 404s"
  → test: apps/web/src/views/MaintenanceWindowListView.test.ts "hides the \"End now\" button for ended windows"
- Editing replaces the title and monitor list; deleting asks for confirmation and removes the window.
  → test: apps/api/internal/handler/maintenance_test.go "TestUpdateMaintenanceWindow/success replaces title and monitor list"
  → test: apps/web/src/views/MaintenanceWindowEditView.test.ts "deletes the window and navigates back to the list on confirm"
- A 101st window in one org is refused (ADR limits).
  → test: apps/api/internal/handler/maintenance_test.go "TestCreateMaintenanceWindow/rejects creating a 101st window"
- The window's optional message is shown next to "Under maintenance" on the public page.
  → test: apps/api/internal/handler/status_public_test.go "TestStatusPublicServeHTTP/an active maintenance window overrides the monitor's real status"
