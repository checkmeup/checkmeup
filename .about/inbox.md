# Inbox

## Check Search Console after the changelog release

2026-10-13 · session review

Release notes `noindex`, `/changelog`, the trailing-slash 301 and the path-only canonical shipped on 2026-10-13. Google drops pages only after a recrawl, which can take weeks. The "Page with redirect" count should grow as trailing-slash variants become 301s. The guess that today's 4 redirect pages are the `http://` and `www.` variants was never checked. Settled when Search Console shows the release notes excluded by `noindex`, no trailing-slash duplicates indexed, and the redirect pages accounted for.

## WhatsApp provider

2026-10-13 · decision backlog

PRD whatsapp-alerts needs a provider: Meta's Cloud API directly, or a BSP (Twilio, MessageBird). It decides setup effort, per-message cost and who handles template approval. The down and recovery templates must also be drafted and approved by Meta before the alerts can ship; plan for that lead time. Settled by picking the provider, recorded in ADR alerting.

## Signal alerts: build them at all?

2026-10-13 · decision backlog

Signal has no bot or business API; the only path is self-hosting `signal-cli` with a phone number registered as a bot, which Signal's anti-abuse systems may ban, and which adds an always-on process (ADR workers avoids those). Settled by a go/no-go, ideally from real user demand, recorded in ADR alerting.

## Microsoft Teams webhook mechanism

2026-10-13 · decision backlog

Microsoft is retiring the Office 365 Connector "Incoming Webhook" for Power Automate workflow webhooks (Adaptive Card payload). Settled by re-checking the retirement status and the recommended mechanism right before PRD ms-teams-alerts is built.

## Multi-region check infrastructure

2026-10-13 · decision backlog

PRD multi-region-checks needs compute in more than one region, against one VPS and no broker (ADR workers, ADR infrastructure). Options: extra regional VPSes reporting to the one PostgreSQL, or a check-execution service; and which regions first. Settled by that choice, recorded in ADR workers.

## Read-only scope for API keys

2026-10-13 · decision backlog

`api_keys` has no `scope` column and `RequireAPIKey` checks no read-only/read-write scope (ADR auth). Harmless while every `/api/v1/public` route is a `GET`. Settled when the `scope` column and a `403` for read-only keys on non-`GET` requests land, before the first write endpoint.

## CDN or WAF in front of the VPS

2026-10-13 · decision backlog

Nothing between the internet and kamal-proxy absorbs a volumetric attack; per-route rate limits (ADR limits) only bound one IP or account, and API, database and worker share one box (ADR infrastructure). Options: a CDN/WAF such as Cloudflare, or nothing for now; a proxy must keep real client IPs for the per-IP limits. Settled by that choice, recorded in ADR infrastructure. Not something to test live.

## Mask notification-channel secrets in API responses?

2026-10-13 · adr

The notification-channel API returns each channel's `config` as stored (`toNotificationChannelResponse` in `apps/api/internal/handler/notification_channels.go`), including bearer-like secrets such as a Slack webhook URL. The webhook signing secret must stay visible so users can verify signatures. Settled by deciding whether to mask the bearer-like ones after creation, recorded in ADR alerting.

## Drop the legacy org alert columns

2026-10-13 · notification channels epic

`orgs.telegram_chat_id`, `orgs.alert_email` and `orgs.email_alerts_enabled` were kept after the move to notification channels (ADR alerting) as a rollback path, to be dropped by a later migration once the cutover had run cleanly in production. It has run since 2026-06-21. Settled by a migration that drops them, or by deciding to keep them.

## Confirm where the PRDs follow the code over the old epics

2026-10-13 · prd

Rewriting the epics as PRDs found three places where the code had moved away from the epic, and each PRD now describes the code. Settled when each is confirmed, or the code is changed back:

- SSL and domain monitors: the epics said the hostname can't be edited (delete and recreate); the API allows editing it (PRD ssl-monitoring, PRD domain-expiry-monitoring).
- Telegram webhook with a wrong secret: the epic said `401`; the handler answers `200` and ignores the update, so scanners and Telegram never retry (PRD telegram-alerts, PRD security-hardening).
- Uptime intervals: the epic's 10/30-minute MVP minimum was replaced by the plans' 5/1-minute minimums (PRD uptime-monitoring, ADR billing).

## Hand-built alert messages per monitor type

2026-10-14 · tech-debts.md

`checkOverdue`, `buildUptimeDownAlert` and `buildPortDownAlert` each build their own `AlertMessage` (Telegram, email subject and HTML, webhook, Slack, SMS: one `fmt.Sprintf` per field), unlike SSL and domain, which share `expiredMessages`/`expiringSoonMessages`. Their reason text really differs per type, so one shared template would lose the phrasing or take as many parameters as it saves. Left as is; settled by extracting a shared builder when a fourth near-identical case shows up, or by deciding never to.

## SMS delivery failures after Twilio accepts a send

2026-10-14 · twilio-setup.md

Checkmeup records a send as failed only when Twilio rejects the API call. A message Twilio accepts and a carrier later drops (an unreachable number, a STOP opt-out on a two-way sender) still reads as delivered, and the org's email fallback never fires. Catching it needs a status-callback endpoint (e.g. `POST /webhooks/twilio/sms-status`) verified by `X-Twilio-Signature`, set as the Messaging Service's status callback. Settled by building it or deciding the gap is acceptable.
