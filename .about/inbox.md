# Inbox

## Check Search Console after the changelog release

2026-10-13 · session review · kept 2026-10-14

Release notes `noindex`, `/changelog`, the trailing-slash 301 and the path-only canonical shipped on 2026-10-13. Google drops pages only after a recrawl, which can take weeks. The "Page with redirect" count should grow as trailing-slash variants become 301s. The guess that today's 4 redirect pages are the `http://` and `www.` variants was never checked. On 2026-10-17 the report (data as of 2026-10-12) still listed three release notes and a trailing-slash URL as indexed, and the sitemap as last read with 54 pages; the sitemap was resubmitted, and indexing requested for `/about` and `/blog`, both crawled but not indexed. Settled when Search Console shows the release notes excluded by `noindex`, no trailing-slash duplicates indexed, and the redirect pages accounted for.

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

## CLAUDE.md restructure not started

2026-10-15 · session review

The user asked to restructure `CLAUDE.md`: drop instructions that change nothing, and disclose detail only when it's needed. Three subagents should each go further than the last, ending in one PR, and with no `CODING_STANDARDS.md` (the user ruled that out). Work stopped before any pass. Its empty worktree and branch were removed on 2026-10-17. Settled when the PR is open, or the item is deleted if the restructure is dropped.

## Per-incident alert cap can't trigger for uptime, port and DNS

2026-10-17 · session review

`alertUptimeIncident`, `alertPortIncident` and `alertDNSIncident` check `max_alerts_per_incident` against the incident's `alert_count`, but each is called only once, right after the incident is created, when the count is still 0, so the cap never stops anything (ADR alerting says the cap is what bounds SMS cost). Harmless today, since these types send one down alert per incident anyway; `alerts_enabled` is now tested for all three. Settled by dropping the dead check from the three workers, or keeping it on purpose for a future re-alert while still down, written into ADR alerting.

## Landing hero's author line

2026-10-15 · session review

The redesigned hero (variant A) was meant to sit next to a first-person line from the founder. It didn't ship: the only draft of the line was written by Claude. The live 90-day strip of checkmeup.net, planned with it, has since shipped. Settled when the founder writes the line or drops it.

## SEO leftovers from the 2026-10-17 check

2026-10-17 · session review

The SEO check fixed long descriptions and `HEAD` 405s but left smaller findings: every post shares one OG image, so shared links all look alike; `http://www.checkmeup.net` takes two redirects instead of one (kamal-proxy's doing). Core Web Vitals were never checked (PageSpeed Insights or a browser). Settled when each is fixed or judged not worth it, and Core Web Vitals have been looked at once.

## Almost no organic search traffic

2026-10-17 · session review

Search Console showed 0 clicks and 127 impressions from 2026-07-15 to 2026-10-14, almost all for the brand name or typos of it, with nothing for topics like cron job monitoring. On 2026-10-12, 41 pages weren't indexed (26 discovered and 9 crawled but not indexed). GA4 counted 3 users in the 28 days to 2026-10-16, none from search. Technically the site is in order (the SEO check of 2026-10-17); a young domain usually needs outside signals, such as links from tool directories, Product Hunt, Show HN or the GitHub README. Settled by choosing which channels to try, or by Search Console showing clicks on non-brand queries.

## Free tools without sign-up

2026-10-17 · traffic review

Public, no-account checkers people search for: SSL certificate, domain expiry (RDAP/WHOIS), DNS lookup, TCP port, cron expression explainer. The backend already has `rdap`, `whois` and the worker's checks; each result offers to monitor it automatically. Also linkable assets for directories and Show HN (see "Almost no organic search traffic"). Needs per-IP rate limits and SSRF-safe targets. SSL and domain expiry checkers came first (PRD free-tools); DNS lookup, TCP port and the cron explainer remain. Settled when the rest are built or dropped.

## "Alternative to X" pages

2026-10-17 · traffic review

Only one combined comparison post exists (`checkmeup-vs-healthchecks-uptimerobot-cronitor`). Separate pages for "Healthchecks.io alternative", "UptimeRobot alternative" and "Cronitor alternative" catch people already choosing a tool. Settled when the pages ship, or the idea is dropped.

## Product pages per monitor type

2026-10-17 · traffic review

Cron, uptime, SSL, domain, port and DNS monitoring each have only a blog guide; there's no product page such as `/cron-monitoring` with the plan and a sign-up CTA. Settled when the pages ship or the guides are judged enough.

## Stack-specific cron guides

2026-10-17 · traffic review

Long-tail guides: monitoring a Kubernetes CronJob, a GitHub Actions schedule, Laravel's scheduler, systemd timers, Celery beat. Pairs with the CLI wrapper in `docs/proposals/bucket-list.md`. Settled when the stacks to cover are picked and written, or the idea is dropped.

## Launch and directory listings

2026-10-17 · traffic review

Channels only the founder can post to: Show HN, Product Hunt, Reddit (r/webdev, r/selfhosted, r/freelance), Indie Hackers, AlternativeTo, SaaSHub, awesome lists on GitHub; build-in-public posts such as "Three Days to MVP" cross-posted to dev.to with a canonical link back. The outside signals "Almost no organic search traffic" calls for. Tried so far: awesome-monitoring, [PR #121](https://github.com/Enapiuz/awesome-monitoring/pull/121) opened 2026-10-17 (the maintainer last merged on 2026-08-04, 26 PRs open). Settled when the channels are chosen and tried.

## Main-commit hook checks the wrong repo

2026-10-17 · session review

`.claude/hooks/block_main_commit.py` reads the branch with `git rev-parse` in the hook's own working directory, the checkmeup checkout, not the repo the command commits in. So a commit in another repo (hit with an awesome-monitoring fork in the scratchpad) is blocked while checkmeup sits on `main`, and `git -C <path>` slips past its regex altogether. Its regex also matches the words inside heredocs and quoted text, so a command that only writes about committing is blocked too. Settled when the hook resolves the target repo from the command's `cd` or `-C` and ignores quoted text, or the cases are judged rare enough to leave.

## SMS consent wording never reviewed

2026-10-18 · tidy

`docs/reference/twilio-setup.md` asked for the SMS opt-in checkbox's consent wording (PRD sms-alerts) to be reviewed against TCPA-style rules in the markets Checkmeup has users in "before launch"; SMS shipped on 2026-07-04 with no record of that review. A legal-text question, not a Twilio setting. Settled when the wording is reviewed (and changed if needed), or judged fine as is.
