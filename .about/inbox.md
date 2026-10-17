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

## Shipped PRD criteria with no test

2026-10-14 · session review

Writing the PRDs found about 20 acceptance criteria of shipped features with no `→ test`, among them the rate limits on auth endpoints and pings, the 64 KB request body cap, the daily SSL and domain check frequency, the channel-creation cap and the cross-channel alert cap, the prerendered HTML, and the badges' `Cache-Control`. Nothing would notice if one of them broke. Settled when each has a test linked from its PRD, or is reworded or dropped as not worth testing.

## Architecture deepening candidates

2026-10-14 · session review

An architecture review found three places where each monitor type repeats logic that should exist once:

- Worker incident handling (strong): uptime, port and DNS each count failures, open incidents, apply the alert cap, alert and resolve on their own. One module would own those steps, each type supplying its check, queries and message wording, with a small adapter per type for sqlc's per-table structs. Cron, SSL and domain stay out.
- Plan checks in the monitor handlers (strong): the create-limit check and its 402 are copied across six handlers, the interval clamp three times, and not-found handled three ways. One plan check would write the 402 itself; pause, resume and delete would share one not-found path.
- SSL and domain as one "Expiry monitor" (worth exploring): their workers, handlers and edit views differ only slightly; merging them would need a glossary term.

The review recommended the worker first, as alerting is the core of the product, or the plan checks as a cheaper first step. Settled by picking one and agreeing a plan, or by dropping the candidates.

## Codacy API token rejected

2026-10-15 · session review

`CODACY_API_TOKEN` in `apps/api/.env` gets "Bad credentials" from the Codacy API, so the codacy-triage skill can't list issues; when a PR's Codacy check fails, the GitHub check run shows only the count, not the issues. Settled when a new account-level token is in `apps/api/.env` (and the CI secret, if it's the same one) and the skill's fetch returns issues.

## Small checkmarks in the Pricing comparison table

2026-10-15 · session review

The checkmarks in the plan comparison table on `/pricing` are too small to scan after the redesign. Settled when they're resized to read at a glance in both themes.

## Blog isn't highlighted in the nav on a post

2026-10-15 · session review

The marketing nav highlights the current page, but on a blog post (`/blog/<slug>`) "Blog" isn't highlighted. Settled when the Blog link is active on every `/blog/…` route, with a test.

## Literal asterisks in the DNS guide post

2026-10-15 · session review

`apps/web/src/blog/posts/dns-record-monitoring-guide.ts` renders literal asterisks around a word ("\*something\*"): the post's content blocks don't render Markdown emphasis. Settled when the text drops the asterisks or uses the content format's own emphasis.

## Unused logo-grey.svg

2026-10-15 · session review

`apps/web/src/assets/logo-grey.svg` is referenced nowhere since the footer switched to the regular logo files. Claude left it because the logo assets aren't to be touched without the owner. Settled by deleting it or naming where it's still needed.

## One failing SSL or domain monitor turns the whole summary strip red

2026-10-15 · session review

SSL and domain monitors store no daily history, so a status page colors all 90 of their days by their current status. The combined strip in `/status/<slug>/summary.json` takes each day's worst tone, so one such monitor in error turns every day red, including in the landing hero's live strip of checkmeup.net. Settled by deciding how those monitors count in the combined strip (today only, or left out of past days) and a test for it.

## OG image can't be regenerated from the repo

2026-10-15 · session review

`apps/web/public/img/checkmeup-og.png` (2400×1260, redrawn in PR 156 to match the landing hero: headline, crontab line, missed-ping alert, 90-day strip) was made by screenshotting a 1200×630 HTML page in the light theme's tokens with Chrome's `--headless --screenshot` at 2×. That HTML page isn't in the repo, so the next redraw starts from scratch, and the image can drift from the tokens unnoticed. Settled by committing the source page and the render command (e.g. under `apps/web/scripts/`), or deciding the PNG alone is enough.

## Agreed footer rework was lost

2026-10-15 · session review

On `feat/footer-labels` the user agreed a reworked marketing footer in `apps/web/src/layouts/LandingLayout.vue`: About moved into the first column with Docs and Pricing, Sign in and Sign up alone in the middle column, no column headings, Status (`/status/checkmeup-net`, a plain `<a>` since the backend serves it) and GitHub at the right of the bottom row, and the inner line above the copyright row kept. It was never committed and is gone from the working tree, and `main` still has the old footer. Settled by redoing it as a PR, or dropping it if the discard was deliberate.

## CLAUDE.md restructure not started

2026-10-15 · session review

The user asked to restructure `CLAUDE.md`: drop instructions that change nothing, and disclose detail only when it's needed. Three subagents should each go further than the last, ending in one PR, and with no `CODING_STANDARDS.md` (the user ruled that out). Work stopped before any pass. A worktree at `../checkmeup-claude-md` on branch `docs/claude-md-restructure` still exists with no changes. Settled when the PR is open, or the worktree and branch are removed if it's dropped.

## Alert cap and alerts-enabled untested for port and DNS

2026-10-14 · session review

Uptime, port and DNS monitors each have their own copy of the down/recover/incident/alert steps (`worker_uptime.go`, `worker_port.go`, `worker_dns.go`). `TestCheckUptimeMonitors` now checks that `alerts_enabled = false` sends nothing and that `max_alerts_per_incident` stops further alerts, but `TestCheckPortMonitors` and `TestCheckDNSMonitors` still don't (ADR alerting). A copy can drift without a test failing. Settled when port and DNS have those tests, or when the steps move into one tested module (Inbox: Architecture deepening candidates).

## Billing usage bars turn red from 90%

2026-10-15 · session review

On Billing, a usage bar at 90% or more of its plan limit (say 9 of 10 monitors on Hobby) is filled with the down-status red (`BillingView.vue`, `>= 90 ? 'var(--status-down)'`), which reads as an error rather than "nearly full" or "plan full". Settled by deciding what a nearly full and a full plan should look like (red, amber, or neutral with an upgrade hint) and styling the bars to match.

## Landing hero's author line

2026-10-15 · session review

The redesigned hero (variant A) was meant to sit next to a first-person line from the founder. It didn't ship: the only draft of the line was written by Claude. The live 90-day strip of checkmeup.net, planned with it, has since shipped. Settled when the founder writes the line or drops it.
