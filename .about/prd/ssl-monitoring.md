# SSL monitoring: warn before a certificate expires

## Problem

An expired certificate takes a site down with a browser warning, usually discovered by a visitor or a client.

## Users

Developers and freelancers responsible for their own or their clients' domains.

## Success signal

Every certificate is renewed before it expires, with three warnings in hand: 30, 14 and 7 days ahead.

## Scope

- An SSL monitor for a hostname, checking the certificate's issuer, expiry and validity.
- One alert at each of 30, 14 and 7 days before expiry, and at once when it is expired or invalid.
- List, detail, edit, pause and resume, delete.

## Non-goals

- Renewing certificates.
- Checking the whole chain against every client's trust store.

## Acceptance criteria

- A full URL entered as the domain is reduced to its hostname.
  → test: apps/api/internal/handler/ssl_monitors_test.go "TestCreateSSLMonitor/hostname is normalised from a full URL"
- A new monitor starts as "waiting" and the detail page says the first check is pending.
  → test: apps/api/internal/handler/ssl_monitors_test.go "TestCreateSSLMonitor/success defaults to waiting status with alerts enabled"
  → test: apps/web/src/views/SSLMonitorDetailView.test.ts "shows the first-check notice when the monitor has never been checked"
- A check records issuer, expiry and days left; more than 30 days left is up.
  → test: apps/api/internal/worker/worker_test.go "TestBuildSSLCheckResult/more than 30 days left is up and resets alert state"
- An unreachable host or an invalid certificate is recorded as an error, with its message shown on the detail page.
  → test: apps/api/internal/worker/worker_test.go "TestCheckSSLMonitors/a connection failure is recorded as an error status"
  → test: apps/web/src/views/SSLMonitorDetailView.test.ts "shows the error message block when the monitor has an error"
- Crossing 30, 14 or 7 days left sends one alert per threshold, never twice for the same one.
  → test: apps/api/internal/worker/worker_test.go "TestSslCrossedThreshold/crosses the 30-day threshold without touching 14d or 7d"
  → test: apps/api/internal/worker/worker_test.go "TestSslCrossedThreshold/does not re-alert once every threshold up to daysLeft has already fired"
- An expired certificate is marked expired and alerted on at once.
  → test: apps/api/internal/worker/worker_test.go "TestBuildSSLCheckResult/expired sets status expired"
  → test: apps/api/internal/worker/worker_test.go "TestSslThresholdAlert/builds an expired-phrased alert for negative days left"
- The list shows status, expiry date, days left and issuer; a negative day count shows "Expired".
  → test: apps/web/src/views/MonitorListView.test.ts "renders an SSL row with status, expiry, and issuer"
  → test: apps/web/src/views/MonitorListView.test.ts "shows \"Expired\" for a negative days-until-expiry even with an expiry date"
- Editing changes the name and hostname.
  → test: apps/api/internal/handler/ssl_monitors_test.go "TestUpdateSSLMonitor/success applies changes and normalises the hostname"
- Pausing stops alerts; resuming restores them.
  → test: apps/api/internal/handler/ssl_monitors_test.go "TestPauseResumeSSLMonitor/pause then resume round-trips status"
- Deleting asks for confirmation, then removes the monitor and its history.
  → test: apps/web/src/views/SSLMonitorDetailView.test.ts "deletes the monitor and navigates back to the list on confirm"
- Each certificate is checked once a day.
  → test: apps/api/internal/worker/worker_test.go "TestCheckSSLMonitors/a checked certificate is not checked again until a day later"
