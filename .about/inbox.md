# Inbox

## Check Search Console after the changelog release

2026-10-13 · session review · kept 2026-10-14

Release notes `noindex`, `/changelog`, the trailing-slash 301 and the path-only canonical shipped on 2026-10-13. Google drops pages only after a recrawl, which can take weeks. The "Page with redirect" count should grow as trailing-slash variants become 301s. The guess that today's 4 redirect pages are the `http://` and `www.` variants was never checked. Settled when Search Console shows the release notes excluded by `noindex`, no trailing-slash duplicates indexed, and the redirect pages accounted for.

## WhatsApp provider

2026-10-13 · decision backlog · kept 2026-10-14

PRD whatsapp-alerts needs a provider: Meta's Cloud API directly, or a BSP (Twilio, MessageBird). It decides setup effort, per-message cost and who handles template approval. The down and recovery templates must also be drafted and approved by Meta before the alerts can ship; plan for that lead time. Settled by picking the provider, recorded in ADR alerting.

## Multi-region check infrastructure

2026-10-13 · decision backlog · kept 2026-10-14

PRD multi-region-checks needs compute in more than one region, against one VPS and no broker (ADR workers, ADR infrastructure). Options: extra regional VPSes reporting to the one PostgreSQL, or a check-execution service; and which regions first. Settled by that choice, recorded in ADR workers.

## Read-only scope for API keys

2026-10-13 · decision backlog · kept 2026-10-14

`api_keys` has no `scope` column and `RequireAPIKey` checks no read-only/read-write scope (ADR auth). Harmless while every `/api/v1/public` route is a `GET` (six status reads as of 2026-10-14). Settled when the `scope` column and a `403` for read-only keys on non-`GET` requests land, before the first write endpoint.

## CDN or WAF in front of the VPS

2026-10-13 · decision backlog · kept 2026-10-14

Nothing between the internet and kamal-proxy absorbs a volumetric attack; per-route rate limits (ADR limits) only bound one IP or account, and API, database and worker share one box (ADR infrastructure). Options: a CDN/WAF such as Cloudflare, or nothing for now; a proxy must keep real client IPs for the per-IP limits. Settled by that choice, recorded in ADR infrastructure. Not something to test live.

## Drop the legacy org alert columns

2026-10-13 · notification channels epic

`orgs.telegram_chat_id`, `orgs.alert_email` and `orgs.email_alerts_enabled` were kept after the move to notification channels (ADR alerting) as a rollback path, to be dropped by a later migration once the cutover had run cleanly in production. It has run since 2026-06-21. Since 2026-10-14 no code reads or writes them: `CreateOrg` lists its columns instead of `RETURNING *`, and the unused `UpdateOrg*` queries are gone. Settled by the migration that drops them, shipped in a release after that one is deployed, since migrations run at container start while the previous container still serves (`cmd/api/main.go`).
