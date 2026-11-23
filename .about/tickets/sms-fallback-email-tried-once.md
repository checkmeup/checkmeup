# The Fallback email after a failed only-delivery SMS is tried once

2026-11-24

## Symptom

An alert whose only delivery was an SMS is lost if that SMS fails at the carrier and the Fallback email sent then fails too, as during a Resend outage. Nobody hears about the outage, and nothing tries again. The alert should be sent again every 5 minutes until it goes out, as any other undelivered alert is (PRD alert-delivery).

## Evidence

- `apps/api/internal/worker/worker.go:670` `HandleSMSStatus`: on a `failed` or `undelivered` callback for an SMS stored with a fallback, it calls `dispatchFallbackEmail` once and drops the result.
- `TakeSMSMessage` (`apps/api/queries/sms_messages.sql:5`) deletes the `sms_messages` row as it reads it. The row holds only `org_id`, `channel_id` and the email's subject and body, not the incident or monitor the alert was for.
- The alert already counts as delivered once Twilio accepted the SMS: the incident's `alert_count` is 1, or an SSL or domain monitor's pending `alert` is cleared. So `Claim*AlertRetries` (`apps/api/internal/worker/alert_retry.go`) never picks it up again.
- ADR alerting: "when the SMS was the alert's only delivery, the Fallback email goes out then".

## Acceptance criteria

- An uptime, port, DNS or cron down alert whose only delivery was an SMS: Twilio's callback reports it failed and the Fallback email fails too. The alert is attempted again on the first tick at least 5 minutes later, and every 5 minutes after, until something delivers it, while its incident is open and its monitor down, unmuted and under no maintenance window, as PRD alert-delivery's retried down alerts are.
- An SSL error alert or SSL warning, or a domain warning or expired alert, in the same case is attempted again the same way while the monitor isn't fine again, paused or under a maintenance window, with alerts on.
- Such a retry tries an SMS channel only while the alert's SMS attempts, the failed one counted, stay within 3 (PRD alert-delivery's anti-goal).
- Once a retry delivers the alert, it isn't sent again. A failed SMS whose Fallback email delivered isn't retried either.
- A recovery or stuck-run alert in the same case stays one-shot, as PRD alert-delivery's non-goal says.

## Comments

### 2026-11-24 · cause

The SMS was counted as the alert's delivery when Twilio accepted it, which the PRD allows: a carrier failure was to be covered by the Fallback email. That email is the one send with no retry behind it. Its failure is only logged, and the row that could have tied it back to the alert is already deleted.

### 2026-11-24 · fix option: hand the alert back to its retry

Store the alert's retry source on `sms_messages` (monitor type plus the incident's ID, or the SSL or domain monitor's). When the callback's Fallback email fails, undo the delivery: set the incident's `alert_count` back to 0, or restore the monitor's pending `alert`, with `alert_attempted_at` left as it was. The existing `Claim*AlertRetries` then retries it with the checks they already make (still open, still down, not muted). Costs a migration and the alert text kept on SSL and domain monitors until the SMS's final status, since they clear it on delivery today. It reuses the whole retry path, so this is the one I'd pick.

### 2026-11-24 · fix option: retry the Fallback email on its own

Keep the `sms_messages` row when its Fallback email fails, with an attempt time, and have the worker resend just that email every 5 minutes until it goes out or the row is pruned after 7 days. Smaller, but it can't tell whether the problem still holds, so it may report an outage that has already ended. And it is a second retry mechanism beside `Claim*AlertRetries`.

### 2026-11-24 · chosen: hand the alert back to its retry

Picked over retrying the Fallback email on its own, because the existing retries already check whether the problem still holds. How it works:

- `sms_messages` gets the alert's retry source: monitor type plus the incident's ID (uptime, port, DNS, cron) or the monitor's (SSL, domain). It also gets the alert as JSON for SSL and domain, whose pending `alert` is cleared on delivery. Recovery and stuck-run SMS are stored without a source and stay one-shot.
- The source reaches `sendSMSAlert` explicitly, as a field on `MonitorRef` set only by the sends that retry: `trackDown`, `processOverdueMonitor`, the SSL and domain checks, and `retryAlerts`. It can't be inferred from the monitor: a cron monitor can have an open down incident while a separate stuck-run alert goes out.
- When `HandleSMSStatus`'s Fallback email fails, it undoes the delivery. For an incident: `alert_count = 0` while it's still open. For SSL or domain: the stored alert goes back into the monitor's `alert`, only if that is empty, so a newer one from the next daily check wins. `alert_attempted_at` is left alone, so the retry comes 5 minutes after the last attempt as usual.
- `Claim*AlertRetries` then does the rest: it checks that the problem still holds, and it applies the 3-SMS cap through `alert_attempts`. So at most 3 rounds of "SMS accepted, failed at the carrier, Fallback email failed" happen, then retries go out without the SMS.

### 2026-11-24 · open: callback before the count

`alert_count` is set after `DispatchAlert` returns, and `sendSMSAlert` has already stored the SMS by then. A failure callback arriving in between would undo nothing, and the count would then mark the alert delivered for good. Callbacks take seconds and the count follows within milliseconds, so this is unlikely but possible. Two ways to close it, to settle while planning:

- When the undo finds nothing to undo, keep the `sms_messages` row and retry the undo on Twilio's next callback, or from the worker.
- Count the alert before the SMS goes out, and undo the count if the SMS send itself fails.
