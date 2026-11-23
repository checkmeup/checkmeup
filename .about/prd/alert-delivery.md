# Alert delivery: no alert lost to a failed send or a dead worker

## Problem

A customer pays Checkmeup to hear when something is down, and today a down alert is tried once. When every channel and the Fallback email fail (a Resend or Twilio outage, a database blip while reading the channels), or the process stops mid-send (every merge deploys, and a crash), the alert is gone: the incident stays open with nothing sent, and nothing tries again. And when the worker itself hangs or dies, every monitor goes silent while `/api/v1/health` still answers "ok", so nobody, the operator included, finds out.

A monitor's state moves on in several statements, and anything that runs between them leaves it wrong. A monitor marked up whose incident then fails to resolve keeps that incident open for good, with no recovery alert, and the next outage opens a second one beside it. A cron ping that lands while the worker is marking its monitor down brings a down alert for a job that just ran, or leaves an incident open on a monitor that's up. A pause made while a check is running is overwritten by the check, so the monitor silently resumes. Two processes during a deploy check the same monitor and can open two incidents for one outage. An SSL or domain check whose write fails sends its alert anyway, and again on every tick while the write keeps failing. And a cron monitor's "skip the first N failures" counts 30-second ticks, not missed runs, so it alerts a minute or two after the first missed run whatever N says.

## Users

Every org with a monitor: the alerts are theirs. The operator, for the worker's own health.

## Success signal

- Open incidents with alerts on, on a monitor still down, whose down alert never went out: from staying so forever to none older than 10 minutes, except while every channel and the Fallback email keep failing.
- A hung or dead worker: from unnoticed to the operator alerted within 10 minutes.
- Open incidents on a monitor that is up or paused, and monitors with more than one open incident: from possible to none.

## Scope

- A down alert (uptime, port, DNS, cron), an SSL error alert or SSL expiry warning, and a domain expiry warning or expired alert that did not go out, for any reason, is sent again every 5 minutes until it goes out, while what it reports still holds.
- The time of an alert's last attempt is kept in PostgreSQL, so a retry, a restarted process and the two processes running side by side during a deploy all see it.
- `/api/v1/health`, and `/up` with it, fails when the worker has stopped finishing ticks.
- The operator points an external uptime service at `/api/v1/health` (a step outside the code).
- ADR alerting changed with it: each channel had one attempt per alert, never a retry, and an alert still sending when the process stopped was lost.
- A monitor's status and its incident change together, in one statement, and only from the state the change expects: a recovery resolves the open incidents as it marks the monitor up, an incident opens only on a monitor that is up or waiting, and no check result overwrites a pause. The database holds at most one open incident per monitor (a unique partial index), and a migration first resolves the open incidents the older code left on monitors that aren't down, and all but the newest where a monitor has several.
- An SSL or domain check whose result can't be written sends no alert; the next tick checks it again.
- A cron monitor's alert filter counts missed runs.
- A cron ping is never lost to a failed write of its history row, and a database error never answers it 404. A ping whose monitor can't be read at all is still lost: there is nothing to record it on.
- A stuck run of a paused cron monitor doesn't alert.
- An alert retried this way whose only delivery was an SMS that then fails at the carrier, and whose Fallback email fails too, goes back to its retry: the SMS keeps the alert's incident, or its SSL or domain monitor, so its status callback can undo the delivery.

## Non-goals

- Retrying recovery and stuck-run alerts: losing one leaves a problem reported but not its end, never silence during a problem. They stay one-shot.
- Exactly-once delivery: an alert that went out but wasn't recorded (a restart right after sending) may go out twice. Twice is better than never.
- Retrying a single channel that failed when another one, or the Fallback email, delivered: as today, the alert counts as delivered once anything delivered it (PRD notification-channels), and Settings shows the failed channel's last delivery.
- Treating an SMS as undelivered until Twilio's status callback: as today, an SMS Twilio accepted counts as delivered, and a later carrier failure sends the Fallback email (ADR alerting); only when that fails too does the alert go back to its retry (see the acceptance criteria).
- A status page or customer-facing notice of Checkmeup's own outage.
- Claiming due monitors so two processes during a deploy never check the same one: a duplicate check row for those few seconds is harmless once an incident can't open twice.

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
- A down alert (uptime, port, DNS, cron), an SSL error alert or SSL warning, or a domain warning or expired alert, whose only delivery was an SMS that Twilio's status callback then reports failed, and whose Fallback email fails too, is attempted again as an undelivered one is: on the first tick at least 5 minutes after its last attempt, and every 5 minutes after, while what it reports still holds, until something delivers it, and not again after. Its failed SMS count toward the 3 tries of its SMS channel. One whose Fallback email delivered isn't retried. A callback that arrives before the alert was counted delivered hands it back on the next tick. A hand-back never takes back an attempt made after the SMS, nor replaces a newer SSL or domain alert. A recovery or stuck-run SMS in the same case stays one-shot.
  → test: apps/api/internal/worker/alert_retry_test.go "TestRetryAfterFailedSMS/a down alert whose only SMS and Fallback email fail is sent again 5 minutes later, until delivered, then not again"
  → test: apps/api/internal/worker/alert_retry_test.go "TestRetryAfterFailedSMS/every incident-tracked type hands a failed only-delivery SMS back to its retry"
  → test: apps/api/internal/worker/alert_retry_test.go "TestRetryAfterFailedSMS/a cron missed-run alert whose only SMS and Fallback email fail is sent again"
  → test: apps/api/internal/worker/alert_retry_test.go "TestRetryAfterFailedSMS/an SSL error alert whose only SMS and Fallback email fail is sent again 5 minutes later, until delivered"
  → test: apps/api/internal/worker/alert_retry_test.go "TestRetryAfterFailedSMS/a domain warning whose only SMS and Fallback email fail is sent again 5 minutes later, until delivered"
  → test: apps/api/internal/worker/alert_retry_test.go "TestRetryAfterFailedSMS/an alert handed back tries its SMS at most 3 times, the failed ones counted"
  → test: apps/api/internal/worker/alert_retry_test.go "TestRetryAfterFailedSMS/a failed SMS whose Fallback email delivered isn't retried"
  → test: apps/api/internal/worker/alert_retry_test.go "TestRetryAfterFailedSMS/a failed SMS on an incident no longer open hands nothing back"
  → test: apps/api/internal/worker/alert_retry_test.go "TestRetryAfterFailedSMS/a failure callback that arrives before the alert is counted undoes it on the next tick"
  → test: apps/api/internal/worker/alert_retry_test.go "TestRetryAfterFailedSMS/an undo never takes back a delivery made after the SMS"
  → test: apps/api/internal/worker/alert_retry_test.go "TestRetryAfterFailedSMS/a handed-back SSL or domain alert never replaces a newer one"
  → test: apps/api/internal/worker/alert_retry_test.go "TestRetryAfterFailedSMS/a failed recovery or stuck-run SMS stays one-shot"
  → test: apps/api/internal/worker/worker_test.go "TestDispatchAlert/an accepted sms keeps its alert's retry source only when the send is retried"
