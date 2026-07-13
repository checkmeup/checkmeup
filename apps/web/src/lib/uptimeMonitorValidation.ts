// Pure/static validation logic, shared by UptimeMonitorCreateView and
// UptimeMonitorEditView — mirrors validateChannelSaveInput's shape in
// notificationChannelTypes.ts (a plain string return, empty when valid).
export function validateUptimeMonitorForm(name: string, url: string, keyword: string): string {
  if (!name.trim()) {
    return 'Name is required'
  }
  if (!url.trim()) {
    return 'URL is required'
  }
  if (!url.match(/^https?:\/\//)) {
    return 'URL must start with http:// or https://'
  }
  if (keyword.trim().length > 500) {
    return 'Keyword must be 500 characters or fewer'
  }
  return ''
}
