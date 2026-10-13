# Email alerts

## Problem

Not everyone uses Telegram, and some teams route alerts to a shared mailbox.

## Users

Anyone who wants down and recovery alerts by email, and every org as the fallback channel (ADR alerting).

## Success signal

An email alert arrives within one check cycle with the monitor name and "DOWN" in the subject, so inbox filters can catch it.

## Scope

- Email notification channels, any number of addresses (PRD notification-channels).
- A test email before relying on it.
- Down and recovery emails; the fallback to every org user's email when nothing else delivers.

## Non-goals

- Email digests or summaries.

## Acceptance criteria

- Adding an email channel needs an address.
  → test: apps/api/internal/handler/notification_channels_test.go "TestCreateNotificationChannel/email missing email"
  → test: apps/api/internal/handler/notification_channels_test.go "TestCreateNotificationChannel/creates an email channel"
- "Send test email" sends a test alert; without a Resend key in development it succeeds without sending.
  → test: apps/api/internal/email/email_test.go "TestSendTestAlertEmail"
  → test: apps/api/internal/handler/notification_channels_test.go "TestTestNotificationChannel/email with no Resend API key configured no-ops successfully (ADR-012 dev behavior)"
- A down email names the monitor, its type, the reason and the time, with the monitor name and "DOWN" in the subject; a recovery email names the monitor and the downtime.
  → test: apps/api/internal/email/email_test.go "TestSendAlertEmail"
- An email channel still delivers when another channel on the same monitor fails.
  → test: apps/api/internal/worker/worker_test.go "TestDispatchAlert/email channel succeeds (no-op-safe in dev) even when a telegram channel on the same monitor fails"
- A monitor with no channel, or whose channels all failed, emails every user in the org.
  → test: apps/api/internal/worker/worker_test.go "TestDispatchAlert/no channel attached falls back to emailing every org user (ADR-023)"
  → test: apps/api/internal/worker/worker_test.go "TestDispatchAlert/every attached channel failing falls back to org user email (generalized ADR-032 fallback)"
- Text from a monitored service (a DNS answer, say) is escaped in emails, not injected as HTML.
  → test: apps/api/internal/worker/worker_dns_test.go "TestBuildDNSDownAlert/HTML in a DNS response is escaped in Telegram/email, not injected raw"
