# Inbox

## Check Search Console after the changelog release

2026-10-13 · session review · kept 2026-10-14

Release notes `noindex`, `/changelog`, the trailing-slash 301 and the path-only canonical shipped on 2026-10-13. Google drops pages only after a recrawl, which can take weeks. The "Page with redirect" count should grow as trailing-slash variants become 301s. The guess that today's 4 redirect pages are the `http://` and `www.` variants was never checked. On 2026-10-17 the report (data as of 2026-10-12) still listed three release notes and a trailing-slash URL as indexed, and the sitemap as last read with 54 pages; the sitemap was resubmitted, and indexing requested for `/about` and `/blog`, both crawled but not indexed. On 2026-10-18 Google had read the sitemap that day (33 pages, status Success), but the Pages report was still dated 2026-10-12 (21 indexed, 41 not). On 2026-10-20 URL Inspection showed `/about`, `/blog` and all eleven new marketing pages indexed, while the Pages report was still dated 2026-10-12. Settled when Search Console shows the release notes excluded by `noindex`, no trailing-slash duplicates indexed, and the redirect pages accounted for.

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

## Shipped PRD criteria the code doesn't meet

2026-10-14 · session review · updated 2026-10-18

A test pass on 2026-10-18 linked tests to every other shipped criterion. These are left, because the code doesn't do what they say or nothing on the web side can check them:

- annual-billing: the renewal date a year after purchase comes from Paddle's `next_billed_at` (`resolveRenewsAt`); only a Go test can check it.
- dns-monitoring: the detail page's change log (from what, to what) isn't built; there is only the per-lookup check log.
- telegram-alerts: a muted monitor isn't marked as muted in the monitor list.
- notification-channels: a new monitor's form starts with no channel selected; the API attaches every enabled channel when none is sent, so the outcome matches but the form says alerts fall back to email.
- webhook-alerts: Settings explains signature checking in one sentence, with no code snippet.
- The body cap and the slow webhook have items of their own below; the read-only key scope too.

Settled when each is built, tested, or reworded in its PRD.

## Architecture deepening candidates

2026-10-14 · session review · kept 2026-10-18

An architecture review found three places where each monitor type repeats logic that should exist once:

- Worker incident handling (strong): uptime, port and DNS each count failures, open incidents, apply the alert cap, alert and resolve on their own. One module would own those steps, each type supplying its check, queries and message wording, with a small adapter per type for sqlc's per-table structs. Cron, SSL and domain stay out.
- Plan checks in the monitor handlers (strong): the create-limit check and its 402 are copied across six handlers, the interval clamp three times, and not-found handled three ways. One plan check would write the 402 itself; pause, resume and delete would share one not-found path.
- SSL and domain as one "Expiry monitor" (worth exploring): their workers, handlers and edit views differ only slightly; merging them would need a glossary term.

The review recommended the worker first, as alerting is the core of the product, or the plan checks as a cheaper first step. Settled by picking one and agreeing a plan, or by dropping the candidates.

## Landing hero's author line

2026-10-15 · session review · kept 2026-10-18

The redesigned hero (variant A) was meant to sit next to a first-person line from the founder. It didn't ship: the only draft of the line was written by Claude. The live 90-day strip of checkmeup.net, planned with it, has since shipped. Settled when the founder writes the line or drops it.

## SEO leftovers from the 2026-10-17 check

2026-10-17 · session review · kept 2026-10-18

