# Alert delivery: no alert lost to a failed send or a dead worker

## Problem

A customer pays Checkmeup to hear when something is down, and today a down alert is tried once. When every channel and the Fallback email fail (a Resend or Twilio outage, a database blip while reading the channels), or the process stops mid-send (every merge deploys, and a crash), the alert is gone: the incident stays open with nothing sent, and nothing tries again. And when the worker itself hangs or dies, every monitor goes silent while `/api/v1/health` still answers "ok", so nobody, the operator included, finds out.

## Users

Every org with a monitor: the alerts are theirs. The operator, for the worker's own health.

## Success signal

- Open incidents with alerts on, on a monitor still down, whose down alert never went out: from staying so forever to none older than 10 minutes, except while every channel and the Fallback email keep failing.
- A hung or dead worker: from unnoticed to the operator alerted within 10 minutes.

## Scope

- A down alert (uptime, port, DNS, cron), an SSL error alert or SSL expiry warning, and a domain expiry warning or expired alert that did not go out, for any reason, is sent again every 5 minutes until it goes out, while what it reports still holds.
- The time of an alert's last attempt is kept in PostgreSQL, so a retry, a restarted process and the two processes running side by side during a deploy all see it.
- `/api/v1/health`, and `/up` with it, fails when the worker has stopped finishing ticks.
- The operator points an external uptime service at `/api/v1/health` (a step outside the code).
- ADR alerting changes with it: today each channel gets one attempt per alert, never a retry, and an alert still sending when the process stops is lost.

## Non-goals

- Retrying recovery and stuck-run alerts: losing one leaves a problem reported but not its end, never silence during a problem. They stay one-shot.
- Exactly-once delivery: an alert that went out but wasn't recorded (a restart right after sending) may go out twice. Twice is better than never.
- Retrying a single channel that failed when another one, or the Fallback email, delivered: as today, the alert counts as delivered once anything delivered it (PRD notification-channels), and Settings shows the failed channel's last delivery.
- Treating an SMS as undelivered until Twilio's status callback: as today, an SMS Twilio accepted counts as delivered, and a later carrier failure sends the Fallback email (ADR alerting).
- Resolving the incident of a monitor paused while down (in the Inbox): such an incident is not retried, since its monitor is not down.
- A status page or customer-facing notice of Checkmeup's own outage.

## Anti-goals

- A provider outage burning a customer's SMS credit: an alert tries each SMS channel at most 3 times in all.
- A customer getting the same down alert again once it went out.
- Retries holding up checks: a tick doesn't wait on an alert being sent (ADR alerting).
- A broker, queue or Redis to do it (ADR workers): what's pending lives in PostgreSQL.

## Acceptance criteria

- An uptime, port, DNS or cron monitor goes down and every attached channel and the Fallback email fail: the down alert is attempted again on the first tick at least 5 minutes after the previous attempt started, and so on, until one attempt delivers it; after that it is not sent again.
  → test: apps/api/internal/worker/alert_retry_test.go "TestRetryDownAlerts/a down alert that didn't go out is sent again 5 minutes later, until delivered, then not again"
  → test: apps/api/internal/worker/alert_retry_test.go "TestRetryDownAlerts/every incident-tracked type retries its undelivered down alert"
- A failed read of the monitor's channels counts as a failed attempt and is retried the same way.
  → test: apps/api/internal/worker/alert_retry_test.go "TestDispatchAlert_ChannelsUnreadableIsNotDelivered"
- An open incident whose down alert was never delivered and whose last attempt started more than 5 minutes ago, or that has none, as after a process stopped mid-send, gets its down alert on the next tick; one whose last attempt started less than 5 minutes ago is not sent a second time, by this process or another.
  → test: apps/api/internal/worker/alert_retry_test.go "TestRetryDownAlerts/every incident-tracked type retries its undelivered down alert"
  → test: apps/api/internal/worker/alert_retry_test.go "TestRetryDownAlerts/a down alert that didn't go out is sent again 5 minutes later, until delivered, then not again"
- A down alert is retried only while its incident is open and its monitor is down, not paused, under no maintenance window and with alerts on; once the incident is resolved it is dropped, and the recovery alert goes out as today. A monitor whose alerts are turned on again while still down gets its down alert then.
  → test: apps/api/internal/worker/alert_retry_test.go "TestRetryDownAlerts/no retry once the problem it reports is gone or muted"
  → test: apps/api/internal/worker/alert_retry_test.go "TestRetryDownAlerts/a down alert opened with alerts off goes out once they are turned on"
- An SMS channel is attempted at most 3 times for one alert; the other channels and the Fallback email keep being retried after that.
  → test: apps/api/internal/worker/alert_retry_test.go "TestRetryMessage_SMSTriedAtMostThreeTimes"
- An SSL error alert or SSL warning, or a domain warning or expired alert, that fails is attempted again on the first tick at least 5 minutes after the previous attempt started, until delivered, while the monitor isn't fine again, paused or under a maintenance window, and with alerts on; a delivered one is not sent again, and it counts once toward `max_alerts_per_incident`, whatever its attempts. The check itself is recorded at once, so the monitor's status doesn't wait for the alert. If the next daily check raises another alert meanwhile, it replaces the undelivered one.
  → test: apps/api/internal/worker/alert_retry_test.go "TestRetryExpiryAndSSLAlerts/an SSL error alert that fails is sent again 5 minutes later, until delivered, and counts once"
  → test: apps/api/internal/worker/alert_retry_test.go "TestRetryExpiryAndSSLAlerts/an SSL alert isn't sent again once the certificate is fine"
  → test: apps/api/internal/worker/alert_retry_test.go "TestRetryExpiryAndSSLAlerts/a domain warning that fails is sent again 5 minutes later, until delivered"
- `/api/v1/health` and `/up` answer 503 with the code `worker_stalled` when no worker tick has finished in the last 2 minutes, with the first 2 minutes after the process starts counting as finished (so a deploy isn't held up), and 200 otherwise. A database that doesn't answer still gives `db_unavailable`, checked first. A tick slower than 2 minutes, which delays every check by as much, gives 503 too, on purpose.
  → test: apps/api/internal/server/server_test.go "TestHandleHealth_WorkerStalled"
  → test: apps/api/internal/server/server_test.go "TestHandleHealth_DBUnavailable"
- An external uptime service checks `https://checkmeup.net/api/v1/health` at least every 5 minutes and emails the operator when it fails (checked by hand once: the service shows the check as up, and its test alert reaches the operator's inbox).
  → set up: a Better Stack monitor on `https://checkmeup.net/api/v1/health`, 2026-11-23
