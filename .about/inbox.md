# Inbox

## Check Search Console after the changelog release

2026-10-13 · session review · kept 2026-10-31

Release notes `noindex`, `/changelog`, the trailing-slash 301 and the path-only canonical shipped on 2026-10-13. Google drops pages only after a recrawl, which can take weeks. The "Page with redirect" count should grow as trailing-slash variants become 301s. The guess that today's 4 redirect pages are the `http://` and `www.` variants was never checked. On 2026-10-17 the report (data as of 2026-10-12) still listed three release notes and a trailing-slash URL as indexed, and the sitemap as last read with 54 pages; the sitemap was resubmitted, and indexing requested for `/about` and `/blog`, both crawled but not indexed. On 2026-10-18 Google had read the sitemap that day (33 pages, status Success), but the Pages report was still dated 2026-10-12 (21 indexed, 41 not). On 2026-10-20 URL Inspection showed `/about`, `/blog` and all eleven new marketing pages indexed, while the Pages report was still dated 2026-10-12. On 2026-10-31 the Pages report (data as of 2026-10-23) showed 40 indexed and 18 not. The 4 redirect pages are accounted for: `http://checkmeup.net/`, `http://www.checkmeup.net/`, `https://www.checkmeup.net/` and an old `/blog/v1-28-release-notes`. Excluded by `noindex`: 3 release notes recrawled since the change and `/blog/checkmeup-vs-competitors`. Still indexed: 10 release notes and one trailing-slash URL (`/blog/v1-28-release-notes/`), all last crawled between 2026-08-24 and 2026-10-03, before the change; production serves them `noindex, follow` and the slash variant a 301, so only a recrawl is missing. Requesting indexing doesn't speed it up: URL Inspection's live test sees the `noindex` and rejects the request (tried on `/blog/v1-15-release-notes` on 2026-10-31). Settled when Search Console shows the remaining release notes excluded by `noindex` and the trailing-slash URL gone from the index.

## WhatsApp provider

2026-10-13 · decision backlog · kept 2026-10-31

PRD whatsapp-alerts needs a provider: Meta's Cloud API directly, or a BSP (Twilio, MessageBird). It decides setup effort, per-message cost and who handles template approval. The down and recovery templates must also be drafted and approved by Meta before the alerts can ship; plan for that lead time. No user has asked for it as of 2026-10-29 (about 13 people a day on the site). Settled by picking the provider, recorded in ADR alerting.

## Multi-region check infrastructure

2026-10-13 · decision backlog · kept 2026-10-31

PRD multi-region-checks needs compute in more than one region, against one VPS and no broker (ADR workers, ADR infrastructure). Options: extra regional VPSes reporting to the one PostgreSQL, or a check-execution service; and which regions first. Revisit when paying users report false alarms from the single region (none as of 2026-10-29). Settled by that choice, recorded in ADR workers.

## Shipped PRD criteria the code doesn't meet

2026-10-14 · session review · kept 2026-10-31

A test pass on 2026-10-18 linked tests to every other shipped criterion. These are left, because the code doesn't do what they say or nothing on the web side can check them:

- annual-billing: the renewal date a year after purchase comes from Paddle's `next_billed_at` (`resolveRenewsAt`); only a Go test can check it.
- dns-monitoring: the detail page's change log (from what, to what) isn't built; there is only the per-lookup check log.
- telegram-alerts: a muted monitor isn't marked as muted in the monitor list.
- notification-channels: a new monitor's form starts with no channel selected; the API attaches every enabled channel when none is sent, so the outcome matches but the form says alerts fall back to email.
- webhook-alerts: Settings explains signature checking in one sentence, with no code snippet.
- The slow webhook has an item of its own below.

Settled when each is built, tested, or reworded in its PRD.

## Architecture deepening candidates

2026-10-14 · session review · kept 2026-10-31

An architecture review found three places where each monitor type repeats logic that should exist once:

