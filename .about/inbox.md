# Inbox

## Move the project's docs into .about/

2026-10-13 · session review

ADRs still live in `docs/decisions/` (numbered, scaffolded by `new-adr`), stories in `docs/stories/` (`new-story`) and open questions in `docs/decisions/backlog.md`. baloo reads `.about/adr/`, `.about/prd/` and `.about/inbox.md`. The agreed PR moves the ADRs to `.about/adr/`, the stories to PRDs and `backlog.md` into this Inbox, deletes the `new-adr` and `new-story` skills, and fixes every link. Settled when that PR merges.

## Check Search Console once the changelog branch ships

2026-10-13 · session review

`feat/changelog-noindex-release-notes` (release notes `noindex`, `/changelog`, trailing-slash 301, path-only canonical) isn't merged yet. Google drops pages only after a recrawl, which can take weeks. The "Page with redirect" count should grow as trailing-slash variants become 301s. The guess that today's 4 redirect pages are the `http://` and `www.` variants was never checked. Settled when Search Console shows the release notes excluded by `noindex`, no trailing-slash duplicates indexed, and the redirect pages accounted for.

## WhatsApp provider

2026-10-13 · docs/decisions/backlog.md

PRD whatsapp-alerts needs a provider: Meta's Cloud API directly, or a BSP (Twilio, MessageBird). It decides setup effort, per-message cost and who handles template approval. The down and recovery templates must also be drafted and approved by Meta before the alerts can ship; plan for that lead time. Settled by picking the provider, recorded in ADR alerting.

## Signal alerts: build them at all?

2026-10-13 · docs/decisions/backlog.md

Signal has no bot or business API; the only path is self-hosting `signal-cli` with a phone number registered as a bot, which Signal's anti-abuse systems may ban, and which adds an always-on process (ADR workers avoids those). Settled by a go/no-go, ideally from real user demand, recorded in ADR alerting.

## Microsoft Teams webhook mechanism

2026-10-13 · docs/decisions/backlog.md

Microsoft is retiring the Office 365 Connector "Incoming Webhook" for Power Automate workflow webhooks (Adaptive Card payload). Settled by re-checking the retirement status and the recommended mechanism right before PRD ms-teams-alerts is built.

## Multi-region check infrastructure

2026-10-13 · docs/decisions/backlog.md

PRD multi-region-checks needs compute in more than one region, against one VPS and no broker (ADR workers, ADR infrastructure). Options: extra regional VPSes reporting to the one PostgreSQL, or a check-execution service; and which regions first. Settled by that choice, recorded in ADR workers.

## Read-only scope for API keys

2026-10-13 · docs/decisions/backlog.md

`api_keys` has no `scope` column and `RequireAPIKey` checks no read-only/read-write scope (ADR auth). Harmless while every `/api/v1/public` route is a `GET`. Settled when the `scope` column and a `403` for read-only keys on non-`GET` requests land, before the first write endpoint.

## CDN or WAF in front of the VPS

2026-10-13 · docs/decisions/backlog.md

Nothing between the internet and kamal-proxy absorbs a volumetric attack; per-route rate limits (ADR limits) only bound one IP or account, and API, database and worker share one box (ADR infrastructure). Options: a CDN/WAF such as Cloudflare, or nothing for now; a proxy must keep real client IPs for the per-IP limits. Settled by that choice, recorded in ADR infrastructure. Not something to test live.

## Mask notification-channel secrets in API responses?

2026-10-13 · adr

The notification-channel API returns each channel's `config` as stored (`toNotificationChannelResponse` in `apps/api/internal/handler/notification_channels.go`), including bearer-like secrets such as a Slack webhook URL. The webhook signing secret must stay visible so users can verify signatures. Settled by deciding whether to mask the bearer-like ones after creation, recorded in ADR alerting.
