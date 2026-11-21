# Inbox

## WhatsApp provider

2026-10-13 · decision backlog · kept 2026-11-13

PRD whatsapp-alerts needs a provider: Meta's Cloud API directly, or a BSP (Twilio, MessageBird). It decides setup effort, per-message cost and who handles template approval. The down and recovery templates must also be drafted and approved by Meta before the alerts can ship; plan for that lead time. No user has asked for it as of 2026-11-13, and the site still gets between 1 and 23 real visitors a day. Settled by picking the provider, recorded in ADR alerting.

## Multi-region check infrastructure

2026-10-13 · decision backlog · kept 2026-11-13

PRD multi-region-checks needs compute in more than one region, against one VPS and no broker (ADR workers, ADR infrastructure). Options: extra regional VPSes reporting to the one PostgreSQL, or a check-execution service; and which regions first. Revisit when paying users report false alarms from the single region (none as of 2026-11-13; the 30-second recheck that shipped on 2026-10-28 already takes out the one-off failures a second region would have caught). Settled by that choice, recorded in ADR workers.

## Shipped PRD criteria the code doesn't meet

2026-10-14 · session review · kept 2026-11-13

A test pass on 2026-10-18 linked tests to every other shipped criterion. These are left, because the code doesn't do what they say or nothing on the web side can check them:

- dns-monitoring: the detail page's change log (from what, to what) isn't built; there is only the per-lookup check log.

The annual-billing one is settled: `TestResolveRenewsAt` covers `resolveRenewsAt` since 2026-11-13, and PRD annual-billing links the criterion to it. Settled when each is built, tested, or reworded in its PRD.

## Architecture deepening candidates

2026-10-14 · session review · kept 2026-11-13

An architecture review found three places where each monitor type repeats logic that should exist once:

- Worker incident handling (strong): uptime, port and DNS each count failures, open incidents, apply the alert cap, alert and resolve on their own. One module would own those steps, each type supplying its check, queries and message wording, with a small adapter per type for sqlc's per-table structs. Cron, SSL and domain stay out.
- Plan checks in the monitor handlers (strong): the create-limit check and its 402 are copied across six handlers, the interval clamp three times, and not-found handled three ways. One plan check would write the 402 itself; pause, resume and delete would share one not-found path.
- SSL and domain as one "Expiry monitor" (worth exploring): their workers, handlers and edit views differ only slightly; merging them would need a glossary term.

The review recommended the worker first, as alerting is the core of the product, or the plan checks as a cheaper first step. None was started as of 2026-11-13; the worker candidate wants a session of its own (`/baloo:architecture`), and what the product waits on now is traffic rather than code. Alerts have been sent outside the check since 2026-11-05 (`worker.GoAlert`, ADR alerting), which the module would take over. Settled by picking one and agreeing a plan, or by dropping the candidates.

## Free tools without sign-up

2026-10-17 · traffic review · kept 2026-11-13

Public, no-account checkers people search for: SSL certificate, domain expiry (RDAP/WHOIS), DNS lookup, TCP port, cron expression explainer. The backend already has `rdap`, `whois` and the worker's checks; each result offers to monitor it automatically. Also linkable assets for directories and Show HN (see "Launch and directory listings"). Needs per-IP rate limits and SSRF-safe targets. SSL and domain expiry checkers came first (PRD free-tools), the DNS lookup followed on 2026-10-31 (`/tools/dns-lookup`), and the cron explainer on 2026-11-21 (`/tools/cron-expression`), the one that leads to cron monitoring, the product's core. Only the TCP port checker remains, a non-goal of PRD free-tools for now, waiting on the tools' 2026-12-15 success signal. Settled when it is built or dropped.

## Stack-specific cron guides

2026-10-17 · traffic review · kept 2026-11-13

Long-tail guides: monitoring a Kubernetes CronJob, a GitHub Actions schedule, Laravel's scheduler, systemd timers, Celery beat. Pairs with the CLI wrapper in `docs/proposals/bucket-list.md`. The GitHub Actions one shipped on 2026-11-13 (`/blog/github-actions-cron-monitoring-guide`), picked because the head term "cron job monitoring" sits at position 73.6 while the tail is open, and because GitHub disabling a schedule after 60 days without commits is a failure the product's own dead man's switch catches. The other four wait on what that one does: whether it takes impressions, and on which queries. Settled when the stacks to cover are picked and written, or the idea is dropped.

## Launch and directory listings

2026-10-17 · traffic review · kept 2026-11-13

Channels only the founder can post to, for the outside signals a young domain needs to rank: almost no clicks come from search yet (Search Console and `mise run traffic` hold the numbers). Done: awesome-monitoring listed ([PR #121](https://github.com/Enapiuz/awesome-monitoring/pull/121), merged 2026-10-30); AlternativeTo and SaaSHub pages filled in, SaaSHub's "Product of the Day" judged not worth the weeks of voting; awesome-status-pages is closed to Checkmeup (CLAUDE.md). Left to do:

- awesome-devops: reopen [PR #636](https://github.com/wmariuss/awesome-devops/pull/636) after 2026-12-12, when `checkmeup/checkmeup` turns 6 months old (`gh pr reopen 636 -R wmariuss/awesome-devops`, then an empty push to rerun the check).
- Show HN: [item 50235049](https://news.ycombinator.com/item?id=50235049) stays `dead` after asking `hn@ycombinator.com` to unflag it; repost after a few weeks of commenting from the founder's account, on a Tuesday–Thursday around 8–10am US Eastern, without changing the text. HN and Reddit gate posting on karma, and generated comments are against HN's rules, so Claude doesn't comment.
- Cross-posts with a canonical link to the blog, one every few days, each marked AI-assisted as [DEV's guidelines](https://dev.to/guidelines-for-ai-assisted-articles-on-dev) require (DEV's AI Disclosure set to "Some AI" and a closing line, added to the earlier ones on 2026-11-19): dev.to has RDAP vs WHOIS, the SSL story, the cron guide and, since 2026-11-19, the domain expiry post, the first with a cover image; "Three Days to MVP" is next; [checkmeup.hashnode.dev](https://checkmeup.hashnode.dev) has the same four, with the same closing line, covers on the cron guide and the domain expiry post, and the publication's description, logos and social image filled in (2026-11-19); Hashnode drops a cover set on a draft when it is published, so a cover goes on after publishing, through Update. No visit has come from either yet. HackerNoon is dropped: it rejected the GitHub Actions guide on 2026-11-18 because a canonical link to checkmeup.net makes a post brand content, published only for a paid brand credit.
- Product Hunt, Reddit (r/webdev, r/selfhosted, r/freelance), Indie Hackers: not tried.

Settled when the channels are chosen and tried.