The SEO check fixed long descriptions and `HEAD` 405s but left smaller findings: every post shares one OG image, so shared links all look alike; `http://www.checkmeup.net` takes two redirects instead of one (kamal-proxy's doing). Core Web Vitals were never checked (PageSpeed Insights or a browser). Settled when each is fixed or judged not worth it, and Core Web Vitals have been looked at once.

## Almost no organic search traffic

2026-10-17 · session review · kept 2026-10-18

Search Console showed 0 clicks and 127 impressions from 2026-07-15 to 2026-10-14, almost all for the brand name or typos of it, with nothing for topics like cron job monitoring. On 2026-10-12, 41 pages weren't indexed (26 discovered and 9 crawled but not indexed). GA4 counted 3 users in the 28 days to 2026-10-16, none from search. Technically the site is in order (the SEO check of 2026-10-17); a young domain usually needs outside signals, such as links from tool directories, Product Hunt, Show HN or the GitHub README. The free SSL and domain checkers shipped 2026-10-17 (PRD free-tools); their success signal is due 2026-12-15, and what they bring should inform the next channels. On 2026-10-20 the 28 days to 2026-10-18 showed 0 clicks and 76 impressions (average position 32.8), now mostly non-brand: "port monitoring" 8, "network port monitoring" 7, "cron job monitoring" 4; impressions rose from 1–5 a day to 9–18 a day from 2026-10-16, led by the port, comparison and cron guides on the blog. Settled by choosing which channels to try, or by Search Console showing clicks on non-brand queries.

## Free tools without sign-up

2026-10-17 · traffic review · kept 2026-10-18

Public, no-account checkers people search for: SSL certificate, domain expiry (RDAP/WHOIS), DNS lookup, TCP port, cron expression explainer. The backend already has `rdap`, `whois` and the worker's checks; each result offers to monitor it automatically. Also linkable assets for directories and Show HN (see "Almost no organic search traffic"). Needs per-IP rate limits and SSRF-safe targets. SSL and domain expiry checkers came first (PRD free-tools); DNS lookup, TCP port and the cron explainer remain, waiting on the first two tools' 2026-12-15 success signal. Settled when the rest are built or dropped.

## Stack-specific cron guides

2026-10-17 · traffic review · kept 2026-10-18

Long-tail guides: monitoring a Kubernetes CronJob, a GitHub Actions schedule, Laravel's scheduler, systemd timers, Celery beat. Pairs with the CLI wrapper in `docs/proposals/bucket-list.md`. Settled when the stacks to cover are picked and written, or the idea is dropped.

## Launch and directory listings

2026-10-17 · traffic review · kept 2026-10-18

Channels only the founder can post to: Show HN, Product Hunt, Reddit (r/webdev, r/selfhosted, r/freelance), Indie Hackers, AlternativeTo, SaaSHub, awesome lists on GitHub; build-in-public posts such as "Three Days to MVP" cross-posted to dev.to with a canonical link back. The outside signals "Almost no organic search traffic" calls for. Tried so far: awesome-monitoring, [PR #121](https://github.com/Enapiuz/awesome-monitoring/pull/121) opened 2026-10-17 (the maintainer last merged on 2026-08-04, 26 PRs open); still open and unreviewed on 2026-10-18. Settled when the channels are chosen and tried.

## SMS consent wording never reviewed

2026-10-18 · tidy · kept 2026-10-18

`docs/reference/twilio-setup.md` asked for the SMS opt-in checkbox's consent wording (PRD sms-alerts) to be reviewed against TCPA-style rules in the markets Checkmeup has users in "before launch"; SMS shipped on 2026-07-04 with no record of that review. A legal-text question, not a Twilio setting. Settled when the wording is reviewed (and changed if needed), or judged fine as is.

## Body cap answers 400, not 413

2026-10-18 · PRD test pass

PRD security-hardening says a request body over 64 KB is refused with `413` before any handler reads it. `http.MaxBytesReader` only fails once a handler reads the body, and the handler then answers `400 bad_request`: a 70 KB sign-up got that. The cap works, the status and timing differ. Settled by answering `413` (checking `Content-Length` up front, or mapping `*http.MaxBytesError` to `413`) with a test, or by rewording the criterion to what the code does.

## A slow webhook can delay the check loop

2026-10-18 · PRD test pass

PRD webhook-alerts says a slow endpoint is cut off after 10 seconds and never delays the check loop. The 10-second cut-off is real (`deliver.Timeout`), but `DispatchAlert` runs inside each check's goroutine, and `checkUptimeMonitors` waits for all of them (`wg.Wait`) before the tick moves on, so one slow endpoint can hold a 30-second tick for up to 10 seconds. The timeout also can't be shortened in a test without a real 10-second wait. Settled by dispatching alerts outside the check's wait group (and making the timeout injectable) with a test, or by rewording the criterion.