- Worker incident handling (strong): uptime, port and DNS each count failures, open incidents, apply the alert cap, alert and resolve on their own. One module would own those steps, each type supplying its check, queries and message wording, with a small adapter per type for sqlc's per-table structs. Cron, SSL and domain stay out.
- Plan checks in the monitor handlers (strong): the create-limit check and its 402 are copied across six handlers, the interval clamp three times, and not-found handled three ways. One plan check would write the 402 itself; pause, resume and delete would share one not-found path.
- SSL and domain as one "Expiry monitor" (worth exploring): their workers, handlers and edit views differ only slightly; merging them would need a glossary term.

The review recommended the worker first, as alerting is the core of the product, or the plan checks as a cheaper first step. None was started as of 2026-10-29; the worker candidate is to go with "A slow webhook can delay the check loop", in a session of its own (`/baloo:architecture`). Settled by picking one and agreeing a plan, or by dropping the candidates.

## Almost no organic search traffic

2026-10-17 · session review · kept 2026-10-31

Search Console showed 0 clicks and 127 impressions from 2026-07-15 to 2026-10-14, almost all for the brand name or typos of it, with nothing for topics like cron job monitoring. On 2026-10-12, 41 pages weren't indexed (26 discovered and 9 crawled but not indexed). GA4 counted 3 users in the 28 days to 2026-10-16, none from search. Technically the site is in order (the SEO check of 2026-10-17); a young domain usually needs outside signals, such as links from tool directories, Product Hunt, Show HN or the GitHub README. The free SSL and domain checkers shipped 2026-10-17 (PRD free-tools); their success signal is due 2026-12-15, and what they bring should inform the next channels. On 2026-10-20 the 28 days to 2026-10-18 showed 0 clicks and 76 impressions (average position 32.8), now mostly non-brand: "port monitoring" 8, "network port monitoring" 7, "cron job monitoring" 4; impressions rose from 1–5 a day to 9–18 a day from 2026-10-16, led by the port, comparison and cron guides on the blog. On 2026-10-24 impressions had jumped to 60–67 a day from 2026-10-19 (235 in the week to 2026-10-21, 1 brand click), 169 of them for `/tools/domain-expiry-checker` on "domain expiry check(er)" and "check domain expiration" queries, but at an average position of 70–85, too deep for clicks. On 2026-10-25 the 28 days to 2026-10-23 showed 1 click (the home page, likely brand) and 397 impressions at an average position of 68, holding at 60–76 a day; `/tools/domain-expiry-checker` had 285 of them at position 79, and the comparison post was the only page near the first results page (position 8.6, 16 impressions). GA4 counted 18 sessions from 2026-09-27 to 2026-10-24: 13 direct, 4 referral, 1 organic. On 2026-10-29 `mise run traffic` showed 13 visitors who loaded the app's JS that day; external referrers were Google 1, SaaSHub 1, and a few directory or spam sites (five.co.in, allinone.co.in, 4anything.net). The proxy log only went back to 2026-10-28, most likely reset when a deploy recreated kamal-proxy. On 2026-10-31 it held only that day again: 17 visitors, 1 of them loading the app's JS, 28 pageviews (23 on `/`) against 91 bot pageviews; the only external referrer was five.co.in (2). Settled by choosing which channels to try, or by Search Console showing clicks on non-brand queries.

## Free tools without sign-up

2026-10-17 · traffic review · kept 2026-10-29

Public, no-account checkers people search for: SSL certificate, domain expiry (RDAP/WHOIS), DNS lookup, TCP port, cron expression explainer. The backend already has `rdap`, `whois` and the worker's checks; each result offers to monitor it automatically. Also linkable assets for directories and Show HN (see "Almost no organic search traffic"). Needs per-IP rate limits and SSRF-safe targets. SSL and domain expiry checkers came first (PRD free-tools), and the DNS lookup followed on 2026-10-31 (`/tools/dns-lookup`); TCP port and the cron explainer remain, waiting on the tools' 2026-12-15 success signal. Settled when the rest are built or dropped.

## Stack-specific cron guides

2026-10-17 · traffic review · kept 2026-10-31

Long-tail guides: monitoring a Kubernetes CronJob, a GitHub Actions schedule, Laravel's scheduler, systemd timers, Celery beat. Pairs with the CLI wrapper in `docs/proposals/bucket-list.md`. Settled when the stacks to cover are picked and written, or the idea is dropped. Revisit when Search Console shows which cron queries bring impressions.