- `/api/v1/health` and `/up` answer 503 with the code `worker_stalled` when no worker tick has finished in the last 2 minutes, with the first 2 minutes after the process starts counting as finished (so a deploy isn't held up), and 200 otherwise. A database that doesn't answer still gives `db_unavailable`, checked first. A tick slower than 2 minutes, which delays every check by as much, gives 503 too, on purpose.
  → test: apps/api/internal/server/server_test.go "TestHandleHealth_WorkerStalled"
  → test: apps/api/internal/server/server_test.go "TestHandleHealth_DBUnavailable"
- An external uptime service checks `https://checkmeup.net/api/v1/health` at least every 5 minutes and emails the operator when it fails (checked by hand once: the service shows the check as up, and its test alert reaches the operator's inbox).
  → set up: a Better Stack monitor on `https://checkmeup.net/api/v1/health`, checked every 3 minutes, alerting by email; its test alert reached the operator, 2026-11-23
- An uptime, port or DNS check that passes on a down monitor, or a ping to a down cron monitor, marks it up and resolves its open incidents in one statement: when either write fails (a test injects the failure with a trigger that raises), neither is done, the monitor stays down, no recovery alert goes out, and the next passing check or ping recovers it. A passing check or a ping on a monitor that isn't down but has an open incident resolves that incident too, silently: its downtime is unknown, and a recovery after days would mislead.
  → test: apps/api/internal/worker/worker_test.go "TestRecoveryResolvesIncidentAtomically"
  → test: apps/api/internal/handler/ping_test.go "TestReceivePing/a ping whose incident fails to resolve leaves the monitor down for the next ping"
- An incident opens only on a monitor that is up or waiting, so two checks of one failing monitor, as from two processes during a deploy, open one incident and send one down alert; the second finds the monitor down and does nothing more. Inserting a second open incident for one monitor fails in the database.
  → test: apps/api/internal/worker/worker_test.go "TestIncidentOpensOnce"
- A cron monitor's missed run is counted, and the monitor marked down, only while it is still up and overdue: a ping recorded after the worker read it as overdue leaves it up, its count and next expected ping as the ping set them, with no incident and no alert.
  → test: apps/api/internal/worker/worker_test.go "TestCheckOverdue/a ping after the worker read the monitor as overdue wins"
- A check that finishes after its monitor was paused writes nothing to the monitor (status, failure count, schedule, and for SSL and domain monitors expiry, issuer and alert flags) and to its incidents, and sends no alert, for every monitor type; an uptime, port or DNS check row itself may be kept.
  → test: apps/api/internal/worker/worker_test.go "TestCheckFinishingAfterPause"
  → test: apps/api/internal/worker/worker_test.go "TestExpiryCheckAlertsOnlyOnceWritten"
- An SSL or domain check whose result fails to write sends no alert, recovery included, and leaves the monitor due, so the next tick checks it again; one alert goes out once a write succeeds.
  → test: apps/api/internal/worker/worker_test.go "TestExpiryCheckAlertsOnlyOnceWritten"
- A cron monitor set to skip the first N failures alerts on the (N+1)th missed run in a row: each missed run counts once, and the monitor is then expected at its next scheduled run after the tick that counted it, plus the grace period (as a ping at that moment would set it). Runs missed while the worker was down count as one. A ping resets the count.
  → test: apps/api/internal/worker/worker_test.go "TestCheckOverdue/the alert filter skips missed runs, not ticks"
- A ping whose history row fails to write still marks its monitor up and recovers it, and a ping that fails to read its monitor for any reason but an unknown token answers 200, not 404.
  → test: apps/api/internal/handler/ping_test.go "TestReceivePing/a ping whose history row fails to write still recovers the monitor"
  → test: apps/api/internal/handler/ping_test.go "TestReceivePing/a ping that can't read its monitor answers 200, not 404"
- A run of a paused cron monitor that passes its maximum duration sends no stuck-run alert while the monitor is paused; once it is resumed, a run still open alerts then.
  → test: apps/api/internal/worker/worker_test.go "TestCheckStuckCronRuns/a paused monitor's stuck run alerts only once the monitor is resumed"
