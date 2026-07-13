// Pure/static validation logic, shared by MaintenanceWindowCreateView and
// MaintenanceWindowEditView — mirrors validateChannelSaveInput's shape in
// notificationChannelTypes.ts (a plain string return, empty when valid).
export function validateMaintenanceWindowForm(
  title: string,
  startsAt: string,
  noEnd: boolean,
  endsAt: string,
  monitorCount: number,
): string {
  if (!title.trim()) {
    return 'Title is required'
  }
  if (!startsAt) {
    return 'Start time is required'
  }
  if (!noEnd && !endsAt) {
    return 'End time is required, or check "no end date"'
  }
  if (monitorCount === 0) {
    return 'Select at least one monitor'
  }
  return ''
}