## Launch and directory listings

2026-10-17 · traffic review · kept 2026-10-29

Channels only the founder can post to: Show HN, Product Hunt, Reddit (r/webdev, r/selfhosted, r/freelance), Indie Hackers, AlternativeTo, SaaSHub, awesome lists on GitHub; build-in-public posts such as "Three Days to MVP" cross-posted to dev.to with a canonical link back. The outside signals "Almost no organic search traffic" calls for. Tried so far: awesome-monitoring, [PR #121](https://github.com/Enapiuz/awesome-monitoring/pull/121) opened 2026-10-17 (the maintainer last merged on 2026-08-04, 26 PRs open); still open and unreviewed on 2026-10-25. AlternativeTo: listed since 2026-08-07 as an alternative to 10 products, but with 0 likes it ranks low (not among the first 12 of 47 Healthchecks.io alternatives on 2026-10-25), and the listing predates the free checkers. SaaSHub: product page filled in on 2026-10-25 (description, pricing, competitors, features, Q&A, screenshots); its "Product of the Day" round needs weeks of daily voting to build a nominator rating, judged not worth it. Show HN: [item 50235049](https://news.ycombinator.com/item?id=50235049) posted 2026-10-25 (a Sunday) was `[flagged]` within a minute, most likely a spam filter on the 1-karma account; the next step is asking `hn@ycombinator.com` to unflag it, else repost after a few weeks of commenting, on a Tuesday–Thursday around 8–10am US Eastern, without changing the text. HN and Reddit gate posting on karma, which would mean commenting from the founder's account; generated comments are against HN's rules, so not done by Claude. Lists without a karma gate: awesome-status-pages, [PR #252](https://github.com/ivbeg/awesome-status-pages/pull/252) opened 2026-10-27 and closed by the maintainer the same day: its description carried a "Generated with Claude Code" line, and the list bans AI-generated PRs and won't consider the service, so that list is likely closed to Checkmeup; awesome-devops, [PR #636](https://github.com/wmariuss/awesome-devops/pull/636) opened and closed 2026-10-27, because its bot requires a repository at least 6 months old and `checkmeup/checkmeup` was created 2026-06-12: reopen it after 2026-12-12 (`gh pr reopen 636 -R wmariuss/awesome-devops`, then an empty push to rerun the check). dev.to (no karma gate, posts carry a canonical link back to the blog): [RDAP vs WHOIS](https://dev.to/andrewmolyuk/rdap-vs-whois-why-some-domains-hide-their-expiry-2mci) published 2026-10-27, and [the SSL story](https://dev.to/andrewmolyuk/the-3-am-ssl-cert-that-cost-us-a-client-5f71) on 2026-10-31, ending with a link to the free SSL checker; the cron guide and "Three Days to MVP" are next, one every few days. HN: `hn@ycombinator.com` asked on 2026-10-27 to unflag the Show HN; on 2026-10-29 the item was still `dead`, and awesome-monitoring PR #121 still open with no review. On 2026-10-31 the Show HN was still `dead`; awesome-monitoring [PR #121](https://github.com/Enapiuz/awesome-monitoring/pull/121) was merged on 2026-10-30. Settled when the channels are chosen and tried.

## A slow webhook can delay the check loop

2026-10-18 · PRD test pass · kept 2026-10-31

PRD webhook-alerts says a slow endpoint is cut off after 10 seconds and never delays the check loop. The 10-second cut-off is real (`deliver.Timeout`), but `DispatchAlert` runs inside each check's goroutine, and `checkUptimeMonitors` waits for all of them (`wg.Wait`) before the tick moves on, so one slow endpoint can hold a 30-second tick for up to 10 seconds. The timeout also can't be shortened in a test without a real 10-second wait. Settled by dispatching alerts outside the check's wait group (and making the timeout injectable) with a test, or by rewording the criterion. `DispatchAlert`'s result decides whether the alert counts toward `alert_count`, so dispatching outside the wait group has to keep that count right; it touches the same steps as the worker incident-handling candidate under "Architecture deepening candidates", so do them together.
