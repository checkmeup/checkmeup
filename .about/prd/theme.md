# Light and dark theme

## Problem

The app was dark-only; some people work better, or in brighter rooms, with a light interface.

## Users

Everyone using the app or the marketing site.

## Success signal

The chosen theme applies everywhere instantly and is still there on the next visit, with no flash of the other one.

## Scope

- A light palette for every neutral design token; brand and status colors the same in both themes.
- A switch in Settings, in the app shell and in the marketing header.
- The choice kept per browser; first-time visitors follow the OS setting.

## Non-goals

- A dark public status page: it has its own fixed light design.
- Syncing the preference across devices.

## Acceptance criteria

- Settings shows the current theme and switches it on click, without a reload.
  → test: apps/web/src/views/SettingsView.test.ts "renders the appearance section with the current theme highlighted"
  → test: apps/web/src/views/SettingsView.test.ts "switches the theme when a theme button is clicked"
- A returning visitor gets their saved theme before the first paint.
- A first-time visitor gets the theme matching their OS's `prefers-color-scheme`.
